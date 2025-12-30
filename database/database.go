package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/nanoteck137/pyrin/ember"
	"github.com/nanoteck137/authlab/database/migrations"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var ErrItemNotFound = errors.New("database: item not found")
var ErrItemAlreadyExists = errors.New("database: item already exists")

var dialect = ember.SqliteDialect()

type DB struct {
	db ember.DB
}

type Tx struct {
	DB

	tx *ember.Tx
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

type Database struct {
	DB

	db *ember.Database
}

func (db *Database) RunMigrateUp() error {
	return migrations.RunMigrateUp(db.db.DB.DB)
}

func (db *Database) RunMigrateDown() error {
	return migrations.RunMigrateDown(db.db.DB.DB)
}

func (db *Database) Begin() (Tx, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return Tx{}, err
	}

	return Tx{
		DB: DB{
			db: tx,
		},
		tx: tx,
	}, nil
}

func Open(dbFile string) (*Database, error) {
	// dbUrl := fmt.Sprintf("file:%s?_foreign_keys=true", dbFile)
	dbUrl := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=ON&_serialized=1&_synchronous=NORMAL", dbFile)
	db, err := ember.OpenDatabase("sqlite3", dbUrl)
	if err != nil {
		return nil, err
	}

	db.ErrorHandler = handleErr

	return &Database{
		db: db,
		DB: DB{
			db: db,
		},
	}, nil
}

func handleErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrItemNotFound
	}

	var e sqlite3.Error
	if errors.As(err, &e) {
		switch e.ExtendedCode {
		case sqlite3.ErrConstraintPrimaryKey:
			return ErrItemAlreadyExists
		}
	}

	return err
}
