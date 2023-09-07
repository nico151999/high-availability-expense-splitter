package util_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
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
		groupId := util.GenerateIdWithPrefix("group")
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "groups" (.+) WHERE (.+)"id" = '%s'(.+)`, groupId)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				FromCSVString(groupId))
		model, err := util.CheckResourceExists[*groupv1.Group](context.Background(), bun.NewDB(db, pgdialect.New()), groupId)
		if err != nil {
			t.Error(err)
		}
		if model.Id != groupId {
			t.Errorf("expected ID to be %s but it was %s", groupId, model.Id)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})
}
