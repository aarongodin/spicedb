package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

type sqliteReader struct {
	db       *sql.DB
	q        *queries
	executor common.QueryExecutor
	filterer func(sq.SelectBuilder) sq.SelectBuilder
}

const (
	errUnableToReadConfig     = "unable to read namespace config: %w"
	errUnableToListNamespaces = "unable to list namespaces: %w"
	errUnableToQueryTuples    = "unable to query tuples: %w"
	errListCaveats            = "unable to list caveats: %w"
	errReadCaveat             = "unable to read caveat: %w"
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
	loaded, version, err := loadNamespace(ctx, r.db, r.filterer(r.q.selectNamespace), nsName)
	switch {
	case errors.As(err, &datastore.ErrNamespaceNotFound{}):
		return nil, datastore.NoRevision, err
	case err == nil:
		return loaded, version, nil
	default:
		return nil, datastore.NoRevision, fmt.Errorf(errUnableToReadConfig, err)
	}
}

func loadNamespace(ctx context.Context, db *sql.DB, baseQuery sq.SelectBuilder, namespace string) (*core.NamespaceDefinition, datastore.Revision, error) {
	ctx, span := tracer.Start(ctx, "loadNamespace")
	defer span.End()

	query, args, err := baseQuery.Where(sq.Eq{colNamespace: namespace}).ToSql()
	if err != nil {
		return nil, datastore.NoRevision, err
	}

	var config []byte
	var txID uint64
	err = db.QueryRowContext(ctx, query, args...).Scan(&config, &txID)
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

	query, args, err := filteredQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf(errUnableToListNamespaces, err)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(errUnableToListNamespaces, err)
	}
	defer common.LogOnError(ctx, rows.Close)

	for rows.Next() {
		var config []byte
		var txID uint64
		if err := rows.Scan(&config, &txID); err != nil {
			return nil, fmt.Errorf(errUnableToListNamespaces, err)
		}

		loaded := &core.NamespaceDefinition{}
		if err := loaded.UnmarshalVT(config); err != nil {
			return nil, fmt.Errorf(errUnableToListNamespaces, err)
		}

		nsDefs = append(nsDefs, datastore.RevisionedNamespace{
			Definition:          loaded,
			LastWrittenRevision: revisions.NewForTransactionID(txID),
		})
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf(errUnableToListNamespaces, rows.Err())
	}

	return nsDefs, nil
}

func (r *sqliteReader) ReadCaveatByName(ctx context.Context, name string) (*core.CaveatDefinition, datastore.Revision, error) {
	caveats, err := r.LookupCaveatsWithNames(ctx, []string{name})
	if err != nil {
		return nil, nil, fmt.Errorf(errReadCaveat, err)
	}
	if len(caveats) != 1 {
		return nil, nil, fmt.Errorf("read caveats: data integrity error - multiple caveats with name %s for a single revision", name)
	}
	return caveats[0].Definition, caveats[0].LastWrittenRevision, nil
}

func (r *sqliteReader) LookupCaveatsWithNames(ctx context.Context, caveatNames []string) ([]datastore.RevisionedCaveat, error) {
	if len(caveatNames) == 0 {
		return nil, nil
	}
	return r.listCaveats(ctx, caveatNames)
}

func (r *sqliteReader) ListAllCaveats(ctx context.Context) ([]datastore.RevisionedCaveat, error) {
	return r.listCaveats(ctx, nil)
}

func (r *sqliteReader) listCaveats(ctx context.Context, caveatNames []string) ([]datastore.RevisionedCaveat, error) {
	caveatsWithNames := r.q.selectCaveat
	if len(caveatNames) > 0 {
		caveatsWithNames = caveatsWithNames.Where(sq.Eq{colName: caveatNames})
	}

	listSQL, listArgs, err := r.filterer(caveatsWithNames).ToSql()
	if err != nil {
		return nil, fmt.Errorf(errListCaveats, err)
	}

	caveats := make([]datastore.RevisionedCaveat, 0)
	rows, err := r.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, fmt.Errorf(errListCaveats, err)
	}
	defer common.LogOnError(ctx, rows.Close)

	for rows.Next() {
		var defBytes []byte
		var txID uint64

		err = rows.Scan(&defBytes, &txID)
		if err != nil {
			return nil, fmt.Errorf(errListCaveats, err)
		}
		c := core.CaveatDefinition{}
		err = c.UnmarshalVT(defBytes)
		if err != nil {
			return nil, fmt.Errorf(errListCaveats, err)
		}
		caveats = append(caveats, datastore.RevisionedCaveat{
			Definition:          &c,
			LastWrittenRevision: revisions.NewForTransactionID(txID),
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf(errListCaveats, err)
	}

	return caveats, nil
}
