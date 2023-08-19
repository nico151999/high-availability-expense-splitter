package person

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	"github.com/nats-io/nats.go"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
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

func (s *personServer) DeletePerson(ctx context.Context, req *connect.Request[personsvcv1.DeletePersonRequest]) (*connect.Response[personsvcv1.DeletePersonResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"personId",
				req.Msg.GetPersonId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := deletePerson(ctx, s.natsClient, s.dbClient, req.Msg.GetPersonId()); err != nil {
		if eris.Is(err, errDeletePerson) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "deleting person from database failed",
						Domain: environment.GetDBDeleteErrorReason(ctx),
					},
				})
		} else if eris.Is(err, errNoPersonWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the person ID does not exist"))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&personsvcv1.DeletePersonResponse{}), nil
}

func deletePerson(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, personId string) error {
	log := otel.NewOtelLoggerFromContext(ctx)
	person := personv1.Person{
		Id: personId,
	}
	if err := dbClient.NewDelete().Model(&person).WherePK().Returning("group_id").Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Info("person not found", logging.Error(err))
			return errNoPersonWithId
		}
		log.Error("failed deleting person", logging.Error(err))
		return errDeletePerson
	}

	marshalled, err := proto.Marshal(&personprocv1.PersonDeleted{
		PersonId: personId,
	})
	if err != nil {
		log.Error("failed marshalling person deleted event", logging.Error(err))
		return errMarshalPersonDeleted
	}
	if err := nc.Publish(environment.GetPersonDeletedSubject(person.GroupId, personId), marshalled); err != nil {
		log.Error("failed publishing person deleted event", logging.Error(err))
		return errPublishPersonDeleted
	}

	return nil
}
