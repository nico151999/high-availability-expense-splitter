package expensestake_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	expensestakesvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/expensestake/v1"
	expensestakeTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/expensestake/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestListExpenseStakeIdsInGroup(t *testing.T) {
	log := logging.GetLogger().Named("testListExpenseStakeIdsInGroup")
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

	t.Run("List ExpenseStake Ids in expense successfully", func(t *testing.T) {
		expenseId := "expense-543210987654321"
		expensestakeIds := []string{"expensestake-123456789012345"}
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "expense_stakes" (.+) WHERE (.+)expense_id = '%s'(.+)`, expenseId)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				FromCSVString(strings.Join(expensestakeIds, "\n")))
		resp, err := client.ListExpenseStakeIdsInExpense(ctx, connect.NewRequest(&expensestakesvcv1.ListExpenseStakeIdsInExpenseRequest{
			ExpenseId: expenseId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if len(resp.Msg.GetIds()) != len(expensestakeIds) {
			t.Errorf("expected response to have %d elements but it had %d", len(resp.Msg.GetIds()), len(expensestakeIds))
		}
		for i, id := range resp.Msg.GetIds() {
			if id != expensestakeIds[i] {
				t.Errorf("expected the %dth expense stake ID to be %s but it was %s", i, expensestakeIds[i], id)
			}
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})
}
