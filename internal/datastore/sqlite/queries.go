package sqlite

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/pkg/datastore"
)

const (
	colNamespace         = "namespace"
	colConfig            = "serialized_config"
	colCreatedTxn        = "created_transaction"
	colDeletedTxn        = "deleted_transaction"
	colObjectID          = "object_id"
	colRelation          = "relation"
	colUsersetNamespace  = "userset_namespace"
	colUsersetObjectID   = "userset_object_id"
	colUsersetRelation   = "userset_relation"
	colCaveatName        = "name"
	colCaveatDefinition  = "definition"
	colCaveatContextName = "caveat_name"
	colCaveatContext     = "caveat_context"
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
	selectNamespace         sq.SelectBuilder
	selectLastTransactionID sq.SelectBuilder

	// Insert
	insertNamespace sq.InsertBuilder

	// Update
	updateNamespace sq.UpdateBuilder
}

func newQueries(tables *Tables) *queries {
	return &queries{
		// Select
		newTransactionSelector: func(rev datastore.Revision) selector {
			// This selector will apply a where clause that ensures the rows are restricted to only the given revision
			return func(base sq.SelectBuilder) sq.SelectBuilder {
				return base.Where(sq.LtOrEq{colCreatedTxn: rev.(revisions.TransactionIDRevision).TransactionID()}).
					Where(sq.Or{
						sq.Eq{colDeletedTxn: liveDeletedTxnID},
						sq.Gt{colDeletedTxn: rev},
					})
			}
		},
		selectTuple: builder.Select(
			colNamespace,
			colObjectID,
			colRelation,
			colUsersetNamespace,
			colUsersetObjectID,
			colUsersetRelation,
			colCaveatContextName,
			colCaveatContext,
		).From(tables.tableTuple),
		selectNamespace:         builder.Select(colConfig, colCreatedTxn).From(tables.tableNamespace),
		selectLastTransactionID: builder.Select("MAX(id)").From(tables.tableTransaction).Limit(1),

		// Insert
		insertNamespace: builder.Insert(tables.tableNamespace).Columns(
			colNamespace, colConfig, colCreatedTxn,
		),

		// Update
		updateNamespace: builder.Update(tables.tableNamespace),
	}
}
