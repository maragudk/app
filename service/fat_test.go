package service_test

import (
	"testing"

	"maragu.dev/is"

	"app/model"
	"app/servicetest"
)

func TestFat_Signup(t *testing.T) {
	t.Run("signs up a new user", func(t *testing.T) {
		f := servicetest.NewFat(t)

		err := f.Signup(t.Context(), model.SignupInput{
			Name:  "Test User",
			Email: "test@example.com",
		})
		is.NotError(t, err)
	})
}
