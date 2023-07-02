package group

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/nats-io/nats.go"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/connect/server"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/proto"
)

var errMarshalGroupCreationRequested = eris.New("failed marshalling group creation requested event")
var errPublishGroupCreationRequested = eris.New("failed publishing group creation requested event")

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
		var conError *connect.Error
		if eris.Is(err, errMarshalGroupCreationRequested) || eris.Is(err, errPublishGroupCreationRequested) {
			conError = server.CreateErrorWithDetails(
				ctx,
				connect.CodeInternal,
				"failed requesting group creation",
				environment.GetTaskPublicationErrorReason(ctx))
		} else {
			conError = connect.NewError(connect.CodeInternal, eris.New("an unexpected error occurred"))
		}
		return nil, conError
	}

	return connect.NewResponse(&groupsvcv1.CreateGroupResponse{
		GroupId: groupId,
	}), nil
}

func createGroup(ctx context.Context, nc *nats.Conn, req *groupsvcv1.CreateGroupRequest) (string, error) {
	log := otel.NewOtelLogger(ctx, logging.FromContext(ctx))
	// TODO: generate group ID, check if it is not already taken, add "group creation requested" event to NATS and finally return the generated group ID if adding to queue was successful

	groupId := "my-group-id"    // TODO: generate group ID function
	requestorEmail := "ab@c.de" // TODO: take user email from context

	marshalled, err := proto.Marshal(&groupprocv1.GroupCreationRequested{
		GroupId:        groupId,
		Name:           req.GetName(),
		RequestorEmail: requestorEmail,
	})
	if err != nil {
		log.Error("failed marshalling group creation requested event", logging.Error(err))
		return "", errMarshalGroupCreationRequested
	}
	// Simple Publisher
	if err := nc.Publish(environment.GetGroupCreationRequestedSubject(), marshalled); err != nil {
		log.Error("failed publishing group creation requested event", logging.Error(err))
		return "", errPublishGroupCreationRequested
	}
	return groupId, nil
}
