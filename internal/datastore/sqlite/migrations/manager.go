package migrations

import (
	"database/sql"

	"github.com/authzed/spicedb/internal/datastore/sqlite"
	"github.com/authzed/spicedb/pkg/migrate"
)

const migrationNamePattern = `^[_a-zA-Z]*$`

var (
	noNonatomicMigration migrate.MigrationFunc[Wrapper]
	noTxMigration        migrate.TxMigrationFunc[TxWrapper] // nolint: deadcode, unused, varcheck
	Manager              = migrate.NewManager[*SQLiteDriver, Wrapper, TxWrapper]()
)

type Wrapper struct {
	db     *sql.DB
	tables *sqlite.Tables
}

type TxWrapper struct {
	tx     *sql.Tx
	tables *sqlite.Tables
}

func mustRegisterMigration(version, replaces string, up migrate.MigrationFunc[Wrapper], upTx migrate.TxMigrationFunc[TxWrapper]) {
	if err := Manager.Register(version, replaces, up, upTx); err != nil {
		panic("failed to register migration  " + err.Error())
	}
}
