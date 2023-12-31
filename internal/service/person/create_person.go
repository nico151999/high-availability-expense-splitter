package person

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *personServer) CreatePerson(ctx context.Context, req *connect.Request[personsvcv1.CreatePersonRequest]) (*connect.Response[personsvcv1.CreatePersonResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"personName",
				req.Msg.GetName())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	personId, err := createPerson(ctx, s.natsClient, s.dbClient, req.Msg)
	if err != nil {
		if eris.Is(err, errPublishPersonCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed finalizing person creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetMessagePublicationErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errInsertPerson) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBInsertErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&personsvcv1.CreatePersonResponse{
		Id: personId,
	}), nil
}

func createPerson(ctx context.Context, nc *nats.EncodedConn, db bun.IDB, req *personsvcv1.CreatePersonRequest) (string, error) {
	log := logging.FromContext(ctx)

	personId := util.GenerateIdWithPrefix("person")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := util.CheckResourceExists[*groupv1.Group](ctx, tx, req.GetGroupId()); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(&personv1.Person{
			Id:      personId,
			GroupId: req.GetGroupId(),
			Name:    req.GetName(),
		}).Exec(ctx); err != nil {
			log.Error("failed inserting person", logging.Error(err))
			return errInsertPerson
		}

		if err := nc.Publish(environment.GetPersonCreatedSubject(req.GetGroupId(), personId), &personprocv1.PersonCreated{
			Id:             personId,
			GroupId:        req.GetGroupId(),
			Name:           req.GetName(),
			RequestorEmail: requestorEmail,
		}); err != nil {
			log.Error("failed publishing person created event", logging.Error(err))
			return errPublishPersonCreated
		}
		return nil
	}); err != nil {
		return "", err
	}
	return personId, nil
}
