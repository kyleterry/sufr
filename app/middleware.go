package app

import (
	"net/http"
	"strings"

	"github.com/gorilla/context"
)

func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !loggedIn(r) {
			http.Redirect(w, r, reverse("login"), http.StatusSeeOther)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func SetLoggedInHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		ctx["LoggedIn"] = loggedIn(r)
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}

func SetActiveTabHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		prefix := ""
		switch {
		case strings.HasPrefix(r.RequestURI, "/url"):
			prefix = "urls"
		case strings.HasPrefix(r.RequestURI, "/tag"):
			prefix = "tags"
		case strings.HasPrefix(r.RequestURI, "/import"):
			prefix = "imports"
		case strings.HasPrefix(r.RequestURI, "/"):
			prefix = "urls" // hack for now
		}
		ctx["ActiveTab"] = prefix
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}
