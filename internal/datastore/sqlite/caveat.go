package sqlite

import (
	"context"

	"github.com/authzed/spicedb/pkg/datastore"
	core "github.com/authzed/spicedb/pkg/proto/core/v1"
)

const (
	errWriteCaveats  = "unable to write caveats: %w"
	errDeleteCaveats = "unable delete caveats: %w"
	errListCaveats   = "unable to list caveats: %w"
	errReadCaveat    = "unable to read caveat: %w"
)

type caveatContext map[string]any

func (r *sqliteReader) ReadCaveatByName(ctx context.Context, name string) (*core.CaveatDefinition, datastore.Revision, error) {
	// TODO(aarongodin): implement
	return nil, nil, nil
}

func (r *sqliteReader) LookupCaveatsWithNames(ctx context.Context, caveatNames []string) ([]datastore.RevisionedCaveat, error) {
	return nil, nil
}

func (r *sqliteReader) ListAllCaveats(ctx context.Context) ([]datastore.RevisionedCaveat, error) {
	return r.lookupCaveats(ctx, nil)
}

func (r *sqliteReader) lookupCaveats(ctx context.Context, caveatNames []string) ([]datastore.RevisionedCaveat, error) {
	return nil, nil
}

// func (rwt *pgReadWriteTXN) WriteCaveats(ctx context.Context, caveats []*core.CaveatDefinition) error {
// 	if len(caveats) == 0 {
// 		return nil
// 	}
// 	write := writeCaveat
// 	writtenCaveatNames := make([]string, 0, len(caveats))
// 	for _, caveat := range caveats {
// 		definitionBytes, err := caveat.MarshalVT()
// 		if err != nil {
// 			return fmt.Errorf(errWriteCaveats, err)
// 		}
// 		valuesToWrite := []any{caveat.Name, definitionBytes}
// 		write = write.Values(valuesToWrite...)
// 		writtenCaveatNames = append(writtenCaveatNames, caveat.Name)
// 	}

// 	// mark current caveats as deleted
// 	err := rwt.deleteCaveatsFromNames(ctx, writtenCaveatNames)
// 	if err != nil {
// 		return fmt.Errorf(errWriteCaveats, err)
// 	}

// 	// store the new caveat revision
// 	sql, args, err := write.ToSql()
// 	if err != nil {
// 		return fmt.Errorf(errWriteCaveats, err)
// 	}
// 	if _, err := rwt.tx.Exec(ctx, sql, args...); err != nil {
// 		return fmt.Errorf(errWriteCaveats, err)
// 	}
// 	return nil
// }

// func (rwt *pgReadWriteTXN) DeleteCaveats(ctx context.Context, names []string) error {
// 	// mark current caveats as deleted
// 	return rwt.deleteCaveatsFromNames(ctx, names)
// }

// func (rwt *pgReadWriteTXN) deleteCaveatsFromNames(ctx context.Context, names []string) error {
// 	sql, args, err := deleteCaveat.
// 		Set(colDeletedXid, rwt.newXID).
// 		Where(sq.Eq{colCaveatName: names}).
// 		ToSql()
// 	if err != nil {
// 		return fmt.Errorf(errDeleteCaveats, err)
// 	}

// 	if _, err := rwt.tx.Exec(ctx, sql, args...); err != nil {
// 		return fmt.Errorf(errDeleteCaveats, err)
// 	}
// 	return nil
// }
