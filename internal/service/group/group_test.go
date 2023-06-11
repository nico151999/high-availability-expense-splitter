package group_test

// this package contains helpers for testing the ranproj package

import (
	"context"
	"testing"

	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	grouptesting "github.com/nico151999/high-availability-expense-splitter/internal/service/group/testing"
	"github.com/uptrace/bun"
)

// setupGroupTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client
func setupGroupTest(t *testing.T, db bun.IDB) (groupv1.GroupServiceClient, func(ctx context.Context) error, func(ctx context.Context) error) {
	ln, grpcServer := grouptesting.StartGroupTestServer(t, db)
	cl := grouptesting.SetupGroupTestClient(t, ln)
	return cl.GetGRPCClient(), grpcServer, cl.Close
}
