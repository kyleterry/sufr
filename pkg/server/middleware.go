package server

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kyleterry/sufr/pkg/store"
)

type userContextKey struct{}

type middlewareFunc func(http.Handler) http.Handler

func NewSessionAuthenticationMiddleware(store sessions.Store, db store.Manager) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			session, err := store.New(r, userAuthSessionKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			doRedirect := func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}

			raw, ok := session.Values["userID"]
			if !ok {
				doRedirect(w, r)

				return
			}

			id, ok := raw.(string)
			if !ok || id == "" {
				doRedirect(w, r)

				return
			}

			user, err := db.Users().GetByID(ctx, id)
			if err != nil {
				log.Println(err)
				doRedirect(w, r)

				return
			}

			ctx = context.WithValue(ctx, userContextKey{}, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
