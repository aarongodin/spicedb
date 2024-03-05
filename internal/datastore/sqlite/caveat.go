package sqlite

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"

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

func (cc *caveatContext) Scan(val any) error {
	v, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type: %T", v)
	}
	return json.Unmarshal(v, &cc)
}

func (cc *caveatContext) Value() (driver.Value, error) {
	return json.Marshal(&cc)
}

func (r *sqliteReader) ReadCaveatByName(ctx context.Context, name string) (*core.CaveatDefinition, datastore.Revision, error) {
	return nil, nil, nil
}

func (r *sqliteReader) LookupCaveatsWithNames(ctx context.Context, caveatNames []string) ([]datastore.RevisionedCaveat, error) {
	return nil, nil
}

func (r *sqliteReader) ListAllCaveats(ctx context.Context) ([]datastore.RevisionedCaveat, error) {
	return nil, nil
}
