package person

import (
	"context"
	"time"

	"connectrpc.com/connect"
	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *personServer) ListPersonIdsInGroup(ctx context.Context, req *connect.Request[personsvcv1.ListPersonIdsInGroupRequest]) (*connect.Response[personsvcv1.ListPersonIdsInGroupResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	personIds, err := listPersonIds(ctx, s.dbClient, req.Msg.GetGroupId())
	if err != nil {
		if eris.Is(err, errSelectPersonIds) {
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
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&personsvcv1.ListPersonIdsInGroupResponse{
		Ids: personIds,
	}), nil
}

func listPersonIds(ctx context.Context, dbClient bun.IDB, groupId string) ([]string, error) {
	log := logging.FromContext(ctx)
	var personIds []string
	if err := dbClient.NewSelect().Model((*personv1.Person)(nil)).Where("group_id = ?", groupId).Column("id").Order("name ASC").Scan(ctx, &personIds); err != nil {
		log.Error("failed getting person IDs", logging.Error(err))
		// TODO: determine reason why person ID couldn't be fetched and return error-specific ErrVariable; e.g. use unit testing with dummy return values to determine potential return values unless there is something in the bun documentation
		return nil, errSelectPersonIds
	}

	return personIds, nil
}
