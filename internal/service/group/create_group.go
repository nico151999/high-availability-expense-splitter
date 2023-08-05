package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/nats-io/nats.go"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (s *groupServer) CreateGroup(ctx context.Context, req *connect.Request[groupsvcv1.CreateGroupRequest]) (*connect.Response[groupsvcv1.CreateGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupName",
				req.Msg.GetName())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	groupId, err := createGroup(ctx, s.natsClient, s.dbClient, req.Msg)
	if err != nil {
		if eris.Is(err, errMarshalGroupCreated) || eris.Is(err, errPublishGroupCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed finalizing group creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "failed publishing group created task",
						Domain: environment.GetMessagePublicationErrorReason(ctx),
					},
				})
		} else if eris.Is(err, errInsertGroup) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed interacting with database",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "inserting group into database failed",
						Domain: environment.GetDBInsertErrorReason(ctx),
					},
				})
		} else {
			return nil, connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
	}

	return connect.NewResponse(&groupsvcv1.CreateGroupResponse{
		GroupId: groupId,
	}), nil
}

func createGroup(ctx context.Context, nc *nats.Conn, db bun.IDB, req *groupsvcv1.CreateGroupRequest) (string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)

	groupId := util.GenerateIdWithPrefix("group")
	requestorEmail := "ab@c.de" // TODO: take user email from context

	if _, err := db.NewInsert().Model(&groupv1.Group{
		Id:   groupId,
		Name: req.GetName(),
	}).Exec(ctx); err != nil {
		log.Error("failed inserting group", logging.Error(err))
		return "", errInsertGroup
	}

	marshalled, err := proto.Marshal(&groupprocv1.GroupCreated{
		GroupId:        groupId,
		Name:           req.GetName(),
		RequestorEmail: requestorEmail,
	})
	if err != nil {
		log.Error("failed marshalling group created event", logging.Error(err))
		return "", errMarshalGroupCreated
	}
	if err := nc.Publish(environment.GetGroupCreatedSubject(groupId), marshalled); err != nil {
		log.Error("failed publishing group created event", logging.Error(err))
		return "", errPublishGroupCreated
	}
	return groupId, nil
}
