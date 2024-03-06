package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/spicedb/pkg/datastore"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
	"google.golang.org/protobuf/proto"
)

const (
	errUnableToWriteRelationships     = "unable to write relationships: %w"
	errUnableToBulkWriteRelationships = "unable to bulk write relationships: %w"
	errUnableToDeleteRelationships    = "unable to delete relationships: %w"
	errUnableToWriteConfig            = "unable to write namespace config: %w"
	errUnableToDeleteConfig           = "unable to delete namespace config: %w"
)

type sqliteReadWriteTransaction struct {
	*sqliteReader
	q             *queries
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
	deletedNamespaceClause := sq.Or{}
	writeQuery := rwt.q.insertNamespace

	for _, ns := range newConfigs {
		serialized, err := proto.Marshal(ns)
		if err != nil {
			return fmt.Errorf(errUnableToWriteConfig, err)
		}
		deletedNamespaceClause = append(deletedNamespaceClause, sq.Eq{colNamespace: ns.Name})
		writeQuery = writeQuery.Values(ns.Name, serialized, rwt.transactionID)
	}

	delQuery, delArgs, err := rwt.q.updateNamespace.
		Where(sq.Eq{colDeletedTxn: liveDeletedTxnID}).
		Set(colDeletedTxn, rwt.transactionID).
		Where(sq.And{sq.Eq{colDeletedTxn: liveDeletedTxnID}, deletedNamespaceClause}).
		ToSql()
	if err != nil {
		return fmt.Errorf(errUnableToWriteConfig, err)
	}

	_, err = rwt.tx.ExecContext(ctx, delQuery, delArgs...)
	if err != nil {
		return fmt.Errorf(errUnableToWriteConfig, err)
	}

	query, args, err := writeQuery.ToSql()
	if err != nil {
		return fmt.Errorf(errUnableToWriteConfig, err)
	}

	_, err = rwt.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf(errUnableToWriteConfig, err)
	}

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
