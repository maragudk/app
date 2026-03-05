package http

import (
	"log/slog"

	"maragu.dev/glue/http"

	"app/service"
)

func InjectHTTPRouter(log *slog.Logger, svc *service.Fat) func(*Router) {
	return func(r *Router) {
		r.Use(AddUserToContext(log, svc))

		Signup(r, log, svc)

		r.Group(func(r *http.Router) {
			Home(r, log)
		})
	}
}
