package person

import (
	"context"
	"database/sql"
	"time"

	"connectrpc.com/connect"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *personServer) GetPerson(ctx context.Context, req *connect.Request[personsvcv1.GetPersonRequest]) (*connect.Response[personsvcv1.GetPersonResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"personId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	person, err := getPerson(ctx, s.dbClient, req.Msg.GetId())
	if err != nil {
		if eris.Is(err, errSelectPerson) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBSelectErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
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

	return connect.NewResponse(&personsvcv1.GetPersonResponse{
		Person: person,
	}), nil
}

func getPerson(ctx context.Context, dbClient bun.IDB, personId string) (*personv1.Person, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	person := personv1.Person{
		Id: personId,
	}
	if err := dbClient.NewSelect().Model(&person).WherePK().Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			log.Debug("person not found", logging.Error(err))
			return nil, errNoPersonWithId
		}
		log.Error("failed getting person", logging.Error(err))
		return nil, errSelectPerson
	}

	return &person, nil
}
