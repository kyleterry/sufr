package app

import (
	"context"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/kyleterry/sufr/pkg/config"
	"github.com/kyleterry/sufr/pkg/data"
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

func LoggedInOrAPITokenAuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pass, ok := r.BasicAuth(); ok {
			if apiToken, _ := data.GetAPIToken(); apiToken != nil {
				if parsedToken, err := uuid.Parse(pass); err == nil {
					if parsedToken == apiToken.Token {
						h.ServeHTTP(w, r)

						return
					}
				}
			}
		}

		h = AuthHandler(h)

		h.ServeHTTP(w, r)
	})
}

func SetLoggedInHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loggedInKey, loggedIn(r))
		if loggedIn(r) {
			user, err := data.GetUser()
			if err != nil {
				// TODO: Nah, fix this
				panic(err) // if we say we are logged in, but can't get the user, then fucking panic
			}
			ctx = context.WithValue(ctx, userKey, user)
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetSettingsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settings, err := data.GetSettings()
		if err != nil {
			// TODO: please don't panic here
			panic(err)
		}

		apiToken, _ := data.GetAPIToken()
		if apiToken == nil {
			apiToken = &data.APIToken{}
		}

		sm := make(map[string]interface{})
		sm["EmbedPhotos"] = settings.EmbedPhotos
		sm["EmbedVideos"] = settings.EmbedVideos
		sm["PerPage"] = settings.PerPage
		sm["Version"] = config.Version
		sm["BuildTime"] = config.BuildTime
		sm["BuildGitHash"] = config.BuildGitHash
		// TODO pull this from config struct
		sm["DataDir"] = config.DataDir
		sm["APIToken"] = apiToken.Token

		ctx := context.WithValue(r.Context(), settingsKey, sm)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetPinnedTagsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pinnedTags, err := data.GetPinnedTags()
		if err != nil {
			panic(err)
		}

		ctx := context.WithValue(r.Context(), pinnedTagsKey, pinnedTags)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
