package http_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"maragu.dev/is"

	apphttp "app/http"
	"app/model"
)

type mockSignupper struct {
	called bool
	input  model.SignupInput
	err    error
}

func (m *mockSignupper) Signup(_ context.Context, input model.SignupInput) error {
	m.called = true
	m.input = input
	return m.err
}

func TestSignup(t *testing.T) {
	t.Run("calls signup with form values", func(t *testing.T) {
		mock := &mockSignupper{}
		mux := chi.NewMux()
		r := &apphttp.Router{Mux: mux}
		apphttp.Signup(r, slog.New(slog.DiscardHandler), mock)

		form := url.Values{
			"name":  {"Glitter Enthusiast"},
			"email": {"glitter@festival.com"},
		}
		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		is.True(t, mock.called)
		is.Equal(t, "Glitter Enthusiast", mock.input.Name)
		is.Equal(t, model.EmailAddress("glitter@festival.com"), mock.input.Email)
		is.Equal(t, http.StatusSeeOther, rec.Code)
	})
}
