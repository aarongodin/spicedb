package sqlite

const (
	tableNamespaceDefault   = "namespace_config"
	tableTransactionDefault = "relation_tuple_transaction"
	tableTupleDefault       = "relation_tuple"
	tableMigrationVersion   = "migration_version"
	tableMetadataDefault    = "metadata"
	tableCaveatDefault      = "caveat"
)

type Tables struct {
	tableMigrationVersion string
	tableTransaction      string
	tableTuple            string
	tableNamespace        string
	tableMetadata         string
	tableCaveat           string
}

// NewTables creates a struct of table names for this sqlite instance.
// The prefix is optional and can be an empty string.
func NewTables(prefix string) *Tables {
	return &Tables{
		tableMigrationVersion: prefix + tableMigrationVersion,
		tableTransaction:      prefix + tableTransactionDefault,
		tableTuple:            prefix + tableTupleDefault,
		tableNamespace:        prefix + tableNamespaceDefault,
		tableMetadata:         prefix + tableMetadataDefault,
		tableCaveat:           prefix + tableCaveatDefault,
	}
}

func (t *Tables) MigrationVersion() string {
	return t.tableMigrationVersion
}

// RelationTupleTransaction returns the prefixed transaction table name.
func (t *Tables) RelationTupleTransaction() string {
	return t.tableTransaction
}

// RelationTuple returns the prefixed relationship tuple table name.
func (t *Tables) RelationTuple() string {
	return t.tableTuple
}

// Namespace returns the prefixed namespace table name.
func (t *Tables) Namespace() string {
	return t.tableNamespace
}

func (t *Tables) Metadata() string {
	return t.tableMetadata
}

func (t *Tables) Caveat() string {
	return t.tableCaveat
}
