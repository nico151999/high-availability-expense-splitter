package group

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/nats-io/nats.go"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *groupServer) CreateGroup(ctx context.Context, req *connect.Request[groupsvcv1.CreateGroupRequest]) (*connect.Response[groupsvcv1.CreateGroupResponse], error) {
	// TODO: tracing
	log := logging.FromContext(ctx)

	groupId, err := createGroup(ctx, s.natsClient, req.Msg)
	if err != nil {
		statusCode := codes.Internal
		if s, ok := status.FromError(eris.Cause(err)); ok {
			statusCode = s.Code()
		}
		log.Error("failed connecting to NATS", logging.Error(err))
		st, err := status.New(statusCode, "failed").WithDetails(&errdetails.ErrorInfo{
			Reason: environment.GetDBSelectErrorReason(ctx),
			Domain: environment.GetGlobalDomain(ctx),
		})
		if err != nil {
			log.Panic("unexpected error attaching metadata", logging.Error(err))
		}
		return nil, st.Err()
	}

	return connect.NewResponse(&groupsvcv1.CreateGroupResponse{
		GroupId: groupId,
	}), nil
}

func createGroup(ctx context.Context, nc *nats.Conn, req *groupsvcv1.CreateGroupRequest) (string, error) {
	// TODO: generate group ID, check if it is not already taken, add "group creation requested" event to NATS and finally return the generated group ID if adding to queue was successful

	groupId := "my-group-id"    // TODO: generate group ID function
	requestorEmail := "ab@c.de" // TODO: take user email from context

	marshalled, err := proto.Marshal(&groupprocv1.GroupCreationRequested{
		GroupId:        groupId,
		Name:           req.GetName(),
		RequestorEmail: requestorEmail,
	})
	if err != nil {
		return "", eris.Wrap(err, "failed marshalling group creation requested event")
	}
	// Simple Publisher
	if err := nc.Publish(environment.GroupCreationRequested, marshalled); err != nil {
		return "", eris.Wrap(err, "failed marshalling group creation requested event")
	}
	return groupId, nil
}
