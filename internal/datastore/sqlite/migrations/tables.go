package migrations

type Tables struct {
	migrationVersion string
	transaction      string
	tuple            string
	namespace        string
	metadata         string
	caveat           string
}

// NewTables creates a struct of table names for this sqlite instance.
// The prefix is optional and can be an empty string.
func NewTables(prefix string) *Tables {
	return &Tables{
		migrationVersion: prefix + "migration_version",
		transaction:      prefix + "relation_tuple_transaction",
		tuple:            prefix + "relation_tuple",
		namespace:        prefix + "namespace_config",
		metadata:         prefix + "metadata",
		caveat:           prefix + "caveat",
	}
}

func (t *Tables) MigrationVersion() string {
	return t.migrationVersion
}

func (t *Tables) Transaction() string {
	return t.transaction
}

func (t *Tables) Tuple() string {
	return t.tuple
}

func (t *Tables) Namespace() string {
	return t.namespace
}

func (t *Tables) Metadata() string {
	return t.metadata
}

func (t *Tables) Caveat() string {
	return t.caveat
}
