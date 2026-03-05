package service_test

import (
	"testing"

	"maragu.dev/is"

	"app/model"
	"app/service"
	"app/sqlitetest"
)

func TestFat_Signup(t *testing.T) {
	t.Run("signs up a new user", func(t *testing.T) {
		db := sqlitetest.NewDatabase(t)
		f := &service.Fat{DB: db}

		err := f.Signup(t.Context(), model.SignupInput{
			Name:  "Test User",
			Email: "test@example.com",
		})
		is.NotError(t, err)
	})
}
