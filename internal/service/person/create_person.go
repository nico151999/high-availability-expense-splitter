package person

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
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
		if eris.Is(err, errMarshalPersonCreated) || eris.Is(err, errPublishPersonCreated) {
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
		} else if eris.Is(err, errNoGroupWithId) {
			return nil, connect.NewError(connect.CodeNotFound, eris.New("the group ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&personsvcv1.CreatePersonResponse{
		PersonId: personId,
	}), nil
}

func createPerson(ctx context.Context, nc *nats.Conn, db bun.IDB, req *personsvcv1.CreatePersonRequest) (string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	personId := util.GenerateIdWithPrefix("person")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if _, err := db.NewInsert().Model(&personv1.Person{
		Id:      personId,
		GroupId: req.GetGroupId(),
		Name:    req.GetName(),
	}).Exec(ctx); err != nil {
		log.Error("failed inserting person", logging.Error(err))
		// Check for foreign key violation as documented here: https://www.postgresql.org/docs/current/errcodes-appendix.html
		if pgErr := new(pgdriver.Error); eris.As(err, pgErr) && pgErr.Field('C') == "23503" {
			return "", errNoGroupWithId
		}
		return "", errInsertPerson
	}

	marshalled, err := proto.Marshal(&personprocv1.PersonCreated{
		PersonId:       personId,
		GroupId:        req.GetGroupId(),
		Name:           req.GetName(),
		RequestorEmail: requestorEmail,
	})
	if err != nil {
		log.Error("failed marshalling person created event", logging.Error(err))
		return "", errMarshalPersonCreated
	}
	if err := nc.Publish(environment.GetPersonCreatedSubject(req.GetGroupId(), personId), marshalled); err != nil {
		log.Error("failed publishing person created event", logging.Error(err))
		return "", errPublishPersonCreated
	}
	return personId, nil
}
