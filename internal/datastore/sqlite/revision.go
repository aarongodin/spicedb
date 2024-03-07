package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/authzed/spicedb/internal/datastore/revisions"
	"github.com/authzed/spicedb/pkg/datastore"
)

const (
	errRevision       = "unable to find revision: %w"
	errCheckRevision  = "unable to check revision: %w"
	errRevisionFormat = "invalid revision format: %w"
)

func (ds *sqliteDatastore) OptimizedRevision(ctx context.Context) (datastore.Revision, error) {
	return ds.HeadRevision(ctx)
}

func (ds *sqliteDatastore) HeadRevision(ctx context.Context) (datastore.Revision, error) {
	query, args, err := ds.q.selectLastTransactionID.ToSql()
	if err != nil {
		return datastore.NoRevision, fmt.Errorf(errRevision, err)
	}
	revision, err := ds.loadRevision(ctx, query, args)
	if err != nil {
		return datastore.NoRevision, err
	}
	if revision == 0 {
		return datastore.NoRevision, nil
	}

	return revisions.NewForTransactionID(revision), nil
}

func (ds *sqliteDatastore) loadRevision(ctx context.Context, query string, args []any) (uint64, error) {
	ctx, span := tracer.Start(ctx, "loadRevision")
	defer span.End()

	var revision *uint64
	err := ds.db.QueryRowContext(ctx, query, args...).Scan(&revision)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf(errRevision, err)
	}

	if revision == nil {
		return 0, nil
	}

	return *revision, nil
}

func (ds *sqliteDatastore) CheckRevision(ctx context.Context, rev datastore.Revision) error {
	if rev == datastore.NoRevision {
		return datastore.NewInvalidRevisionErr(rev, datastore.CouldNotDetermineRevision)
	}

	revision, ok := rev.(revisions.TransactionIDRevision)
	if !ok {
		return fmt.Errorf("expected transaction revision, got %T", rev)
	}

	query, args, err := ds.q.selectTransaction.Where(sq.Eq{colID: revision}).ToSql()
	if err != nil {
		return fmt.Errorf(errRevision, err)
	}
	result, err := ds.loadRevision(ctx, query, args)
	if err != nil {
		return fmt.Errorf(errCheckRevision, err)
	}
	if uint64(revision) != result {
		return fmt.Errorf("revision not found: %d", revision)
	}

	return nil
}
