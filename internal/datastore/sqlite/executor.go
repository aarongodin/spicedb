package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/authzed/spicedb/internal/datastore/common"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type querier interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

func newSqliteExecutor(tx querier) common.ExecuteQueryFunc {
	// TODO(aarongodin): there is a note in the mysql implementation that this is
	// intentionally not run in a transaction - do we want the same logic here?
	return func(ctx context.Context, sqlQuery string, args []interface{}) ([]*core.RelationTuple, error) {
		span := trace.SpanFromContext(ctx)

		rows, err := tx.QueryContext(ctx, sqlQuery, args...)
		if err != nil {
			return nil, fmt.Errorf(errUnableToQueryTuples, err)
		}
		defer common.LogOnError(ctx, rows.Close)

		span.AddEvent("Query issued to database")

		var tuples []*core.RelationTuple
		for rows.Next() {
			nextTuple := &core.RelationTuple{
				ResourceAndRelation: &core.ObjectAndRelation{},
				Subject:             &core.ObjectAndRelation{},
			}

			var caveatName string
			// TODO(aarongodin): the mysql datastore implements this is a map[string]any
			// we probably want to have something similar to store misc caveats on a tuple, but caveats have yet to be handled
			// var caveatContext caveatContextWrapper
			err := rows.Scan(
				&nextTuple.ResourceAndRelation.Namespace,
				&nextTuple.ResourceAndRelation.ObjectId,
				&nextTuple.ResourceAndRelation.Relation,
				&nextTuple.Subject.Namespace,
				&nextTuple.Subject.ObjectId,
				&nextTuple.Subject.Relation,
				&caveatName,
				// &caveatContext,
			)
			if err != nil {
				return nil, fmt.Errorf(errUnableToQueryTuples, err)
			}

			// TODO(aarongodin): assign the .Caveat
			// nextTuple.Caveat, err = common.ContextualizedCaveatFrom(caveatName, caveatContext)
			// if err != nil {
			// 	return nil, fmt.Errorf(errUnableToQueryTuples, err)
			// }

			tuples = append(tuples, nextTuple)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf(errUnableToQueryTuples, err)
		}
		span.AddEvent("Tuples loaded", trace.WithAttributes(attribute.Int("tupleCount", len(tuples))))
		return tuples, nil
	}
}
