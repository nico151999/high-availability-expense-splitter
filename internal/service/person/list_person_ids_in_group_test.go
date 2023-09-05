package person_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	personTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/person/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestListPersonIdsInGroup(t *testing.T) {
	log := logging.GetLogger().Named("testListPersonIdsInGroup")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, _, closeServer := personTesting.SetupPersonTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing person server: %+v", err)
		}
	}()

	t.Run("List Person Ids in group successfully", func(t *testing.T) {
		groupId := "group-543210987654321"
		personIds := []string{"person-123456789012345"}
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "people" (.+) WHERE (.+)group_id = '%s'(.+)`, groupId)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				FromCSVString(strings.Join(personIds, "\n")))
		resp, err := client.ListPersonIdsInGroup(ctx, connect.NewRequest(&personsvcv1.ListPersonIdsInGroupRequest{
			GroupId: groupId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if len(resp.Msg.GetIds()) != len(personIds) {
			t.Errorf("expected response to have %d elements but it had %d", len(resp.Msg.GetIds()), len(personIds))
		}
		for i, id := range resp.Msg.GetIds() {
			if id != personIds[i] {
				t.Errorf("expected the %dth person ID to be %s but it was %s", i, personIds[i], id)
			}
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})
}
