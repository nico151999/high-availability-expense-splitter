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

func (s *groupServer) UpdateGroup(ctx context.Context, req *connect.Request[groupsvcv1.UpdateGroupRequest]) (*connect.Response[groupsvcv1.UpdateGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupId",
				req.Msg.GetGroupId())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	group, err := updateGroup(ctx, s.natsClient, s.dbClient, req.Msg.GetGroupId(), req.Msg.GetUpdateFields())
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
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&groupsvcv1.UpdateGroupResponse{
		Group: group,
	}), nil
}

func updateGroup(ctx context.Context, nc *nats.Conn, dbClient bun.IDB, groupId string, params []*groupsvcv1.UpdateGroupRequest_UpdateField) (*groupv1.Group, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
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

		marshalled, err := proto.Marshal(&groupprocv1.GroupUpdated{
			GroupId: groupId,
		})
		if err != nil {
			log.Error("failed marshalling group updated event", logging.Error(err))
			return errMarshalGroupUpdated
		}
		if err := nc.Publish(environment.GetGroupUpdatedSubject(groupId), marshalled); err != nil {
			log.Error("failed publishing group updated event", logging.Error(err))
			return errPublishGroupUpdated
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &group, nil
}
