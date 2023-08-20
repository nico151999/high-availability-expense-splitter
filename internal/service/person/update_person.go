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

func (s *personServer) UpdatePerson(ctx context.Context, req *connect.Request[personsvcv1.UpdatePersonRequest]) (*connect.Response[personsvcv1.UpdatePersonResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"personId",
				req.Msg.GetPersonId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	person, err := updatePerson(ctx, s.natsClient, s.dbClient, req.Msg.GetPersonId(), req.Msg.GetUpdateFields())
	if err != nil {
		if eris.Is(err, errUpdatePerson) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "updating person failed",
						Domain: environment.GetDBUpdateErrorReason(ctx),
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

	return connect.NewResponse(&personsvcv1.UpdatePersonResponse{
		Person: person,
	}), nil
}

func updatePerson(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, personId string, params []*personsvcv1.UpdatePersonRequest_UpdateField) (*personv1.Person, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	person := personv1.Person{
		Id: personId,
	}
	query := dbClient.NewUpdate()
	for _, param := range params {
		switch param.GetUpdateOption().(type) {
		case *personsvcv1.UpdatePersonRequest_UpdateField_Name:
			person.Name = param.GetName()
			query.Column("name")
		}
	}
	if err := query.Model(&person).WherePK().Returning("group_id").Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Info("person not found", logging.Error(err))
			return nil, errNoPersonWithId
		}
		log.Error("failed updating person", logging.Error(err))
		return nil, errUpdatePerson
	}

	marshalled, err := proto.Marshal(&personprocv1.PersonUpdated{
		PersonId: personId,
	})
	if err != nil {
		log.Error("failed marshalling person updated event", logging.Error(err))
		return nil, errMarshalPersonUpdated
	}
	if err := nc.Publish(environment.GetPersonUpdatedSubject(person.GroupId, personId), marshalled); err != nil {
		log.Error("failed publishing person updated event", logging.Error(err))
		return nil, errPublishPersonUpdated
	}

	return &person, nil
}
