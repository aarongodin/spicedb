package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/spicedb/internal/datastore/common"
	"github.com/authzed/spicedb/pkg/datastore"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
	"github.com/authzed/spicedb/pkg/spiceerrors"
	"github.com/authzed/spicedb/pkg/tuple"
	"github.com/jzelinskie/stringz"
	"google.golang.org/protobuf/proto"
)

const (
	errUnableToWriteRelationships     = "unable to write relationships: %w"
	errUnableToBulkWriteRelationships = "unable to bulk write relationships: %w"
	errUnableToDeleteRelationships    = "unable to delete relationships: %w"
	errUnableToWriteConfig            = "unable to write namespace config: %w"
	errUnableToDeleteConfig           = "unable to delete namespace config: %w"
	errWriteCaveats                   = "unable to write caveats: %w"
)

type sqliteReadWriteTransaction struct {
	*sqliteReader
	q             *queries
	tx            *sql.Tx
	transactionID uint64
}

func (rwt *sqliteReadWriteTransaction) WriteRelationships(ctx context.Context, mutations []*core.RelationTupleUpdate) error {
	bulkWrite := rwt.q.insertTuple
	bulkWriteHasValues := false
	selectForUpdateQuery := rwt.q.selectTupleWithID
	clauses := sq.Or{}
	createAndTouchMutationsByTuple := make(map[string]*core.RelationTupleUpdate, len(mutations))

	for _, mut := range mutations {
		tpl := mut.Tuple
		tplString := tuple.StringWithoutCaveat(tpl)

		switch mut.Operation {
		case core.RelationTupleUpdate_CREATE:
			createAndTouchMutationsByTuple[tplString] = mut

		case core.RelationTupleUpdate_TOUCH:
			createAndTouchMutationsByTuple[tplString] = mut
			clauses = append(clauses, rwt.q.tupleEquality(tpl))

		case core.RelationTupleUpdate_DELETE:
			clauses = append(clauses, rwt.q.tupleEquality(tpl))

		default:
			return spiceerrors.MustBugf("unknown mutation operation")
		}
	}

	if len(clauses) > 0 {
		query, args, err := selectForUpdateQuery.
			Where(clauses).
			Where(sq.GtOrEq{colDeletedTxn: rwt.transactionID}).
			ToSql()
		if err != nil {
			return fmt.Errorf(errUnableToWriteRelationships, err)
		}

		rows, err := rwt.tx.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf(errUnableToWriteRelationships, err)
		}
		defer common.LogOnError(ctx, rows.Close)

		foundTpl := &core.RelationTuple{
			ResourceAndRelation: &core.ObjectAndRelation{},
			Subject:             &core.ObjectAndRelation{},
		}

		var caveatName string
		var caveatContext caveatContext

		tupleIdsToDelete := make([]int64, 0, len(clauses))
		for rows.Next() {
			var tupleID int64
			if err := rows.Scan(
				&tupleID,
				&foundTpl.ResourceAndRelation.Namespace,
				&foundTpl.ResourceAndRelation.ObjectId,
				&foundTpl.ResourceAndRelation.Relation,
				&foundTpl.Subject.Namespace,
				&foundTpl.Subject.ObjectId,
				&foundTpl.Subject.Relation,
				&caveatName,
				&caveatContext,
			); err != nil {
				return fmt.Errorf(errUnableToWriteRelationships, err)
			}

			// if the relationship to be deleted is for a TOUCH operation and the caveat
			// name or context has not changed, then remove it from delete and create.
			tplString := tuple.StringWithoutCaveat(foundTpl)
			if mut, ok := createAndTouchMutationsByTuple[tplString]; ok {
				foundTpl.Caveat, err = common.ContextualizedCaveatFrom(caveatName, caveatContext)
				if err != nil {
					return fmt.Errorf(errUnableToQueryTuples, err)
				}

				// Ensure the tuples are the same.
				if tuple.Equal(mut.Tuple, foundTpl) {
					delete(createAndTouchMutationsByTuple, tplString)
					continue
				}
			}

			tupleIdsToDelete = append(tupleIdsToDelete, tupleID)
		}

		if rows.Err() != nil {
			return fmt.Errorf(errUnableToWriteRelationships, rows.Err())
		}

		if len(tupleIdsToDelete) > 0 {
			query, args, err := rwt.
				q.updateTuple.
				Where(sq.Eq{colID: tupleIdsToDelete}).
				Set(colDeletedTxn, rwt.transactionID).
				ToSql()
			if err != nil {
				return fmt.Errorf(errUnableToWriteRelationships, err)
			}
			if _, err := rwt.tx.ExecContext(ctx, query, args...); err != nil {
				return fmt.Errorf(errUnableToWriteRelationships, err)
			}
		}
	}

	for _, mut := range createAndTouchMutationsByTuple {
		tpl := mut.Tuple

		var caveatName string
		var caveatContext caveatContext
		if tpl.Caveat != nil {
			caveatName = tpl.Caveat.CaveatName
			caveatContext = tpl.Caveat.Context.AsMap()
		}
		bulkWrite = bulkWrite.Values(
			tpl.ResourceAndRelation.Namespace,
			tpl.ResourceAndRelation.ObjectId,
			tpl.ResourceAndRelation.Relation,
			tpl.Subject.Namespace,
			tpl.Subject.ObjectId,
			tpl.Subject.Relation,
			caveatName,
			&caveatContext,
			rwt.transactionID,
		)
		bulkWriteHasValues = true
	}

	if bulkWriteHasValues {
		query, args, err := bulkWrite.ToSql()
		if err != nil {
			return fmt.Errorf(errUnableToWriteRelationships, err)
		}

		_, err = rwt.tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf(errUnableToWriteRelationships, err)
		}
	}

	return nil
}

