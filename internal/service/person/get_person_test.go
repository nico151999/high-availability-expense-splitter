package person_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	personsvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/person/v1"
	personTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/person/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetPerson(t *testing.T) {
	log := logging.GetLogger().Named("testGetPerson")
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

	t.Run("Get Person successfully", func(t *testing.T) {
		personName := "test-person"
		groupId := "group-543210987654321"
		personId := "person-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "people" (.+) WHERE (.+)"id" = '%s'(.+)`, personId)).
			WillReturnRows(sqlmock.NewRows([]string{"name", "group_id"}).
				FromCSVString(fmt.Sprintf("%s,%s", personName, groupId)))
		resp, err := client.GetPerson(ctx, connect.NewRequest(&personsvcv1.GetPersonRequest{
			PersonId: personId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.Msg.GetPerson().GetName() != personName {
			t.Errorf("expected person name to be '%s' but it was '%s'", personName, resp.Msg.GetPerson().GetName())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting Person due to empty ID", func(t *testing.T) {
		resp, err := client.GetPerson(ctx, connect.NewRequest(&personsvcv1.GetPersonRequest{
			PersonId: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail getting Person due to non existence", func(t *testing.T) {
		personId := "person-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "people" (.+) WHERE (.+)"id" = '%s'(.+)`, personId)).WillReturnError(sql.ErrNoRows)
		resp, err := client.GetPerson(ctx, connect.NewRequest(&personsvcv1.GetPersonRequest{
			PersonId: personId,
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
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})
}
