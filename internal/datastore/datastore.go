package datastore

import (
	"context"
	"errors"

	pb "github.com/authzed/spicedb/pkg/REDACTEDapi/api"
)

const (
	// Ellipsis is a special relation that is assumed to be valid on the right
	// hand side of a tuple.
	Ellipsis = "..."
)

// Publicly facing errors
var (
	ErrNamespaceNotFound  = errors.New("unable to find namespace")
	ErrRelationNotFound   = errors.New("unable to find relation")
	ErrPreconditionFailed = errors.New("unable to satisfy write precondition")
	ErrWatchDisconnected  = errors.New("watch fell too far behind and was disconnected")
	ErrWatchCanceled      = errors.New("watch was canceled by the caller")
	ErrInvalidRevision    = errors.New("revision was invalid")
)

// RevisionChanges represents the changes in a single transaction.
type RevisionChanges struct {
	Revision uint64
	Changes  []*pb.RelationTupleUpdate
}

// Datastore represents tuple access for a single namespace.
type Datastore interface {
	GraphDatastore

	NamespaceManager

	// WriteTuples takes a list of existing tuples that must exist, and a list of tuple
	// mutations and applies it to the datastore for the specified namespace.
	WriteTuples(preconditions []*pb.RelationTuple, mutations []*pb.RelationTupleUpdate) (uint64, error)

	// Revision gets the currently replicated revision for this datastore.
	Revision(ctx context.Context) (uint64, error)

	// SyncRevision gets a revision that is guaranteed to be at least as fresh as right now.
	SyncRevision(ctx context.Context) (uint64, error)

	// Watch notifies the caller about all changes to tuples.
	//
	// All events following afterRevision will be sent to the caller.
	Watch(ctx context.Context, afterRevision uint64) (<-chan *RevisionChanges, <-chan error)

	// WriteNamespace takes a proto namespace definition and persists it,
	// returning the version of the namespace that was created.
	WriteNamespace(newConfig *pb.NamespaceDefinition) (uint64, error)

	// DeleteNamespace deletes a namespace and any associated tuples.
	DeleteNamespace(nsName string) (uint64, error)
}

// GraphDatastore is a subset of the datastore interface that is passed to graph resolvers.
type GraphDatastore interface {
	// QueryTuples creates a builder for reading tuples from the datastore.
	QueryTuples(namespace string, revision uint64) TupleQuery

	// CheckRevision checks the specified revision to make sure it's valid and hasn't been
	// garbage collected.
	CheckRevision(ctx context.Context, revision uint64) error
}

// NamespaceManager is a subset of the datastore interface that can read (and possibly cache) namespaces.
type NamespaceManager interface {
	// ReadNamespace reads a namespace definition and version and returns it if found.
	ReadNamespace(nsName string) (*pb.NamespaceDefinition, uint64, error)
}

// TupleQuery is a builder for constructing tuple queries.
type TupleQuery interface {
	// WithObjectID adds an object ID filter to the query.
	WithObjectID(objectID string) TupleQuery

	// WithRelation adds a relation filter to the query.
	WithRelation(relation string) TupleQuery

	// WithUserset adds a userset filter to the query.
	WithUserset(userset *pb.ObjectAndRelation) TupleQuery

	// Execute runs the tuple query and returns a result iterator.
	Execute() (TupleIterator, error)
}

// TupleIterator is an iterator over matched tuples.
type TupleIterator interface {
	// Next returns the next tuple in the result set.
	Next() *pb.RelationTuple

	// After receiving a nil response, the caller must check for an error.
	Err() error

	// Close cancels the query and closes any open connections.
	Close()
}
