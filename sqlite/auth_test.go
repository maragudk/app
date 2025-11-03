package sqlite_test

import (
	"testing"

	"app/model"
	"app/sqlitetest"

	"maragu.dev/is"
)

func TestGetUser(t *testing.T) {
	t.Run("gets the admin user", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t, sqlitetest.WithFixtures("admin"))

		user, err := db.GetUser(t.Context(), "u_f4958e9cd27a553b08092c790ea44fbb")
		is.NotError(t, err)
		is.Equal(t, model.UserID("u_f4958e9cd27a553b08092c790ea44fbb"), user.ID)
		is.Equal(t, model.AccountID("a_409f852bf39791ccc2496d23f18c63ac"), user.AccountID)
		is.Equal(t, "Admin", user.Name)
		is.Equal(t, model.EmailAddress("admin@example.com"), user.Email)
		is.True(t, user.Confirmed)
		is.True(t, user.Active)
	})

	t.Run("does not get nonexistent user", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t, sqlitetest.WithFixtures("admin"))

		_, err := db.GetUser(t.Context(), "u_nonexistent")
		is.Error(t, model.ErrorUserNotFound, err)
	})
}
