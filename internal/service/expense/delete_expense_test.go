package expense_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	expenseTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/expense/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestDeleteExpense(t *testing.T) {
	log := logging.GetLogger().Named("testDeleteExpense")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, _, closeServer := expenseTesting.SetupExpenseTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing expense server: %+v", err)
		}
	}()

	t.Run("Delete Expense successfully", func(t *testing.T) {
		mock.ExpectBegin()
		groupId := "group-543210987654321"
		expenseId := "expense-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`DELETE FROM "expenses" (.+) WHERE (.+)"id" = '%s'(.+)`, expenseId)).
			WillReturnRows(sqlmock.NewRows([]string{"group_id"}).
				FromCSVString(groupId))
		mock.ExpectCommit()
		_, err := client.DeleteExpense(ctx, connect.NewRequest(&expensesvcv1.DeleteExpenseRequest{
			Id: expenseId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail deleting Expense due to empty ID", func(t *testing.T) {
		resp, err := client.DeleteExpense(ctx, connect.NewRequest(&expensesvcv1.DeleteExpenseRequest{
			Id: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail deleting Expense due to non existence", func(t *testing.T) {
		mock.ExpectBegin()
		expenseId := "expense-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`DELETE FROM "expenses" (.+) WHERE (.+)"id" = '%s'(.+)`, expenseId)).WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()
		resp, err := client.DeleteExpense(ctx, connect.NewRequest(&expensesvcv1.DeleteExpenseRequest{
			Id: expenseId,
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
