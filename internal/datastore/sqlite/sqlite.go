package sqlite

import (
	"context"
	"database/sql"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	datastore.Engines = append(datastore.Engines, Engine)
}

const (
	Engine = "sqlite"
	// tableNamespace   = "namespace_config"
	// tableTransaction = "relation_tuple_transaction"
	// tableTuple       = "relation_tuple"
	// tableCaveat      = "caveat"

	// colXID               = "xid"
	// colTimestamp         = "timestamp"
	// colNamespace         = "namespace"
	// colConfig            = "serialized_config"
	// colCreatedXid        = "created_xid"
	// colDeletedXid        = "deleted_xid"
	// colSnapshot          = "snapshot"
	// colObjectID          = "object_id"
	// colRelation          = "relation"
	// colUsersetNamespace  = "userset_namespace"
	// colUsersetObjectID   = "userset_object_id"
	// colUsersetRelation   = "userset_relation"
	// colCaveatName        = "name"
	// colCaveatDefinition  = "definition"
	// colCaveatContextName = "caveat_name"
	// colCaveatContext     = "caveat_context"

	errUnableToInstantiate = "unable to instantiate datastore"

	// // The parameters to this format string are:
	// // 1: the created_xid or deleted_xid column name
	// //
	// // The placeholders are the snapshot and the expected boolean value respectively.
	// snapshotAlive = "pg_visible_in_snapshot(%[1]s, ?) = ?"

	// // This is the largest positive integer possible in postgresql
	// liveDeletedTxnID = uint64(9223372036854775807)

	// tracingDriverName = "postgres-tracing"

	// gcBatchDeleteSize = 1000

	// livingTupleConstraint = "uq_relation_tuple_living_xid"
)

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
) (datastore.Datastore, error) {
	return newSqliteDatastoreWithInstance(ctx, instance)
}

func newSqliteDatastore(
	ctx context.Context,
	url string,
	options ...Option,
) (datastore.Datastore, error) {
	_, err := generateConfig(options)
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
		store: db,
	}

	return datastore, nil
}

func newSqliteDatastoreWithInstance(
	ctx context.Context,
	instance *sql.DB,
) (datastore.Datastore, error) {
	datastore := &sqliteDatastore{
		store: instance,
	}

	return datastore, nil
}

type sqliteDatastore struct {
	store *sql.DB
	// an owned store is one that has been instantiated by this package
	ownedStore bool
}

func (ds *sqliteDatastore) SnapshotReader(revRaw datastore.Revision) datastore.Reader {
	// TODO(aarongodin): determine if we need this for sqlite - are there aspects of a sqlite revision that will be different?
	// this is a cast to a local type, I think you can presume that there is compatible struct fields but other methods on this postgresRevision type
	// that give you access to postgres-specific things
	// rev := revRaw.(postgresRevision)

	return &sqliteReader{
		// queryFuncs,
		// executor,
		// buildLivingObjectFilterForRevision(rev),
	}
}

// ReadWriteTx starts a read/write transaction, which will be committed if no error is
// returned and rolled back if an error is returned.
func (ds *sqliteDatastore) ReadWriteTx(
	ctx context.Context,
	fn datastore.TxUserFunc,
	opts ...options.RWTOptionsOption,
) (datastore.Revision, error) {
	return nil, nil
}

// These ones should go in their own files eventually:

func (ds *sqliteDatastore) Watch(ctx context.Context, afterRevision datastore.Revision, options datastore.WatchOptions) (<-chan *datastore.RevisionChanges, <-chan error) {
	return nil, nil
}

// ReadyState returns a state indicating whether the datastore is ready to accept data.
// Datastores that require database schema creation will return not-ready until the migrations
// have been run to create the necessary tables.
func (ds *sqliteDatastore) ReadyState(ctx context.Context) (datastore.ReadyState, error) {
	return datastore.ReadyState{
		IsReady: true,
		Message: "ready",
	}, nil
}

// Features returns an object representing what features this
// datastore can support.
func (ds *sqliteDatastore) Features(ctx context.Context) (*datastore.Features, error) {
	return nil, nil
}

// Statistics returns relevant values about the data contained in this cluster.
func (ds *sqliteDatastore) Statistics(ctx context.Context) (datastore.Stats, error) {
	return datastore.Stats{}, nil
}

// Close closes the data store.
func (ds *sqliteDatastore) Close() error {
	// idea: check ds.ownedStore here and delegate the close to a user defined function
	return nil
}
