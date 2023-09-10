package util_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nico151999/high-availability-expense-splitter/internal/db/model"
	"github.com/nico151999/high-availability-expense-splitter/pkg/db/util"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestCheckResourceExistence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	t.Run("Check resource exists successfully", func(t *testing.T) {
		expenseId := util.GenerateIdWithPrefix("expense")
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "expenses" (.+) WHERE (.+)"id" = '%s'(.+)`, expenseId)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				FromCSVString(expenseId))
		model, err := util.CheckResourceExists[*model.Expense](context.Background(), bun.NewDB(db, pgdialect.New()), expenseId)
		if err != nil {
			t.Error(err)
		}
		if model.Id != expenseId {
			t.Errorf("expected ID to be %s but it was %s", expenseId, model.Id)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})
}
