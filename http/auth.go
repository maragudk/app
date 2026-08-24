package http

import (
	"context"
	"log/slog"
	"net/http"

	gluehttp "maragu.dev/glue/http"

	"app/model"
)

const contextUserKey = gluehttp.ContextKey("user")

type userGetter interface {
	GetUser(ctx context.Context, id model.UserID) (model.User, error)
}

// AddUserToContext is [gluehttp.Middleware] to add an authenticated user to the request context, if the user ID is available in the request context.
func AddUserToContext(log *slog.Logger, ug userGetter) gluehttp.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID := gluehttp.GetUserIDFromContext(ctx)
			if userID == nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := ug.GetUser(ctx, *userID)
			if err != nil {
				log.Error("Error getting user from context", "error", err)
				http.Error(w, "error getting user from context", http.StatusBadGateway)
				return
			}

			ctx = context.WithValue(ctx, contextUserKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
