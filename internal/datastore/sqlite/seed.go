package sqlite

import (
	"context"
	"fmt"

	log "github.com/authzed/spicedb/internal/logging"
	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/google/uuid"
)

type seedStatus struct {
	hasRevision      bool
	hasDatabaseIdent bool
}

func (s seedStatus) done() bool {
	return s.hasRevision && s.hasDatabaseIdent
}

func (ds *sqliteDatastore) getSeedStatus(ctx context.Context) (seedStatus, error) {
	headRevision, err := ds.HeadRevision(ctx)
	if err != nil {
		return seedStatus{}, err
	}

	m, err := getMetadata(ctx, ds.db, ds.q)
	if err != nil {
		return seedStatus{}, err
	}

	return seedStatus{
		headRevision != datastore.NoRevision,
		m.DatabaseIdent != uuid.Nil,
	}, nil
}

func (ds *sqliteDatastore) seedDatabase(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "seedDatabase")
	defer span.End()

	status, err := ds.getSeedStatus(ctx)
	if err != nil {
		return err
	}
	if status.done() {
		return nil
	}

	insertFirstRevisionQuery := fmt.Sprintf("INSERT INTO %s (id, timestamp) VALUES (1, datetime(1, 'unixepoch')) ON CONFLICT DO NOTHING;", ds.tables.Transaction())
	result, err := ds.db.ExecContext(ctx, insertFirstRevisionQuery)
	if err != nil {
		return fmt.Errorf("seedDatabase: %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("seedDatabase: failed to get last inserted id: %w", err)
	}

	if lastInsertID != 0 {
		log.Ctx(ctx).Info().Int64("headRevision", lastInsertID).Msg("seeded base datastore headRevision")
	}

	if !status.hasDatabaseIdent {
		if err := createMetadata(ctx, ds.db, ds.q); err != nil {
			return fmt.Errorf("seedDatabase: %w", err)
		}
	}

	return nil
}
