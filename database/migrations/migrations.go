package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrations embed.FS

func init() {
	// TODO(patrik): Move?
	goose.SetBaseFS(migrations)
	goose.SetDialect("sqlite3")
}

func RunMigrateUp(conn *sql.DB) error {
	return goose.Up(conn, ".")
}

func RunMigrateDown(conn *sql.DB) error {
	return goose.Down(conn, ".")
}
