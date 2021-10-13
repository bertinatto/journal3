package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migration/*.sql
var migrationFS embed.FS

type Tx struct {
	*sql.Tx
	now time.Time
}

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

	err = db.migrate()
	if err != nil {
		return err
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

func (db *DB) migrate() error {
	_, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	names, err := fs.Glob(migrationFS, "migration/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	if len(names) == 0 {
		return fmt.Errorf("failed to find at least one migration file")
	}

	for _, name := range names {
		err := db.migrateFile(name)
		if err != nil {
			return fmt.Errorf("failed to execute migration %q: %w", name, err)
		}
	}
	return nil
}

func (db *DB) migrateFile(name string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	err = tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n)
	if err != nil {
		return err
	}
	if n != 0 {
		return nil
	}

	// Read and execute migration file.
	buf, err := fs.ReadFile(migrationFS, name)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}
