// Package sqltest provides testing helpers for the sql package.
package sqlitetest

import (
	"testing"

	"maragu.dev/glue/sqlitetest"

	"app/sqlite"
)

// NewDatabase for testing.
func NewDatabase(t *testing.T) *sqlite.Database {
	t.Helper()

	return sqlite.NewDatabase(sqlite.NewDatabaseOptions{H: sqlitetest.NewHelper(t)})
}
