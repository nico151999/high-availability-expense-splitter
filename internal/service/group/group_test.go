package group_test

// this package contains helpers for testing the ranproj package

import (
	"context"
	"testing"

	"github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	grouptesting "github.com/nico151999/high-availability-expense-splitter/internal/service/group/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
)

// setupGroupTest creates gRPC server and client and returns instances of interfaces allowing to close both the server and the client. The passed context has no effect on the server's lifecycle.
func setupGroupTest(t *testing.T, ctx context.Context, db bun.IDB) (groupv1.GroupServiceClient, func() error, func() error) {
	log := logging.FromContext(ctx).Named("setupGroupTest")
	ctx = logging.IntoContext(ctx, log)

	ln, shutdownServer := grouptesting.StartGroupTestServer(t, ctx, db)
	cl := grouptesting.SetupGroupTestClient(t, ln)
	return cl.GetGRPCClient(), shutdownServer, cl.Close
}
