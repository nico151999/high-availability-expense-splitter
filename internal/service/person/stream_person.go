package person

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/mq/service"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var streamPersonAlive = personsvcv1.StreamPersonResponse{
	Update: &personsvcv1.StreamPersonResponse_StillAlive{},
}

func (s *personServer) StreamPerson(ctx context.Context, req *connect.Request[personsvcv1.StreamPersonRequest], srv *connect.ServerStream[personsvcv1.StreamPersonResponse]) error {
	ctx, cancel := context.WithTimeout(
		logging.IntoContext(
			ctx,
			logging.FromContext(ctx).With(
				logging.String(
					"personId",
					req.Msg.GetId()))),
		time.Hour)
	defer cancel()

	streamSubject := fmt.Sprintf("%s.*", environment.GetPersonSubject("*", req.Msg.GetId()))
	if err := service.StreamResource(ctx, s.natsClient, streamSubject, func(ctx context.Context) (*personsvcv1.StreamPersonResponse, error) {
		return sendCurrentPerson(ctx, s.dbClient, req.Msg.GetId())
	}, srv, &streamPersonAlive); err != nil {
		if eris.Is(err, service.ErrResourceNoLongerFound) {
			return connect.NewError(
				connect.CodeDataLoss,
				eris.New("the person does no longer exist"))
		} else if eris.As(err, &util.ResourceNotFoundError{}) {
			return connect.NewError(
				connect.CodeNotFound,
				eris.New("the person does not exist"))
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

func sendCurrentPerson(ctx context.Context, dbClient bun.IDB, personId string) (*personsvcv1.StreamPersonResponse, error) {
	person, err := util.CheckResourceExists[*personv1.Person](ctx, dbClient, personId)
	if err != nil {
		return nil, err
	}
	return &personsvcv1.StreamPersonResponse{
		Update: &personsvcv1.StreamPersonResponse_Person{
			Person: person,
		},
	}, nil
}
