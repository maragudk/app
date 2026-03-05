package http_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"maragu.dev/is"

	apphttp "app/http"
	"app/servicetest"
)

func TestSignup(t *testing.T) {
	t.Run("calls signup with form values", func(t *testing.T) {
		svc := servicetest.NewFat(t)

		mux := chi.NewMux()
		r := &apphttp.Router{Mux: mux}
		apphttp.Signup(r, slog.New(slog.DiscardHandler), svc)

		form := url.Values{
			"name":  {"Glitter Enthusiast"},
			"email": {"glitter@festival.com"},
		}
		req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		is.Equal(t, http.StatusSeeOther, rec.Code)
	})
}
