package group

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *groupServer) DeleteGroup(ctx context.Context, req *connect.Request[groupsvcv1.DeleteGroupRequest]) (*connect.Response[groupsvcv1.DeleteGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetGroupId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deleteGroup(ctx, s.natsClient, s.dbClient, req.Msg.GetGroupId()); err != nil {
		if eris.Is(err, errDeleteGroup) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBDeleteErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errNoGroupWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the group ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&groupsvcv1.DeleteGroupResponse{}), nil
}

func deleteGroup(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, groupId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)
	group := groupv1.Group{
		Id: groupId,
	}
	if _, err := dbClient.NewDelete().Model(&group).WherePK().Exec(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Debug("group not found", logging.Error(err))
			return errNoGroupWithId
		}
		log.Error("failed deleting group", logging.Error(err))
		return errDeleteGroup
	}

	marshalled, err := proto.Marshal(&groupprocv1.GroupDeleted{
		GroupId: groupId,
	})
	if err != nil {
		log.Error("failed marshalling group deleted event", logging.Error(err))
		return errMarshalGroupDeleted
	}
	if err := nc.Publish(environment.GetGroupDeletedSubject(groupId), marshalled); err != nil {
		log.Error("failed publishing group deleted event", logging.Error(err))
		return errPublishGroupDeleted
	}

	return nil
}
