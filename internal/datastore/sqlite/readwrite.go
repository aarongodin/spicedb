package sqlite

import (
	"context"
	"database/sql"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/spicedb/pkg/datastore"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

type sqliteReadWriteTransaction struct {
	*sqliteReader

	tx            *sql.Tx
	transactionID uint64
}

func (rwt *sqliteReadWriteTransaction) WriteRelationships(ctx context.Context, mutations []*core.RelationTupleUpdate) error {
	return nil
}

func (rwt *sqliteReadWriteTransaction) DeleteRelationships(ctx context.Context, filter *v1.RelationshipFilter) error {
	return nil
}

func (rwt *sqliteReadWriteTransaction) WriteNamespaces(ctx context.Context, newConfigs ...*core.NamespaceDefinition) error {
	return nil
}

func (rwt *sqliteReadWriteTransaction) DeleteNamespaces(ctx context.Context, nsNames ...string) error {
	return nil
}

func (rwt *sqliteReadWriteTransaction) BulkLoad(ctx context.Context, iter datastore.BulkWriteRelationshipSource) (uint64, error) {
	return 0, nil
}

func (rwt *sqliteReadWriteTransaction) WriteCaveats(ctx context.Context, caveats []*core.CaveatDefinition) error {
	return nil
}

func (rwt *sqliteReadWriteTransaction) DeleteCaveats(ctx context.Context, names []string) error {
	return nil
}

var _ datastore.ReadWriteTransaction = &sqliteReadWriteTransaction{}
