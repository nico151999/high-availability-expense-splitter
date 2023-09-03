package expense_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	expensesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expense/v1"
	expenseTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/expense/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetExpense(t *testing.T) {
	log := logging.GetLogger().Named("testGetExpense")
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

	t.Run("Get Expense successfully", func(t *testing.T) {
		expenseName := "test-expense"
		groupId := "group-543210987654321"
		by := "person-123456789012345"
		tsFormat := "2006-01-02 15:04:05-07"
		timestamp := time.Unix(1693523248, 0)
		currencyId := "currency-135791357913579"
		expenseId := "expense-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "expenses" (.+) WHERE (.+)"id" = '%s'(.+)`, expenseId)).
			WillReturnRows(sqlmock.NewRows([]string{"name", "group_id", "by_id", "timestamp", "currency_id"}).
				FromCSVString(fmt.Sprintf("%s,%s,%s,%s,%s", expenseName, groupId, by, timestamp.Format(tsFormat), currencyId)))
		resp, err := client.GetExpense(ctx, connect.NewRequest(&expensesvcv1.GetExpenseRequest{
			Id: expenseId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.Msg.GetExpense().GetName() != expenseName {
			t.Errorf("expected expense name to be '%s' but it was '%s'", expenseName, resp.Msg.GetExpense().GetName())
		}
		if resp.Msg.GetExpense().GetGroupId() != groupId {
			t.Errorf("expected group ID to be '%s' but it was '%s'", groupId, resp.Msg.GetExpense().GetGroupId())
		}
		if resp.Msg.GetExpense().GetById() != by {
			t.Errorf("expected by to be '%s' but it was '%s'", by, resp.Msg.GetExpense().GetById())
		}
		if resp.Msg.GetExpense().GetTimestamp().AsTime().UTC() != timestamp.UTC() {
			t.Errorf("expected timestamp to be '%s' but it was '%s'", timestamp.UTC().Format(tsFormat), resp.Msg.GetExpense().GetTimestamp().AsTime().UTC().Format(tsFormat))
		}
		if resp.Msg.GetExpense().GetCurrencyId() != currencyId {
			t.Errorf("expected by to be '%s' but it was '%s'", currencyId, resp.Msg.GetExpense().GetCurrencyId())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting Expense due to empty ID", func(t *testing.T) {
		resp, err := client.GetExpense(ctx, connect.NewRequest(&expensesvcv1.GetExpenseRequest{
			Id: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail getting Expense due to non existence", func(t *testing.T) {
		expenseId := "expense-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "expenses" (.+) WHERE (.+)"id" = '%s'(.+)`, expenseId)).WillReturnError(sql.ErrNoRows)
		resp, err := client.GetExpense(ctx, connect.NewRequest(&expensesvcv1.GetExpenseRequest{
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