func (rwt *sqliteReadWriteTransaction) DeleteRelationships(ctx context.Context, filter *v1.RelationshipFilter) error {
	// Add clauses for the ResourceFilter
	query := rwt.q.updateTuple.Where(sq.Eq{colNamespace: filter.ResourceType, colDeletedTxn: liveDeletedTxnID})
	if filter.OptionalResourceId != "" {
		query = query.Where(sq.Eq{colObjectID: filter.OptionalResourceId})
	}
	if filter.OptionalRelation != "" {
		query = query.Where(sq.Eq{colRelation: filter.OptionalRelation})
	}

	// Add clauses for the SubjectFilter
	if subjectFilter := filter.OptionalSubjectFilter; subjectFilter != nil {
		query = query.Where(sq.Eq{colUsersetNamespace: subjectFilter.SubjectType})
		if subjectFilter.OptionalSubjectId != "" {
			query = query.Where(sq.Eq{colUsersetObjectID: subjectFilter.OptionalSubjectId})
		}
		if relationFilter := subjectFilter.OptionalRelation; relationFilter != nil {
			query = query.Where(sq.Eq{colUsersetRelation: stringz.DefaultEmpty(relationFilter.Relation, datastore.Ellipsis)})
		}
	}

	query = query.Set(colDeletedTxn, rwt.transactionID)

	querySQL, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf(errUnableToDeleteRelationships, err)
	}

	if _, err := rwt.tx.ExecContext(ctx, querySQL, args...); err != nil {
		return fmt.Errorf(errUnableToDeleteRelationships, err)
	}

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
	if len(caveats) == 0 {
		return nil
	}

	insertQuery := rwt.q.insertCaveat
	caveatNames := make([]string, 0, len(caveats))
	for _, newCaveat := range caveats {
		serialized, err := newCaveat.MarshalVT()
		if err != nil {
			return fmt.Errorf("unable to serialize caveat: %w", err)
		}

		insertQuery = insertQuery.Values(newCaveat.Name, serialized, rwt.transactionID)
		caveatNames = append(caveatNames, newCaveat.Name)
	}

	if err := rwt.DeleteCaveats(ctx, caveatNames); err != nil {
		return err
	}

	query, args, err := insertQuery.ToSql()
	if err != nil {
		return fmt.Errorf(errWriteCaveats, err)
	}

	_, err = rwt.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf(errWriteCaveats, err)
	}

	return nil
}

func (rwt *sqliteReadWriteTransaction) DeleteCaveats(ctx context.Context, names []string) error {
	delSQL, delArgs, err := rwt.q.updateCaveat.
		Set(colDeletedTxn, rwt.transactionID).
		Where(sq.Eq{colName: names}).
		ToSql()
	if err != nil {
		return fmt.Errorf("deleting caveats: %w", err)
	}

	_, err = rwt.tx.ExecContext(ctx, delSQL, delArgs...)
	if err != nil {
		return fmt.Errorf("deleting caveats: %w", err)
	}
	return nil
}

var _ datastore.ReadWriteTransaction = &sqliteReadWriteTransaction{}
