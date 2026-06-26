// Package sqlitetest provides testing helpers for the sqlite package.
package sqlitetest

import (
	"testing"

	"maragu.dev/glue/sqlitetest"

	"app/sqlite"
)

// NewDatabase for testing, with optional options.
// Options:
// - [WithFixtures] to load fixtures after migration.
func NewDatabase(t *testing.T, opts ...NewDatabaseOption) *sqlite.Database {
	t.Helper()

	return sqlite.NewDatabase(sqlite.NewDatabaseOptions{H: sqlitetest.NewHelper(t, opts...)})
}

type NewDatabaseOption = sqlitetest.NewHelperOption

func WithFixtures(fixtures ...string) sqlitetest.NewHelperOption {
	return sqlitetest.WithFixtures(fixtures...)
}
