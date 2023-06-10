package client

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunotel"
)

func NewPostgresDBClient(user, password, addr, db string) *bun.DB {
	bunDb := bun.NewDB(
		sql.OpenDB(
			pgdriver.NewConnector(
				pgdriver.WithUser(user),
				pgdriver.WithPassword(password),
				pgdriver.WithAddr(addr),
				pgdriver.WithDatabase(db))),
		pgdialect.New())
	bunDb.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName(db)))
	return bunDb
}
