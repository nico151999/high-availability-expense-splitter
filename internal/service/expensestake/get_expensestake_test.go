package expensestake_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
	expensestakeTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/expensestake/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetExpenseStake(t *testing.T) {
	log := logging.GetLogger().Named("testGetExpenseStake")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, _, closeServer := expensestakeTesting.SetupExpenseStakeTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing expensestake server: %+v", err)
		}
	}()

	t.Run("Get ExpenseStake successfully", func(t *testing.T) {
		var mainValue int32 = 23
		expenseId := "expense-543210987654321"
		forId := "person-123456789012345"
		expensestakeId := "expensestake-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "expense_stakes" (.+) WHERE (.+)"id" = '%s'(.+)`, expensestakeId)).
			WillReturnRows(sqlmock.NewRows([]string{"main_value", "expense_id", "for_id"}).
				FromCSVString(fmt.Sprintf("%d,%s,%s", mainValue, expenseId, forId)))
		resp, err := client.GetExpenseStake(ctx, connect.NewRequest(&expensestakesvcv1.GetExpenseStakeRequest{
			Id: expensestakeId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.Msg.GetExpenseStake().GetMainValue() != mainValue {
			t.Errorf("expected expensestake name to be '%d' but it was '%d'", mainValue, resp.Msg.GetExpenseStake().GetMainValue())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting ExpenseStake due to empty ID", func(t *testing.T) {
		resp, err := client.GetExpenseStake(ctx, connect.NewRequest(&expensestakesvcv1.GetExpenseStakeRequest{
			Id: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail getting ExpenseStake due to non existence", func(t *testing.T) {
		expensestakeId := "expensestake-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "expense_stakes" (.+) WHERE (.+)"id" = '%s'(.+)`, expensestakeId)).WillReturnError(sql.ErrNoRows)
		resp, err := client.GetExpenseStake(ctx, connect.NewRequest(&expensestakesvcv1.GetExpenseStakeRequest{
			Id: expensestakeId,
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
