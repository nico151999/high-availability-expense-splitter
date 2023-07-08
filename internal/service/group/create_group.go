package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/nats-io/nats.go"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/errors"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var errMarshalGroupCreated = eris.New("failed marshalling group created event")
var errPublishGroupCreated = eris.New("failed publishing group created event")

func (s *groupServer) CreateGroup(ctx context.Context, req *connect.Request[groupsvcv1.CreateGroupRequest]) (*connect.Response[groupsvcv1.CreateGroupResponse], error) {
	ctx = logging.IntoContext(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"groupName",
				req.Msg.GetName())))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	groupId, err := createGroup(ctx, s.natsClient, req.Msg)
	if err != nil {
		if eris.Is(err, errMarshalGroupCreated) || eris.Is(err, errPublishGroupCreated) {
			return nil, errors.NewErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed requesting group creation",
				[]protoreflect.ProtoMessage{
					&errdetails.ErrorInfo{
						Reason: "failed publishing group creation task",
						Domain: environment.GetTaskPublicationErrorReason(ctx),
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

func createGroup(ctx context.Context, nc *nats.Conn, req *groupsvcv1.CreateGroupRequest) (string, error) {
	log := otel.NewOtelLoggerFromContext(ctx)
	// TODO: generate group ID, check if it is not already taken, add "group created" event to NATS and finally return the generated group ID if everything went fine

	groupId := "my-group-id"    // TODO: generate group ID function
	requestorEmail := "ab@c.de" // TODO: take user email from context

	marshalled, err := proto.Marshal(&groupprocv1.GroupCreated{
		GroupId:        groupId,
		Name:           req.GetName(),
		RequestorEmail: requestorEmail,
	})
	if err != nil {
		log.Error("failed marshalling group created event", logging.Error(err))
		return "", errMarshalGroupCreated
	}
	// Simple Publisher
	if err := nc.Publish(environment.GetGroupCreatedSubject(groupId), marshalled); err != nil {
		log.Error("failed publishing group created event", logging.Error(err))
		return "", errPublishGroupCreated
	}
	return groupId, nil
}
