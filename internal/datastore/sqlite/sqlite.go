package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/internal/datastore/sqlite/migrations"
	"github.com/authzed/spicedb/internal/datastore/sqlite/util"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	datastore.Engines = append(datastore.Engines, Engine)
}

const (
	Engine                 = "sqlite"
	errUnableToInstantiate = "unable to instantiate datastore"
	liveDeletedTxnID       = uint64(9_223_372_036_854_775_807) // largest 64-bit integer
)

var (
	tracer = otel.Tracer("spicedb/internal/datastore/sqlite")
)

type CloseHandler func(*sql.DB) error

// NewSqliteDatastore creates a new datastore from the common datastore options provided through the server/CLI.
func NewSqliteDatastore(
	ctx context.Context,
	url string,
	options ...Option,
) (datastore.Datastore, error) {
	ds, err := newSqliteDatastore(ctx, url, options...)
	if err != nil {
		return nil, err
	}
	// TODO(aarongodin): check into NewSeparatingContextDatastoreProxy(..)
	return ds, nil
}

// NewSqliteDatastoreWithDB creates a new datastore using a user-provided DB instance already configured for sqlite.
func NewSqliteDatastoreWithDB(
	ctx context.Context,
	instance *sql.DB,
	closeHandler CloseHandler,
	tablePrefix string,
) (datastore.Datastore, error) {
	return newSqliteDatastoreWithDB(ctx, instance, closeHandler, tablePrefix)
}

func newSqliteDatastore(
	ctx context.Context,
	url string,
	options ...Option,
) (datastore.Datastore, error) {
	config, err := generateConfig(options)
	if err != nil {
		return nil, common.RedactAndLogSensitiveConnString(ctx, errUnableToInstantiate, err, url)
	}

	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, common.RedactAndLogSensitiveConnString(ctx, errUnableToInstantiate, err, url)
	}

	driver := migrations.NewSQLiteDriverFromDB(db, config.tablePrefix)
	// TODO(aarongodin): this 100 batch size is arbitrary. Maybe it doesn't matter much for sqlite
	if err := migrations.MigrateToVersion(ctx, driver, "head", time.Second, 100); err != nil {
		return nil, err
	}
	tables := migrations.NewTables(config.tablePrefix)
	datastore := &sqliteDatastore{
		db:         db,
		tables:     tables,
		q:          newQueries(tables),
		ownedStore: true,
		CommonDecoder: revisions.CommonDecoder{
			Kind: revisions.TransactionID,
		},
	}

	if err := datastore.seedDatabase(ctx); err != nil {
		return nil, fmt.Errorf("failed seeding sqlite store: %w", err)
	}

	return datastore, nil
}

func newSqliteDatastoreWithDB(
	ctx context.Context,
	db *sql.DB,
	closeHandler CloseHandler,
	tablePrefix string,
) (datastore.Datastore, error) {
	driver := migrations.NewSQLiteDriverFromDB(db, tablePrefix)
	// TODO(aarongodin): this 100 batch size is arbitrary. Maybe it doesn't matter much for sqlite
	if err := migrations.MigrateToVersion(ctx, driver, "head", time.Second, 100); err != nil {
		return nil, err
	}
	tables := migrations.NewTables(tablePrefix)
	datastore := &sqliteDatastore{
		db:           db,
		tables:       tables,
		q:            newQueries(tables),
		ownedStore:   false,
		closeHandler: closeHandler,
		CommonDecoder: revisions.CommonDecoder{
			Kind: revisions.TransactionID,
		},
	}

	return datastore, nil
}

type sqliteDatastore struct {
	revisions.CommonDecoder
	db           *sql.DB
	tables       *migrations.Tables
	q            *queries
	ownedStore   bool
	closeHandler CloseHandler
}

func (ds *sqliteDatastore) SnapshotReader(rev datastore.Revision) datastore.Reader {
	return &sqliteReader{
		ds.db,
		ds.q,
		common.QueryExecutor{
			Executor: newSqliteExecutor(ds.db),
		},
		ds.q.newTransactionSelector(rev),
	}
}

