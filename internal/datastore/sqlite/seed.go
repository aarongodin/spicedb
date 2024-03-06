package sqlite

import (
	"context"
	"fmt"

	log "github.com/authzed/spicedb/internal/logging"
	"github.com/authzed/spicedb/pkg/datastore"
)

func (ds *sqliteDatastore) isSeeded(ctx context.Context) (bool, error) {
	headRevision, err := ds.HeadRevision(ctx)
	if err != nil {
		return false, err
	}
	if headRevision == datastore.NoRevision {
		return false, nil
	}

	return true, nil
}

func (ds *sqliteDatastore) seedDatabase(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "seedDatabase")
	defer span.End()

	isSeeded, err := ds.isSeeded(ctx)
	if err != nil {
		return err
	}
	if isSeeded {
		return nil
	}

	insertFirstRevisionQuery := fmt.Sprintf("INSERT INTO %s (id, timestamp) VALUES (1, datetime(1, 'unixepoch')) ON CONFLICT DO NOTHING;", ds.tables.tableTransaction)
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

	return nil
}
