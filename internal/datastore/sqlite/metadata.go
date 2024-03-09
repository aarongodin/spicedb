package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type metadata struct {
	DatabaseIdent uuid.UUID
}

func getMetadata(ctx context.Context, db *sql.DB, q *queries) (metadata, error) {
	var (
		m             = metadata{}
		databaseIdent string
	)

	query, args, err := q.selectMetadata.ToSql()
	if err != nil {
		return m, fmt.Errorf("error preparing sqlite metadata query: %w", err)
	}
	row := db.QueryRowContext(ctx, query, args...)
	if err = row.Scan(&databaseIdent); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m, nil
		}
		return m, fmt.Errorf("error reading sqlite metadata: %w", err)
	}
	if m.DatabaseIdent, err = uuid.Parse(databaseIdent); err != nil {
		return m, fmt.Errorf("error converting database ident to uuid: %w", err)
	}
	return m, nil
}

func createMetadata(ctx context.Context, db *sql.DB, q *queries) error {
	query, args, err := q.insertMetadata.Values(uuid.NewString()).ToSql()
	if err != nil {
		return fmt.Errorf("error preparing create sqlite metadata query: %w", err)
	}
	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error creating sqlite metadata: %w", err)
	}
	return nil
}