func (ds *sqliteDatastore) ReadWriteTx(
	ctx context.Context,
	fn datastore.TxUserFunc,
	opts ...options.RWTOptionsOption,
) (datastore.Revision, error) {
	// TODO(aarongodin): implement usage of config
	// config := options.NewRWTOptionsWithOptions(opts...)
	// config features to support:
	//   - retries

	var (
		err           error
		transactionID uint64
	)

	// TODO(aarongodin): determine if sqlite has support for TxOptions we would like to include in particular, `Isolation`
	err = util.WithTx(ctx, ds.db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		var ierr error
		transactionID, ierr = ds.createTransaction(ctx, tx)
		if ierr != nil {
			return fmt.Errorf("unable to create new txn ID: %w", err)
		}

		executor := common.QueryExecutor{
			Executor: newSqliteExecutor(ds.db),
		}

		rwt := &sqliteReadWriteTransaction{
			&sqliteReader{
				ds.db,
				ds.q,
				executor,
				func(original sq.SelectBuilder) sq.SelectBuilder {
					return original.Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
				},
			},
			ds.q,
			tx,
			transactionID,
		}

		return fn(ctx, rwt)
	})

	if err != nil {
		return datastore.NoRevision, err
	}

	return revisions.NewForTransactionID(transactionID), nil
}

func (ds *sqliteDatastore) createTransaction(ctx context.Context, tx *sql.Tx) (uint64, error) {
	query := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", ds.tables.Transaction())
	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed creating transaction: %w", err)
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last inserted id: %w", err)
	}
	return uint64(lastInsertID), nil
}

func (ds *sqliteDatastore) Watch(ctx context.Context, afterRevision datastore.Revision, options datastore.WatchOptions) (<-chan *datastore.RevisionChanges, <-chan error) {
	// Return nil channels for now since the watch feature is explicitly disabled.
	// Implement this once the watch feature is determined to be required for sqlite.
	return nil, nil
}

func (ds *sqliteDatastore) ReadyState(ctx context.Context) (datastore.ReadyState, error) {
	seedStatus, err := ds.getSeedStatus(ctx)
	if err != nil {
		return datastore.ReadyState{}, err
	}
	if !seedStatus.done() {
		return datastore.ReadyState{
			Message: "datastore is not properly seeded",
			IsReady: false,
		}, nil
	}

	return datastore.ReadyState{
		Message: "",
		IsReady: true,
	}, nil
}

func (ds *sqliteDatastore) Features(ctx context.Context) (*datastore.Features, error) {
	// Explicitly disable the watch feature for now. More research is required to
	// determine if this is either useful or feasible in the sqlite datastore.
	return &datastore.Features{
		Watch: datastore.Feature{Enabled: false, Reason: "sqlite is a single process datastore"},
	}, nil
}

func (ds *sqliteDatastore) Statistics(ctx context.Context) (datastore.Stats, error) {
	// Other datastores do these steps in a more performant way. Let's assume sqlite is not used for SpiceDB
	// when in performance-critical situations.

	metadata, err := getMetadata(ctx, ds.db, ds.q)
	if err != nil {
		return datastore.Stats{}, fmt.Errorf("sqlite statistics: %w", err)
	}

	var lazyCount uint64
	countQuery, _, err := ds.q.selectTupleCount.ToSql()
	if err != nil {
		return datastore.Stats{}, fmt.Errorf("sqlite statistics: %w", err)
	}
	row := ds.db.QueryRowContext(ctx, countQuery)
	if err = row.Scan(&lazyCount); err != nil {
		return datastore.Stats{}, fmt.Errorf("sqlite statistics: %w", err)
	}

	rev, err := ds.HeadRevision(ctx)
	if err != nil {
		return datastore.Stats{}, fmt.Errorf("sqlite statistics: %w", err)
	}

	reader := ds.SnapshotReader(rev)
	nsDefs, err := reader.ListAllNamespaces(ctx)
	if err != nil {
		return datastore.Stats{}, fmt.Errorf("sqlite statistics: %w", err)
	}

	return datastore.Stats{
		UniqueID:                   metadata.DatabaseIdent.String(),
		EstimatedRelationshipCount: lazyCount,
		ObjectTypeStatistics:       datastore.ComputeObjectTypeStats(nsDefs),
	}, nil
}

func (ds *sqliteDatastore) Close() error {
	if ds.ownedStore {
		if err := ds.db.Close(); err != nil {
			return fmt.Errorf("unable to close sqlite store: %w", err)
		}
	} else {
		if ds.closeHandler != nil {
			if err := ds.closeHandler(ds.db); err != nil {
				return fmt.Errorf("unable to close sqlite store: %w", err)
			}
		}
	}

	return nil
}
