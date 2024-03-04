package sqlite

import (
	"context"

	"github.com/authzed/spicedb/pkg/datastore"
)

const (
	errRevision       = "unable to find revision: %w"
	errCheckRevision  = "unable to check revision: %w"
	errRevisionFormat = "invalid revision format: %w"
)

func (mdb *sqliteDatastore) OptimizedRevision(_ context.Context) (datastore.Revision, error) {
	return nil, nil
}

func (pgd *sqliteDatastore) HeadRevision(ctx context.Context) (datastore.Revision, error) {
	return nil, nil
}

func (pgd *sqliteDatastore) CheckRevision(ctx context.Context, revisionRaw datastore.Revision) error {
	return nil
}

// RevisionFromString reverses the encoding process performed by MarshalBinary and String.
func (pgd *sqliteDatastore) RevisionFromString(revisionStr string) (datastore.Revision, error) {
	return ParseRevisionString(revisionStr)
}

// ParseRevisionString parses a revision string into a Sqlite revision.
func ParseRevisionString(revisionStr string) (rev datastore.Revision, err error) {
	return nil, nil
}

type sqliteRevision struct {
}

func (pr sqliteRevision) Equal(rhsRaw datastore.Revision) bool {
	// TODO(aarongodin): implement
	return true
}

func (pr sqliteRevision) GreaterThan(rhsRaw datastore.Revision) bool {
	// TODO(aarongodin): implement
	return true
}

func (pr sqliteRevision) LessThan(rhsRaw datastore.Revision) bool {
	// TODO(aarongodin): implement
	return true
}

func (pr sqliteRevision) DebugString() string {
	// TODO(aarongodin): implement
	return ""
}

func (pr sqliteRevision) String() string {
	// TODO(aarongodin): implement
	return ""
}

// MarshalBinary creates a version of the snapshot that uses relative encoding
// for xmax and xip list values to save bytes when encoded as varint protos.
// For example, snapshot 1001:1004:1001,1003 becomes 1000:3:0,2.
func (pr sqliteRevision) MarshalBinary() ([]byte, error) {
	// TODO(aarongodin): implement
	return nil, nil
}

var _ datastore.Revision = sqliteRevision{}
