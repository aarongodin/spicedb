package migrations

import "fmt"

func createMigrationVersion(t *tables) string {
	return fmt.Sprintf(`CREATE TABLE %s (
		id INTEGER NOT NULL PRIMARY KEY,
		version TEXT NOT NULL);`,
		t.migrationVersion(),
	)
}

func createNamespaceConfig(t *tables) string {
	return fmt.Sprintf(`CREATE TABLE %s (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		namespace TEXT NOT NULL,
		serialized_config BLOB NOT NULL,
		created_transaction INTEGER NOT NULL,
		deleted_transaction INTEGER NOT NULL DEFAULT 9223372036854775807,
		PRIMARY KEY (namespace, created_transaction),
		UNIQUE (namespace, deleted_transaction));`,
		t.Namespace(),
	)
}

func createRelationTuple(t *tables) string {
	return fmt.Sprintf(`CREATE TABLE %s (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		namespace TEXT NOT NULL,
		object_id TEXT NOT NULL,
		relation TEXT NOT NULL,
		userset_namespace TEXT NOT NULL,
		userset_object_id TEXT NOT NULL,
		userset_relation TEXT NOT NULL,
		created_transaction INTEGER NOT NULL,
		deleted_transaction INTEGER NOT NULL DEFAULT 9223372036854775807,
		caveat_name TEXT,
		caveat_context JSON,
		UNIQUE (namespace, object_id, relation, userset_namespace, userset_object_id, userset_relation, created_transaction, deleted_transaction),
		UNIQUE (namespace, object_id, relation, userset_namespace, userset_object_id, userset_relation, deleted_transaction));
    CREATE INDEX ix_relation_tuple_by_subject ON %s (userset_object_id, userset_namespace, userset_relation, namespace, relation);
    CREATE INDEX ix_relation_tuple_by_subject_relation ON %s (userset_namespace, userset_relation, namespace, relation);
    CREATE INDEX ix_relation_tuple_by_deleted_transaction ON %s (deleted_transaction);`,
		t.RelationTuple(), t.RelationTuple(), t.RelationTuple(), t.RelationTuple(),
	)
}

func createRelationTupleTransaction(t *tables) string {
	return fmt.Sprintf(`CREATE TABLE %s (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL);
    CREATE INDEX ix_relation_tuple_transaction_by_timestamp ON %s (timestamp);`,
		t.RelationTupleTransaction(), t.RelationTupleTransaction(),
	)
}

func createCaveatTable(t *tables) string {
	return fmt.Sprintf(`CREATE TABLE %s (
    name TEXT NOT NULL,
    definition BLOB NOT NULL,
    created_transaction INTEGER NOT NULL,
    deleted_transaction INTEGER NOT NULL DEFAULT 9223372036854775807,
    PRIMARY KEY (name, deleted_transaction),
    UNIQUE (name, created_transaction, deleted_transaction));`,
		t.Caveat(),
	)
}

func init() {
	mustRegisterMigration("initial", "", noNonatomicMigration,
		newStatementBatch(
			createMigrationVersion,
			createNamespaceConfig,
			createRelationTuple,
			createRelationTupleTransaction,
			createCaveatTable,
		).execute,
	)
}
