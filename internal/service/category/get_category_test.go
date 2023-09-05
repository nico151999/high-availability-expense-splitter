package category_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	categorysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/category/v1"
	categoryTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/category/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetCategory(t *testing.T) {
	log := logging.GetLogger().Named("testGetCategory")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, _, closeServer := categoryTesting.SetupCategoryTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing category server: %+v", err)
		}
	}()

	t.Run("Get Category successfully", func(t *testing.T) {
		categoryName := "test-category"
		groupId := "group-543210987654321"
		categoryId := "category-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "categories" (.+) WHERE (.+)"id" = '%s'(.+)`, categoryId)).
			WillReturnRows(sqlmock.NewRows([]string{"name", "group_id"}).
				FromCSVString(fmt.Sprintf("%s,%s", categoryName, groupId)))
		resp, err := client.GetCategory(ctx, connect.NewRequest(&categorysvcv1.GetCategoryRequest{
			Id: categoryId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.Msg.GetCategory().GetName() != categoryName {
			t.Errorf("expected category name to be '%s' but it was '%s'", categoryName, resp.Msg.GetCategory().GetName())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting Category due to empty ID", func(t *testing.T) {
		resp, err := client.GetCategory(ctx, connect.NewRequest(&categorysvcv1.GetCategoryRequest{
			Id: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail getting Category due to non existence", func(t *testing.T) {
		categoryId := "category-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "categories" (.+) WHERE (.+)"id" = '%s'(.+)`, categoryId)).WillReturnError(sql.ErrNoRows)
		resp, err := client.GetCategory(ctx, connect.NewRequest(&categorysvcv1.GetCategoryRequest{
			Id: categoryId,
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
