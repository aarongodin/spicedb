package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"go.opentelemetry.io/otel"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	datastore.Engines = append(datastore.Engines, Engine)
}

const (
	Engine           = "sqlite"
	tableNamespace   = "namespace_config"
	tableTransaction = "relation_tuple_transaction"
	tableTuple       = "relation_tuple"
	tableCaveat      = "caveat"

	colTimestamp        = "timestamp"
	colNamespace        = "namespace"
	colConfig           = "serialized_config"
	colCreatedTxn       = "created_transaction"
	colDeletedTxn       = "deleted_transaction"
	colObjectID         = "object_id"
	colRelation         = "relation"
	colUsersetNamespace = "userset_namespace"
	colUsersetObjectID  = "userset_object_id"
	colUsersetRelation  = "userset_relation"
	colCaveatName       = "name"
	// colCaveatDefinition  = "definition"
	colCaveatContextName = "caveat_name"
	colCaveatContext     = "caveat_context"

	errUnableToInstantiate = "unable to instantiate datastore"

	// This is the largest positive integer possible in postgresql
	// TODO(aarongodin): need to determine what is the largest possible int for sqlite
	liveDeletedTxnID = uint64(9223372036854775807)
)

var (
	tracer  = otel.Tracer("spicedb/internal/datastore/sqlite")
	builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
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

	datastore := &sqliteDatastore{
		store:      db,
		tables:     NewTables(config.tablePrefix),
		ownedStore: true,
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
	datastore := &sqliteDatastore{
		store:        instance,
		tables:       NewTables(tablePrefix),
		ownedStore:   false,
		closeHandler: closeHandler,
	}

	return datastore, nil
}

type sqliteDatastore struct {
	revisions.CommonDecoder

	store  *sql.DB
	tables *Tables
	// ownedStore is true given the DB connection has been created by this package with sql.Open
	ownedStore bool
	// user-supplied callback for closing the datastore, given ownedStore is false
	closeHandler CloseHandler
}

func sqlTxClosure(ctx context.Context, db *sql.DB, txOptions *sql.TxOptions, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		rerr := tx.Rollback()
		if rerr != nil {
			return errors.Join(err, rerr)
		}

		return err
	}

	return tx.Commit()
}

func (ds *sqliteDatastore) SnapshotReader(rev datastore.Revision) datastore.Reader {
	createTxFunc := func(ctx context.Context) (*sql.Tx, txCleanupFunc, error) {
		// TODO(aarongodin): determine if we would like to pass in any options to BeginTx(..)
		tx, err := ds.store.BeginTx(ctx, nil)
		if err != nil {
			return nil, nil, err
		}

		return tx, tx.Rollback, nil
	}

	executor := common.QueryExecutor{
		Executor: newSqliteExecutor(ds.store),
	}

	return &sqliteReader{
		txSource: createTxFunc,
		executor: executor,
		filterer: func(original sq.SelectBuilder) sq.SelectBuilder {
			return original.Where(sq.LtOrEq{colCreatedTxn: rev.(revisions.TransactionIDRevision).TransactionID()}).
				Where(sq.Or{
					sq.Eq{colDeletedTxn: liveDeletedTxnID},
					sq.Gt{colDeletedTxn: rev},
				})
		},
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
	err = sqlTxClosure(ctx, ds.store, &sql.TxOptions{}, func(tx *sql.Tx) error {
		var ierr error
		transactionID, ierr = createTransaction(ctx, tx)
		if ierr != nil {
			return fmt.Errorf("unable to create new txn ID: %w", err)
		}

		txSource := func(context.Context) (*sql.Tx, txCleanupFunc, error) {
			return tx, func() error { return nil }, nil
		}

		executor := common.QueryExecutor{
			Executor: newSqliteExecutor(ds.store),
		}

		rwt := &sqliteReadWriteTransaction{
			sqliteReader: &sqliteReader{
				txSource: txSource,
				executor: executor,
				filterer: func(original sq.SelectBuilder) sq.SelectBuilder {
					return original.Where(sq.Eq{colDeletedTxn: liveDeletedTxnID})
				},
			},
			tx:            tx,
			transactionID: transactionID,
			tables:        ds.tables,
		}

		return fn(ctx, rwt)
	})

	if err != nil {
		// TODO(aarongodin): this should return a nicer error that can be understood by spicedb
		return datastore.NoRevision, err
	}

	return revisions.NewForTransactionID(transactionID), nil
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

// Close closes the data store.
func (ds *sqliteDatastore) Close() error {
	if ds.ownedStore {
		if err := ds.store.Close(); err != nil {
			return fmt.Errorf("unable to close sqlite store: %w", err)
		}
	} else {
		if ds.closeHandler != nil {
			if err := ds.closeHandler(ds.store); err != nil {
				return fmt.Errorf("unable to close sqlite store: %w", err)
			}
		}
	}

	// TODO: reference other implementations to see if there are close steps
	// required for ongoing transactions or other aspects of using the db

	return nil
}
