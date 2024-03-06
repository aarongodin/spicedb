package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/authzed/spicedb/internal/datastore/sqlite"
	"github.com/authzed/spicedb/pkg/migrate"

	_ "github.com/mattn/go-sqlite3"
)

const (
	errUnableToInstantiate       = "unable to instantiate SQLiteDriver: %w"
	migrationVersionColumnPrefix = "_meta_version_"
)

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// SQLiteDriver is an implementation of migrate.Driver for SQLite
type SQLiteDriver struct {
	db     *sql.DB
	tables *sqlite.Tables
}

func NewSQLiteDriver(path string, tablePrefix string) (*SQLiteDriver, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf(errUnableToInstantiate, err)
	}
	return NewSQLiteDriverFromDB(db, tablePrefix), nil
}

func NewSQLiteDriverFromDB(db *sql.DB, tablePrefix string) *SQLiteDriver {
	return &SQLiteDriver{db, sqlite.NewTables(tablePrefix)}
}

func (driver *SQLiteDriver) Version(ctx context.Context) (string, error) {
	query, args, err := builder.Select("version").
		From(driver.tables.MigrationVersion()).
		OrderBy("id desc").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("unable to generate query for revision: %w", err)
	}

	row := driver.db.QueryRowContext(ctx, query, args...)
	var version string
	if err := row.Scan(&version); err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return "", nil
		}
		return "", fmt.Errorf("unable to query revision: %w", err)
	}
	return version, nil
}

func (driver *SQLiteDriver) Conn() Wrapper {
	return Wrapper{db: driver.db, tables: driver.tables}
}

func (driver *SQLiteDriver) RunTx(ctx context.Context, f migrate.TxMigrationFunc[TxWrapper]) error {
	return BeginTxFunc(
		ctx,
		driver.db,
		&sql.TxOptions{Isolation: sql.LevelSerializable},
		func(tx *sql.Tx) error {
			return f(ctx, TxWrapper{tx, driver.tables})
		},
	)
}

func BeginTxFunc(ctx context.Context, db *sql.DB, txOptions *sql.TxOptions, f func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		rerr := tx.Rollback()
		if rerr != nil {
			return errors.Join(err, rerr)
		}

		return err
	}

	return tx.Commit()
}

func (driver *SQLiteDriver) WriteVersion(ctx context.Context, txWrapper TxWrapper, version, replaced string) error {
	query, args, err := builder.
		Insert(driver.tables.MigrationVersion()).
		Columns("version").
		Values(version).ToSql()
	if err != nil {
		return fmt.Errorf("unable to write version: %w", err)
	}

	if _, err := txWrapper.tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("unable to write version: %w", err)
	}

	return nil
}

func (driver *SQLiteDriver) Close(_ context.Context) error {
	return driver.db.Close()
}

var _ migrate.Driver[Wrapper, TxWrapper] = &SQLiteDriver{}
