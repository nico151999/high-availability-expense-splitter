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
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *groupServer) UpdateGroup(ctx context.Context, req *connect.Request[groupsvcv1.UpdateGroupRequest]) (*connect.Response[groupsvcv1.UpdateGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	group, err := updateGroup(ctx, s.natsClient, s.dbClient, req.Msg.GetId(), req.Msg.GetUpdateFields())
	if err != nil {
		if eris.Is(err, errUpdateGroup) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: environment.GetDBUpdateErrorReason(ctx),
						Domain: environment.GetGlobalDomain(ctx),
					},
				})
		} else if eris.Is(err, errNoGroupWithId) {
			return nil, connect.NewError(
				connect.CodeNotFound,
				eris.New("the group ID does not exist"))
		} else if resErr := new(util.ResourceNotFoundError); eris.As(err, resErr) {
			return nil, connect.NewError(connect.CodeNotFound, eris.Errorf("the %s with ID %s does not exist", resErr.ResourceName, resErr.ResourceId))
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&groupsvcv1.UpdateGroupResponse{
		Group: group,
	}), nil
}

func updateGroup(ctx context.Context, nc *nats.EncodedConn, dbClient bun.IDB, groupId string, params []*groupsvcv1.UpdateGroupRequest_UpdateField) (*groupv1.Group, error) {
	log := logging.FromContext(ctx)
	group := groupv1.Group{
		Id: groupId,
	}

	if err := dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		query := tx.NewUpdate()
		for _, param := range params {
			switch param.GetUpdateOption().(type) {
			case *groupsvcv1.UpdateGroupRequest_UpdateField_Name:
				group.Name = param.GetName()
				query.Column("name")
			case *groupsvcv1.UpdateGroupRequest_UpdateField_CurrencyId:
				// TODO: check if currency exists
				group.CurrencyId = param.GetCurrencyId()
				query.Column("currency_id")
			}
		}
		if _, err := query.Model(&group).WherePK().Exec(ctx); err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				log.Info("group not found", logging.Error(err))
				return errNoGroupWithId
			}
			log.Error("failed updating group", logging.Error(err))
			return errUpdateGroup
		}

		if err := nc.Publish(environment.GetGroupUpdatedSubject(groupId), &groupprocv1.GroupUpdated{
			Id: groupId,
		}); err != nil {
			log.Error("failed publishing group updated event", logging.Error(err))
			return errPublishGroupUpdated
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &group, nil
}
