package sqlite

import (
	"context"

	"github.com/authzed/spicedb/pkg/datastore"
	"github.com/authzed/spicedb/pkg/datastore/options"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

// reader.go must provide an implementation for Datastore.SnapshotReader(..)

type sqliteReader struct {
	// todo: store the revision here, or something?
}

func (r *sqliteReader) QueryRelationships(
	ctx context.Context,
	filter datastore.RelationshipsFilter,
	opts ...options.QueryOptionsOption,
) (iter datastore.RelationshipIterator, err error) {
	// TOOD(aarongodin): implement
	return nil, nil
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
