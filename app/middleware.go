package app

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/data"
)

func LoggingHandler(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}

func (a *Sufr) AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !loggedIn(r) {
			http.Redirect(w, r, reverse("login"), http.StatusSeeOther)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (a *Sufr) SetLoggedInHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		ctx["LoggedIn"] = loggedIn(r)
		if ctx["LoggedIn"].(bool) {
			user, err := data.GetUser()
			if err != nil {
				// TODO: Nah, fix this
				panic(err) // if we say we are logged in, but can't get the user, then fucking panic
			}
			ctx["User"] = user
		}
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}

func (a *Sufr) SetActiveTabHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		prefix := ""
		switch {
		case strings.HasPrefix(r.RequestURI, "/url"), r.RequestURI == "/":
			prefix = "urls"
		case strings.HasPrefix(r.RequestURI, "/tag"):
			prefix = "tags"
		case strings.HasPrefix(r.RequestURI, "/import"):
			prefix = "imports"
		}
		ctx["ActiveTab"] = prefix
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}

func (a *Sufr) SetSettingsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		settings, err := data.GetSettings()
		if err != nil {
			// TODO: please don't panic here
			panic(err)
		}

		settingsmap := make(map[string]interface{})
		settingsmap["EmbedPhotos"] = settings.EmbedPhotos
		settingsmap["EmbedVideos"] = settings.EmbedVideos
		settingsmap["PerPage"] = settings.PerPage
		settingsmap["Version"] = config.Version
		settingsmap["BuildTime"] = config.BuildTime
		settingsmap["BuildGitHash"] = config.BuildGitHash
		settingsmap["DataDir"] = config.DataDir
		ctx["Settings"] = settingsmap
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}
