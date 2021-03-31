package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db     *sql.DB
	ctx    context.Context
	cancel func()

	DSN string
}

func NewDB(dsn string) *DB {
	db := &DB{
		db:  nil,
		DSN: dsn,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

func (db *DB) Open() error {
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	conn, err := sql.Open("sqlite3", db.DSN)
	if err != nil {
		return err
	}
	db.db = conn

	_, err = db.db.Exec(`PRAGMA journal_mode = wal;`)
	if err != nil {
		return fmt.Errorf("could not enable WAL: %w", err)
	}

	return nil
}

func (db *DB) Close() error {
	db.cancel()
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:  tx,
		now: time.Now().UTC().Truncate(time.Second),
	}, nil
}

type Tx struct {
	*sql.Tx
	now time.Time
}
