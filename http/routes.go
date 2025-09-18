package http

import (
	"log/slog"

	"maragu.dev/glue/http"
	"maragu.dev/glue/s3"

	"app/sqlite"
)

func InjectHTTPRouter(log *slog.Logger, db *sqlite.Database, bucket *s3.Bucket) func(*Router) {
	return func(r *Router) {
		r.Use(AddUserToContext(log, db))

		// Signup(r, log, db)
		// Login(r, log, db, r.SM)

		r.Group(func(r *http.Router) {
			Home(r, log)
		})
	}
}
