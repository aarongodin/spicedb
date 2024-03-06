package sqlite

type sqliteOptions struct {
	tablePrefix string
}

// Option provides the facility to configure how clients within the
// Sqlite datastore interact with the running Sqlite driver.
type Option func(*sqliteOptions)

func generateConfig(options []Option) (sqliteOptions, error) {
	computed := sqliteOptions{}
	for _, option := range options {
		option(&computed)
	}
	return computed, nil
}

// TablePrefix allows defining a sqlite table name prefix.
func TablePrefix(prefix string) Option {
	return func(opts *sqliteOptions) {
		opts.tablePrefix = prefix
	}
}
