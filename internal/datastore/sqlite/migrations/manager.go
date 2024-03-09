package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	log "github.com/authzed/spicedb/internal/logging"
	"github.com/authzed/spicedb/pkg/migrate"
)

var (
	noNonatomicMigration migrate.MigrationFunc[Wrapper]
	noTxMigration        migrate.TxMigrationFunc[TxWrapper] // nolint: deadcode, unused, varcheck
	Manager              = migrate.NewManager[*SQLiteDriver, Wrapper, TxWrapper]()
)

type Wrapper struct {
	db     *sql.DB
	tables *Tables
}

type TxWrapper struct {
	tx     *sql.Tx
	tables *Tables
}

func mustRegisterMigration(version, replaces string, up migrate.MigrationFunc[Wrapper], upTx migrate.TxMigrationFunc[TxWrapper]) {
	if err := Manager.Register(version, replaces, up, upTx); err != nil {
		panic("failed to register migration  " + err.Error())
	}
}

func MigrateToVersion(
	ctx context.Context,
	driver *SQLiteDriver,
	targetRevision string,
	timeout time.Duration,
	backfillBatchSize uint64,
) error {
	// Typically a datastore does not provide a way to run the migrations, but this gives
	// more convenience to those using SQLite for testing so that the migrations can run
	// before the datastore is used.
	log.Ctx(ctx).Info().Str("targetRevision", targetRevision).Msg("running migrations")
	ctxWithBatch := context.WithValue(ctx, migrate.BackfillBatchSize, backfillBatchSize)
	ctx, cancel := context.WithTimeout(ctxWithBatch, timeout)
	defer cancel()
	if err := Manager.Run(ctx, driver, targetRevision, migrate.LiveRun); err != nil {
		return fmt.Errorf("unable to migrate to `%s` revision: %w", targetRevision, err)
	}
	return nil
}
