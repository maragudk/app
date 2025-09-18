package sqlitetest_test

import (
	"testing"

	"maragu.dev/is"

	"app/sqlitetest"
)

func TestNewDatabase(t *testing.T) {
	t.Run("can get a new database and get the sqlite version and migration id", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)

		var version string
		err := db.H.Get(t.Context(), &version, "select sqlite_version()")
		is.NotError(t, err)
		is.Equal(t, "3.50.4", version)

		var migration string
		err = db.H.Get(t.Context(), &migration, "select version from migrations")
		is.NotError(t, err)
		is.True(t, migration != "")
	})
}
