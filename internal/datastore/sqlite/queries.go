package sqlite

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/internal/datastore/sqlite/migrations"
	"github.com/authzed/spicedb/pkg/datastore"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

const (
	colID               = "id"
	colNamespace        = "namespace"
	colConfig           = "serialized_config"
	colCreatedTxn       = "created_transaction"
	colDeletedTxn       = "deleted_transaction"
	colObjectID         = "object_id"
	colRelation         = "relation"
	colUsersetNamespace = "userset_namespace"
	colUsersetObjectID  = "userset_object_id"
	colUsersetRelation  = "userset_relation"
	colName             = "name"
	colDefinition       = "definition"
	colCaveatName       = "caveat_name"
	colCaveatContext    = "caveat_context"
	colDatabaseIdent    = "database_ident"
)

var (
	builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	schema  = common.NewSchemaInformation(
		colNamespace,
		colObjectID,
		colRelation,
		colUsersetNamespace,
		colUsersetObjectID,
		colUsersetRelation,
		colCaveatName,
		common.ExpandedLogicComparison,
	)
)

// selector is any function that modifies a select
type selector func(sq.SelectBuilder) sq.SelectBuilder

type queries struct {
	// Select
	newTransactionSelector func(datastore.Revision) selector

	selectTuple             sq.SelectBuilder
	selectTupleWithID       sq.SelectBuilder
	selectNamespace         sq.SelectBuilder
	selectLastTransactionID sq.SelectBuilder
	selectTransaction       sq.SelectBuilder
	selectMetadata          sq.SelectBuilder
	selectTupleCount        sq.SelectBuilder
	selectCaveat            sq.SelectBuilder

	// Insert
	insertNamespace sq.InsertBuilder
	insertTuple     sq.InsertBuilder
	insertMetadata  sq.InsertBuilder
	insertCaveat    sq.InsertBuilder

	// Update
	updateNamespace sq.UpdateBuilder
	updateTuple     sq.UpdateBuilder
	updateCaveat    sq.UpdateBuilder

	// Conjunctions
	tupleEquality func(*core.RelationTuple) sq.Eq
}

func newQueries(tables *migrations.Tables) *queries {
	return &queries{
		// Select
		newTransactionSelector: func(rev datastore.Revision) selector {
			// This selector will apply a where clause that ensures the rows are restricted to only the given revision
			transactionID := rev.(revisions.TransactionIDRevision).TransactionID()
			return func(base sq.SelectBuilder) sq.SelectBuilder {
				return base.
					Where(sq.LtOrEq{colCreatedTxn: transactionID}).
					Where(sq.Gt{colDeletedTxn: rev})
			}
		},
		selectTuple: builder.Select(
			colNamespace,
			colObjectID,
			colRelation,
			colUsersetNamespace,
			colUsersetObjectID,
			colUsersetRelation,
			colCaveatName,
			colCaveatContext,
		).From(tables.Tuple()),
		selectTupleWithID: builder.Select(
			colID,
			colNamespace,
			colObjectID,
			colRelation,
			colUsersetNamespace,
			colUsersetObjectID,
			colUsersetRelation,
			colCaveatName,
			colCaveatContext,
		).From(tables.Tuple()),
		selectNamespace:         builder.Select(colConfig, colCreatedTxn).From(tables.Namespace()),
		selectTransaction:       builder.Select(colID).From(tables.Transaction()),
		selectLastTransactionID: builder.Select("MAX(id)").From(tables.Transaction()).Limit(1),
		selectMetadata:          builder.Select(colDatabaseIdent).From(tables.Metadata()),
		selectTupleCount:        builder.Select("COUNT(id)").From(tables.Tuple()),
		selectCaveat:            builder.Select(colDefinition, colCreatedTxn).From(tables.Caveat()).OrderBy(colName),

		// Insert
		insertNamespace: builder.Insert(tables.Namespace()).Columns(
			colNamespace, colConfig, colCreatedTxn,
		),
		insertTuple: builder.Insert(tables.Tuple()).Columns(
			colNamespace,
			colObjectID,
			colRelation,
			colUsersetNamespace,
			colUsersetObjectID,
			colUsersetRelation,
			colCaveatName,
			colCaveatContext,
			colCreatedTxn,
		),
		insertMetadata: builder.Insert(tables.Metadata()).Columns(colDatabaseIdent),
		insertCaveat: builder.Insert(tables.Caveat()).Columns(
			colName,
			colDefinition,
			colCreatedTxn,
		),

		// Update
		updateNamespace: builder.Update(tables.Namespace()),
		updateTuple:     builder.Update(tables.Tuple()),
		updateCaveat:    builder.Update(tables.Caveat()).Where(sq.Eq{colDeletedTxn: liveDeletedTxnID}),

		// Conjunctions
		tupleEquality: func(r *core.RelationTuple) sq.Eq {
			return sq.Eq{
				colNamespace:        r.ResourceAndRelation.Namespace,
				colObjectID:         r.ResourceAndRelation.ObjectId,
				colRelation:         r.ResourceAndRelation.Relation,
				colUsersetNamespace: r.Subject.Namespace,
				colUsersetObjectID:  r.Subject.ObjectId,
				colUsersetRelation:  r.Subject.Relation,
			}
		},
	}
}
