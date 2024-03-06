package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/internal/datastore/sqlite/util"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

type sqliteReader struct {
	q         *queries
	txFactory util.TxFactory
	executor  common.QueryExecutor
	filterer  func(sq.SelectBuilder) sq.SelectBuilder
}

const (
	errUnableToReadConfig     = "unable to read namespace config: %w"
	errUnableToListNamespaces = "unable to list namespaces: %w"
	errUnableToQueryTuples    = "unable to query tuples: %w"
)

func (r *sqliteReader) QueryRelationships(
	ctx context.Context,
	filter datastore.RelationshipsFilter,
	opts ...options.QueryOptionsOption,
) (iter datastore.RelationshipIterator, err error) {
	qBuilder, err := common.NewSchemaQueryFilterer(schema, r.filterer(r.q.selectTuple)).FilterWithRelationshipsFilter(filter)
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
	qBuilder, err := common.NewSchemaQueryFilterer(schema, r.filterer(r.q.selectTuple)).FilterWithSubjectsSelectors(subjectsFilter.AsSelector())
	if err != nil {
		return nil, err
	}

	queryOpts := options.NewReverseQueryOptionsWithOptions(opts...)

	if queryOpts.ResRelation != nil {
		qBuilder = qBuilder.
			FilterToResourceType(queryOpts.ResRelation.Namespace).
			FilterToRelation(queryOpts.ResRelation.Relation)
	}

	return r.executor.ExecuteQuery(
		ctx,
		qBuilder,
		options.WithLimit(queryOpts.LimitForReverse),
		options.WithAfter(queryOpts.AfterForReverse),
		options.WithSort(queryOpts.SortForReverse),
	)
}

func (r *sqliteReader) ReadNamespaceByName(ctx context.Context, nsName string) (*core.NamespaceDefinition, datastore.Revision, error) {
	tx, txCleanup, err := r.txFactory(ctx)
	if err != nil {
		return nil, datastore.NoRevision, fmt.Errorf(errUnableToReadConfig, err)
	}
	defer common.LogOnError(ctx, txCleanup)

	loaded, version, err := loadNamespace(ctx, tx, r.filterer(r.q.selectNamespace), nsName)
	switch {
	case errors.As(err, &datastore.ErrNamespaceNotFound{}):
		return nil, datastore.NoRevision, err
	case err == nil:
		return loaded, version, nil
	default:
		return nil, datastore.NoRevision, fmt.Errorf(errUnableToReadConfig, err)
	}
}

func loadNamespace(ctx context.Context, tx *sql.Tx, baseQuery sq.SelectBuilder, namespace string) (*core.NamespaceDefinition, datastore.Revision, error) {
	ctx, span := tracer.Start(ctx, "loadNamespace")
	defer span.End()

	query, args, err := baseQuery.Where(sq.Eq{colNamespace: namespace}).ToSql()
	if err != nil {
		return nil, datastore.NoRevision, err
	}

	var config []byte
	var txID uint64
	err = tx.QueryRowContext(ctx, query, args...).Scan(&config, &txID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = datastore.NewNamespaceNotFoundErr(namespace)
		}
		return nil, datastore.NoRevision, err
	}

	loaded := &core.NamespaceDefinition{}
	if err := loaded.UnmarshalVT(config); err != nil {
		return nil, datastore.NoRevision, err
	}

	return loaded, revisions.NewForTransactionID(txID), nil
}

func (r *sqliteReader) ListAllNamespaces(ctx context.Context) ([]datastore.RevisionedNamespace, error) {
	return r.loadNamespaces(ctx, r.filterer(r.q.selectNamespace))
}

func (r *sqliteReader) LookupNamespacesWithNames(ctx context.Context, nsNames []string) ([]datastore.RevisionedNamespace, error) {
	if len(nsNames) == 0 {
		return nil, nil
	}
	return r.loadNamespaces(ctx, r.filterer(r.q.selectNamespace.Where(sq.Eq{colNamespace: nsNames})))
}

func (r *sqliteReader) loadNamespaces(ctx context.Context, filteredQuery sq.SelectBuilder) ([]datastore.RevisionedNamespace, error) {
	var (
		nsDefs []datastore.RevisionedNamespace
		err    error
	)

	if err = r.txFactory.WithTx(ctx, func(tx *sql.Tx) error {
		query, args, err := filteredQuery.ToSql()
		if err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer common.LogOnError(ctx, rows.Close)

		for rows.Next() {
			var config []byte
			var txID uint64
			if err := rows.Scan(&config, &txID); err != nil {
				return err
			}

			loaded := &core.NamespaceDefinition{}
			if err := loaded.UnmarshalVT(config); err != nil {
				return fmt.Errorf(errUnableToReadConfig, err)
			}

			nsDefs = append(nsDefs, datastore.RevisionedNamespace{
				Definition:          loaded,
				LastWrittenRevision: revisions.NewForTransactionID(txID),
			})
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf(errUnableToListNamespaces, err)
	}

	return nsDefs, nil
}
