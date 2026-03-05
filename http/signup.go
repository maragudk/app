package http

import (
	"context"
	"log/slog"
	"net/http"

	"app/model"
)

type signupper interface {
	Signup(ctx context.Context, input model.SignupInput) error
}

func Signup(r *Router, log *slog.Logger, s signupper) {
	r.Mux.Post("/signup", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Error("Error parsing form", "error", err)
			http.Error(w, "error parsing form", http.StatusBadRequest)
			return
		}

		input := model.SignupInput{
			Name:  r.FormValue("name"),
			Email: model.EmailAddress(r.FormValue("email")),
		}

		if err := s.Signup(r.Context(), input); err != nil {
			log.Error("Error signing up", "error", err)
			http.Error(w, "error signing up", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
