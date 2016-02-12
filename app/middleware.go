package app

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/kyleterry/sufr/config"
)

func LoggingHandler(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}

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
		if ctx["LoggedIn"].(bool) {
			user, err := database.Get(uint64(1), config.BucketNameUser)
			if err != nil {
				panic(err) // if we say we are logged in, but can't get the user, then fucking panic
			}
			ctx["User"] = DeserializeUser(user)
		}
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}

func SetActiveTabHandler(h http.Handler) http.Handler {
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

func SetSettingsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		settingsbytes, err := database.Get(uint64(1), config.BucketNameRoot)
		if err != nil {
			panic(err)
		}
		settings := DeserializeSettings(settingsbytes)
		settingsmap := make(map[string]interface{})
		settingsmap["EmbedPhotos"] = settings.EmbedPhotos
		settingsmap["EmbedVideos"] = settings.EmbedVideos
		settingsmap["Version"] = config.Version
		settingsmap["BuildTime"] = config.BuildTime
		settingsmap["BuildGitHash"] = config.BuildGitHash
		ctx["Settings"] = settingsmap
		context.Set(r, TemplateContext, ctx)
		h.ServeHTTP(w, r)
	})
}
