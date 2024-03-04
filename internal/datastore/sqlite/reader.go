package sqlite

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

type txCleanupFunc func() error

type txFactory func(context.Context) (*sql.Tx, txCleanupFunc, error)

type sqliteReader struct {
	// todo: store the revision here, or something?

	txSource txFactory
	executor common.QueryExecutor
	filterer queryFilterer
}

type queryFilterer func(original sq.SelectBuilder) sq.SelectBuilder

var schema = common.NewSchemaInformation(
	colNamespace,
	colObjectID,
	colRelation,
	colUsersetNamespace,
	colUsersetObjectID,
	colUsersetRelation,
	colCaveatName,
	common.ExpandedLogicComparison,
)

var queryTuples = builder.Select(
	colNamespace,
	colObjectID,
	colRelation,
	colUsersetNamespace,
	colUsersetObjectID,
	colUsersetRelation,
	colCaveatContextName,
	colCaveatContext,
).From(tableTuple)

func (r *sqliteReader) QueryRelationships(
	ctx context.Context,
	filter datastore.RelationshipsFilter,
	opts ...options.QueryOptionsOption,
) (iter datastore.RelationshipIterator, err error) {
	qBuilder, err := common.NewSchemaQueryFilterer(schema, r.filterer(queryTuples)).FilterWithRelationshipsFilter(filter)
	if err != nil {
		return nil, err
	}
	return r.executor.ExecuteQuery(ctx, qBuilder, opts...)
}

func (r *sqliteReader) ReverseQueryRelationships(
	ctx context.Context,
	subjectsFilter datastore.SubjectsFilter,
	opts ...options.ReverseQueryOptionsOption,
) (iter datastore.RelationshipIterator, err error) {
	// TOOD(aarongodin): implement
	return nil, nil
}

func (r *sqliteReader) ReadNamespaceByName(ctx context.Context, nsName string) (*core.NamespaceDefinition, datastore.Revision, error) {
	// TOOD(aarongodin): implement
	return nil, nil, nil
}

func (r *sqliteReader) ListAllNamespaces(ctx context.Context) ([]datastore.RevisionedNamespace, error) {
	// TOOD(aarongodin): implement
	return nil, nil
}

func (r *sqliteReader) LookupNamespacesWithNames(ctx context.Context, nsNames []string) ([]datastore.RevisionedNamespace, error) {
	// TOOD(aarongodin): implement
	return nil, nil
}
