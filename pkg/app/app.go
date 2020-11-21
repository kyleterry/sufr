package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	gorsess "github.com/gorilla/sessions"
	"github.com/justinas/alice"
	"github.com/kyleterry/sufr/pkg/config"
	"github.com/kyleterry/sufr/pkg/data"
	"github.com/pkg/errors"
)

var (
	router = mux.NewRouter()
	store  = gorsess.NewCookieStore([]byte("I gotta glock in my rari")) // TODO(kt): generate secret key instead of using Fetty Wap lyrics
)

// Context key types so we don't clobber the global context store
type ctxKeyTemplateData int
type ctxKeyUser int
type ctxKeyFlashes int
type ctxKeyLoggedIn int
type ctxKeySettings int
type ctxKeyPinnedTags int
type ctxKeyAPIToken int

const (
	templateDataKey ctxKeyTemplateData = 0
	userKey         ctxKeyUser         = 0
	flashesKey      ctxKeyFlashes      = 0
	loggedInKey     ctxKeyLoggedIn     = 0
	settingsKey     ctxKeySettings     = 0
	pinnedTagsKey   ctxKeyPinnedTags   = 0
	apiTokenKey     ctxKeyAPIToken     = 0
)

type errorHandler func(http.ResponseWriter, *http.Request) error

func (fn errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		log.Printf("Got error while processing the request: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// type SufrServerOptions struct {
// 	DB           *data.SufrDB
// 	SessionStore *sessions.CookieStore
// }

// Sufr is the main application struct. It also implements http.Handler so it can
// be passed directly into ListenAndServe
type Sufr struct {
	cfg      *config.Config
	db       *data.SufrDB
	sessions *gorsess.CookieStore
}

// New created a new pointer to Sufr
func New(cfg *config.Config) *Sufr {
	// app := &Sufr{
	// 	db:           opts.DB,
	// 	sessionStore: opts.SessionStore,
	// }

	app := &Sufr{}

	// Wrapped middlware
	all := alice.New(SetSettingsHandler, SetLoggedInHandler, LoggingHandler, SetPinnedTagsHandler)
	auth := alice.New(AuthHandler)
	auth = auth.Extend(all)
	apiAuth := alice.New(LoggedInOrAPITokenAuthHandler)
	apiAuth = apiAuth.Extend(all)

	const idPattern = "{id:(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}"

	// This route is used to initially configure the instance
	router.Handle("/config", LoggingHandler(errorHandler(app.registrationHandler))).
		Methods("POST", "GET").
		Name("config")

	router.Handle("/login", errorHandler(app.loginHandler)).
		Methods("POST", "GET").
		Name("login")

	router.Handle("/logout", errorHandler(app.logoutHandler)).
		Methods("POST", "GET").
		Name("logout")

	router.Handle("/", all.Then(errorHandler(app.urlIndexHandler))).
		Name("url-index")

	urlrouter := router.PathPrefix("/url").Subrouter()

	urlrouter.Handle("/favorites", all.Then(errorHandler(app.urlFavoritesHandler))).
		Methods("GET").
		Name("url-favorites")

	urlrouter.Handle("/new", auth.Then(errorHandler(app.urlNewHandler))).
		Methods("GET").
		Name("url-new")

	urlrouter.Handle("/submit", auth.Then(errorHandler(app.urlSubmitHandler))).
		Methods("POST").
		Name("url-submit")

	urlrouter.Handle("/"+idPattern, all.Then(errorHandler(app.urlViewHandler))).
		Methods("GET").
		Name("url-view")

	urlrouter.Handle("/"+idPattern+"/edit", auth.Then(errorHandler(app.urlEditHandler))).
		Methods("GET").
		Name("url-edit")

	urlrouter.Handle("/"+idPattern+"/save", auth.Then(errorHandler(app.urlSaveHandler))).
		Methods("POST").
		Name("url-save")

	// this should use the DELETE method
	urlrouter.Handle("/"+idPattern+"/delete", auth.Then(errorHandler(app.urlDeleteHandler))).
		Name("url-delete")

	urlrouter.Handle("/"+idPattern+"/toggle-fav", auth.Then(errorHandler(app.urlToggleFavoriteHandler))).
		Methods("POST").
		Name("url-fav-toggle")

	tagrouter := router.PathPrefix("/tag").Subrouter()

	tagrouter.Handle("/"+idPattern, all.Then(errorHandler(app.tagViewHandler))).
		Methods("GET").
		Name("tag-view")

	router.Handle("/settings", auth.Then(errorHandler(app.settingsHandler))).
		Methods("POST", "GET").
		Name("settings")

	tokenRouter := router.PathPrefix("/api-token").Subrouter()

	tokenRouter.Handle("/roll", all.Then(errorHandler(app.apiTokenRollHandler))).
		Methods("GET").
		Name("api-token-roll")
	tokenRouter.Handle("/delete", all.Then(errorHandler(app.apiTokenDeleteHandler))).
		Methods("GET").
		Name("api-token-delete")

	router.Handle("/search", all.Then(errorHandler(app.searchHandler))).
		Methods("GET").
		Name("search")

	router.Handle("/database-backup", apiAuth.Then(errorHandler(data.BackupHandler))).
		Methods("GET").
		Name("database-backup")

	router.Handle("/healthz", errorHandler(app.healthzHandler)).Methods("GET").Name("healthz")

	router.PathPrefix("/static").Handler(LoggingHandler(staticHandler))

	router.NotFoundHandler = errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		return renderTemplate(w, r, "404")
	})

	return app
}

