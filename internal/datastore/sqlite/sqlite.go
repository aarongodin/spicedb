package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
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

// NewSqliteDatastoreWithInstance creates a new datastore using a user-provided DB instance already configured for sqlite.
func NewSqliteDatastoreWithInstance(
	ctx context.Context,
	instance *sql.DB,
	closeHandler CloseHandler,
	tablePrefix string,
) (datastore.Datastore, error) {
	return newSqliteDatastoreWithInstance(ctx, instance, closeHandler, tablePrefix)
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

	// TODO(aarongodin): parsing any additional options and setup of sqlite-specific items
	// goes here
	// TODO(aarongodin) - drive options through this
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, common.RedactAndLogSensitiveConnString(ctx, errUnableToInstantiate, err, url)
	}

	tables := NewTables(config.tablePrefix)
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

func newSqliteDatastoreWithInstance(
	_ context.Context,
	instance *sql.DB,
	closeHandler CloseHandler,
	tablePrefix string,
) (datastore.Datastore, error) {
	tables := NewTables(tablePrefix)
	datastore := &sqliteDatastore{
		db:           instance,
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
	tables       *Tables
	q            *queries
	ownedStore   bool
	closeHandler CloseHandler
}

func (ds *sqliteDatastore) SnapshotReader(rev datastore.Revision) datastore.Reader {
	return &sqliteReader{
		ds.q,
		util.NewTxFactory(ds.db),
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
				ds.q,
				util.NewTxFactoryWithInstance(tx),
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
		// TODO(aarongodin): this should return a nicer error that can be understood by spicedb
		return datastore.NoRevision, err
	}

	return revisions.NewForTransactionID(transactionID), nil
}

func (ds *sqliteDatastore) createTransaction(ctx context.Context, tx *sql.Tx) (uint64, error) {
	query := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", ds.tables.tableTransaction)
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
	// TODO(aarongodin): check the migrations similarly to how other implementations are and report the readystate
	return datastore.ReadyState{IsReady: true}, nil
}

func (ds *sqliteDatastore) Features(ctx context.Context) (*datastore.Features, error) {
	// Explicitly disable the watch feature for now. More research is required to
	// determine if this is either useful or feasible in the sqlite datastore.
	return &datastore.Features{
		Watch: datastore.Feature{Enabled: false},
	}, nil
}

func (ds *sqliteDatastore) Statistics(ctx context.Context) (datastore.Stats, error) {
	return datastore.Stats{}, nil
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
