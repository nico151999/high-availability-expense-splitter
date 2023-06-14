package group_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	groupsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetGroup(t *testing.T) {
	log := logging.GetLogger().Named("setupGroupTest")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, closeServer, _ := setupGroupTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing group server: %+v", err)
		}
	}()

	t.Run("Get Group successfully", func(t *testing.T) {
		groupName := "test-group"
		mock.ExpectQuery("SELECT (.+) FROM \"groups\" (.+)").WillReturnRows(sqlmock.NewRows([]string{"name"}).FromCSVString(groupName))
		resp, err := client.GetGroup(context.Background(), &groupsvcv1.GetGroupRequest{
			GroupId: "group-123456789a",
		})
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.GetGroup().GetName() != groupName {
			t.Errorf("expected group name to be '%s' but it was '%s'", groupName, resp.GetGroup().GetName())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting Group due to empty ID", func(t *testing.T) {
		resp, err := client.GetGroup(context.Background(), &groupsvcv1.GetGroupRequest{
			GroupId: "",
		})
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})
}