func (s Sufr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Get session data
	// 2. Make flashes structure
	// 3. Build on context
	// 4. Check if application is configured
	// 5. Check if the instance is globally private
	// 6. Route using ShiftPath(p)
	// 7. Cleanup

	session, err := store.Get(r, "flashes")
	if err != nil {
		log.Println(err)
	}

	flashes := make(map[string][]interface{})
	flashes["danger"] = session.Flashes("danger")
	flashes["success"] = session.Flashes("success")
	flashes["warning"] = session.Flashes("warning")

	ctx := context.WithValue(r.Context(), flashesKey, flashes)
	templateData := make(map[string]interface{})
	templateData["RequestURI"] = r.RequestURI
	ctx = context.WithValue(ctx, templateDataKey, templateData)

	r = r.WithContext(ctx)

	session.Save(r, w)

	// Have we configured?
	if !applicationConfigured() {
		if r.RequestURI != "/config" &&
			!strings.HasPrefix(r.RequestURI, "/static") {
			http.Redirect(w, r, reverse("config"), http.StatusSeeOther)
			return
		}
		router.ServeHTTP(w, r)
		return
	}

	// Is it a private only instance?
	if r.RequestURI != "/login" && instancePrivate() && !loggedIn(r) && !strings.HasPrefix(r.RequestURI, "/static") {
		http.Redirect(w, r, reverse("login"), http.StatusSeeOther)
		return
	}

	router.ServeHTTP(w, r)
}

// reverse Uses gorilla mux to give us a uri path by name. This is attached to template Funcs
func reverse(name string, params ...interface{}) string {
	s := make([]string, len(params))

	for _, param := range params {
		s = append(s, fmt.Sprint(param))
	}

	route := router.Get(name)
	if route == nil {
		log.Printf("ERROR: %s is not a valid route index", name)
		return ""
	}

	url, err := route.URL(s...)
	if err != nil {
		log.Println(err)
		return ""
	}

	return url.Path
}

func applicationConfigured() bool {
	_, err := data.GetSettings()
	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			return false
		}

		panic(err)
	}

	return true
}

func instancePrivate() bool {
	settings, err := data.GetSettings()
	if err != nil {
		if err == data.ErrNotFound {
			return false
		}

		panic(err)
	}

	if settings.Private {
		return true
	}

	return false
}

func loggedIn(r *http.Request) bool {
	session, err := store.New(r, "auth")
	if err != nil {
		panic(err)
	}

	val := session.Values["userID"]
	if val == nil {
		return false
	}

	id, err := uuid.ParseBytes(val.([]byte))
	if err != nil {
		return false
	}

	user, err := data.GetUser()
	if err != nil {
		if err == data.ErrNotFound {
			return false
		}

		panic(err)
	}

	if user.ID != id {
		return false
	}

	return true
}

func isYoutube(url string) bool {
	return strings.Contains(url, "youtube.com/watch")
}

func youtubevid(video string) string {
	u, _ := url.Parse(video)
	return u.Query()["v"][0]
}

func isImgur(url string) bool {
	// TODO add blacklist for other types of imgur links
	return strings.Contains(url, "imgur.com")
}

// Will match things like FyCch, FyCch.jpg and so on so we can embed raw image links too
// Looks for an id of 3 to 8 chars long.
// TODO: Look at the Imgur API to confirm this :^)
var imgurRE = regexp.MustCompile(`^(?P<id>([a-zA-Z0-9]{3,8}))\.?[a-zA-z]*$`)

func imgurgal(gal string) string {
	u, _ := url.Parse(gal)

	parts := strings.Split(u.Path[1:], "/")
	if len(parts) > 1 {
		if parts[0] == "gallery" {
			parts[0] = "a"
		}
	} else {
		match := imgurRE.FindStringSubmatch(parts[0])
		if len(match) > 1 {
			parts[0] = match[1]
		}
	}

	return strings.Join(parts, "/")
}

func newcontext(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func updatePage(uri string, page int) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("page", strconv.Itoa(page))

	u.RawQuery = q.Encode()

	return u.String(), nil
}
