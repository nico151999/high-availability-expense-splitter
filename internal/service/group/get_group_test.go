package group_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	groupTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/group/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetGroup(t *testing.T) {
	log := logging.GetLogger().Named("testGetGroup")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, _, closeServer := groupTesting.SetupGroupTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing group server: %+v", err)
		}
	}()

	t.Run("Get Group successfully", func(t *testing.T) {
		groupName := "test-group"
		groupId := "group-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "groups" (.+) WHERE (.+)"id" = '%s'(.+)`, groupId)).WillReturnRows(sqlmock.NewRows([]string{"name"}).FromCSVString(groupName))
		resp, err := client.GetGroup(ctx, connect.NewRequest(&groupsvcv1.GetGroupRequest{
			GroupId: groupId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.Msg.GetGroup().GetName() != groupName {
			t.Errorf("expected group name to be '%s' but it was '%s'", groupName, resp.Msg.GetGroup().GetName())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting Group due to empty ID", func(t *testing.T) {
		resp, err := client.GetGroup(ctx, connect.NewRequest(&groupsvcv1.GetGroupRequest{
			GroupId: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail getting Group due to non existence", func(t *testing.T) {
		groupId := "group-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "groups" (.+) WHERE (.+)"id" = '%s'(.+)`, groupId)).WillReturnError(sql.ErrNoRows)
		resp, err := client.GetGroup(ctx, connect.NewRequest(&groupsvcv1.GetGroupRequest{
			GroupId: groupId,
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		if connectErr := new(connect.Error); eris.As(err, &connectErr) {
			if connectErr.Code() != connect.CodeNotFound {
				t.Fatalf("Expected code: %+v; got: %+v", connect.CodeNotFound, connectErr.Code())
			}
		} else {
			t.Fatalf("Expected connect error, got: %+v", err)
		}
	})
}
