package person

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamPersonIdsAlive = personsvcv1.StreamPersonIdsInGroupResponse{
	Update: &personsvcv1.StreamPersonIdsInGroupResponse_StillAlive{},
}

func (s *personServer) StreamPersonIdsInGroup(ctx context.Context, req *connect.Request[personsvcv1.StreamPersonIdsInGroupRequest], srv *connect.ServerStream[personsvcv1.StreamPersonIdsInGroupResponse]) error {
	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	if err := service.StreamResource(ctx, s.natsClient, fmt.Sprintf("%s.*", environment.GetPersonSubject(req.Msg.GetGroupId(), "*")), func(ctx context.Context) (*personsvcv1.StreamPersonIdsInGroupResponse, error) {
		return sendCurrentPersonIds(ctx, s.dbClient, req.Msg.GetGroupId())
	}, srv, &streamPersonIdsAlive); err != nil {
		if eris.Is(err, errSelectPersonIds) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSubscribeResource) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed subscribing to updates",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessageSubscriptionErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendCurrentResourceMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed returning current resource",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendCurrentResourceErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, service.ErrSendStreamAliveMessage) {
			return errors.NewErrorWithDetails(
				ctx,
				connect.CodeCanceled,
				"failed sending alive message to client",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetSendStreamAliveErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else {
			return connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return nil
}

func sendCurrentPersonIds(ctx context.Context, dbClient bun.IDB, groupId string) (*personsvcv1.StreamPersonIdsInGroupResponse, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	var personIds []string
	if err := dbClient.NewSelect().Model((*personv1.Person)(nil)).Where("group_id = ?", groupId).Column("id").Order("name ASC").Scan(ctx, &personIds); err != nil {
		log.Error("failed getting person IDs", logging.Error(err))
		// TODO: determine reason why person IDs couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectPersonIds
	}
	return &personsvcv1.StreamPersonIdsInGroupResponse{
		Update: &personsvcv1.StreamPersonIdsInGroupResponse_Ids{
			Ids: &personsvcv1.StreamPersonIdsInGroupResponse_PersonIds{
				Ids: personIds,
			},
		},
	}, nil
}
