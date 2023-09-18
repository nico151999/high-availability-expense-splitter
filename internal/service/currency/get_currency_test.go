package currency_test // the dedicated _test package prevents import cycles with the testing package

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	currencysvcv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/service/currency/v1"
	currencyTesting "github.com/nico151999/high-availability-expense-splitter/internal/service/currency/testing"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func TestGetCurrency(t *testing.T) {
	log := logging.GetLogger().Named("testGetCurrency")
	ctx := logging.IntoContext(context.Background(), log)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	client, _, closeServer := currencyTesting.SetupCurrencyTest(t, ctx, bun.NewDB(db, pgdialect.New()))
	// we want to close the server only which cascadingly closes the client as well
	defer func() {
		if err := closeServer(); err != nil {
			t.Errorf("failed closing currency server: %+v", err)
		}
	}()

	t.Run("Get Currency successfully", func(t *testing.T) {
		currencyName := "test-currency"
		groupId := "group-543210987654321"
		currencyId := "currency-123456789012345"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "currencies" (.+) WHERE (.+)"id" = '%s'(.+)`, currencyId)).
			WillReturnRows(sqlmock.NewRows([]string{"name", "group_id"}).
				FromCSVString(fmt.Sprintf("%s,%s", currencyName, groupId)))
		resp, err := client.GetCurrency(ctx, connect.NewRequest(&currencysvcv1.GetCurrencyRequest{
			Id: currencyId,
		}))
		if err != nil {
			t.Fatalf("Request failed: %+v", err)
		}
		if resp.Msg.GetCurrency().GetName() != currencyName {
			t.Errorf("expected currency name to be '%s' but it was '%s'", currencyName, resp.Msg.GetCurrency().GetName())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %+v", err)
		}
	})

	t.Run("Fail getting Currency due to empty ID", func(t *testing.T) {
		resp, err := client.GetCurrency(ctx, connect.NewRequest(&currencysvcv1.GetCurrencyRequest{
			Id: "",
		}))
		if err == nil {
			t.Fatalf("Expected request to fail but received a response: %+v", resp)
		}
		t.Logf("Got an error as expected: %+v", err)
	})

	t.Run("Fail getting Currency due to non existence", func(t *testing.T) {
		currencyId := "currency-543210987654321"
		mock.ExpectQuery(fmt.Sprintf(`SELECT (.+) FROM "currencies" (.+) WHERE (.+)"id" = '%s'(.+)`, currencyId)).WillReturnError(sql.ErrNoRows)
		resp, err := client.GetCurrency(ctx, connect.NewRequest(&currencysvcv1.GetCurrencyRequest{
			Id: currencyId,
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
