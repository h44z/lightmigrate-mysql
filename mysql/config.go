package mysql

// DefaultMigrationsTable is the table to use for migration state by default.
const DefaultMigrationsTable = "schema_migrations"

type config struct {
	DatabaseName    string
	MigrationsTable string
	Locking         bool
}
