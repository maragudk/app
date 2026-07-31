package service_test

import (
	"fmt"
	"testing"

	"maragu.dev/is"

	"app/model"
	"app/service"
	"app/servicetest"
	"app/sqlitetest"
)

func TestFat_GetUser(t *testing.T) {
	t.Run("gets a user by ID", func(t *testing.T) {
		// Only the user lookup is wired, which is the whole of what this exercises: a database is what
		// the operation is made of, and it can reach nothing else because nothing else was given.
		fat := servicetest.NewFat(t)
		service.GetUser(fat, sqlitetest.NewDatabase(t, sqlitetest.WithFixtures("admin")))

		user, err := fat.GetUser(t.Context(), model.UserID("u_f4958e9cd27a553b08092c790ea44fbb"))
		is.NotError(t, err)
		is.Equal(t, model.UserID("u_f4958e9cd27a553b08092c790ea44fbb"), user.ID)
		is.Equal(t, "Admin", user.Name)
	})

	// An unwired operation is a mistake in composition rather than a runtime condition, so the method
	// says which function was never called instead of failing somewhere inside the operation.
	t.Run("panics naming the wiring function when called unwired", func(t *testing.T) {
		defer func() {
			r := recover()
			is.True(t, r != nil, "expected a panic")
			is.Equal(t, "service: GetUser not wired; call service.GetUser or service.Setup", fmt.Sprint(r))
		}()

		fat := servicetest.NewFat(t)

		_, _ = fat.GetUser(t.Context(), model.UserID("u_f4958e9cd27a553b08092c790ea44fbb"))
	})

	// The other end of the same lifecycle: wiring happens once, so a second one is a mistake too, and
	// the two panics together mean an operation is called neither before nor after its one wiring.
	t.Run("panics when the operation is wired twice", func(t *testing.T) {
		defer func() {
			r := recover()
			is.True(t, r != nil, "expected a panic")
			is.Equal(t, "service: GetUser already wired", fmt.Sprint(r))
		}()

		fat := servicetest.NewFat(t)
		db := sqlitetest.NewDatabase(t)

		service.GetUser(fat, db)
		service.GetUser(fat, db)
	})
}
