package app

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/db"
)

var (
	router = mux.NewRouter()
	store  = sessions.NewCookieStore([]byte("I gotta glock in my rari")) // TODO(kt): generate secret key instead of using Fetty Wap lyrics
)

const (
	VisibilityPrivate = "private"
	VisibilityPublic  = "public"
)

var (
	SUFRUserAgent = fmt.Sprintf("Linux:SUFR:v%s", config.Version)
)

type templatecontext int

const TemplateContext templatecontext = 0

type errorHandler func(http.ResponseWriter, *http.Request) error

func (fn errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		log.Printf("Got error while processing the request: %s\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Sufr is the main application struct. It also implements http.Handler so it can
// be passed directly into ListenAndServe
type Sufr struct {
	db *db.SufrDB
}

// New created a new pointer to Sufr
func New() *Sufr {
	log.Println("Creating new Sufr instance")
	app := &Sufr{}

	app.db = db.New(config.DatabaseFile)
	if config.Debug {
		go app.db.Statsdumper()
	}
	err := app.db.Open()
	// Panic if we can't open the database
	if err != nil {
		log.Panic(err)
	}

	// This route is used to initially configure the instance
	router.Handle("/config", errorHandler(app.registrationHandler)).Methods("POST", "GET").Name("config")
	router.Handle("/login", errorHandler(app.loginHandler)).Methods("POST", "GET").Name("login")
	router.Handle("/logout", errorHandler(app.logoutHandler)).Methods("POST", "GET").Name("logout")

	all := alice.New(app.SetSettingsHandler, app.SetLoggedInHandler, app.SetActiveTabHandler, LoggingHandler)
	auth := alice.New(app.AuthHandler)
	auth = auth.Extend(all)

	router.Handle("/", all.Then(errorHandler(app.urlIndexHandler))).Name("url-index")

	urlrouter := router.PathPrefix("/url").Subrouter()
	urlrouter.Handle("/new", auth.Then(errorHandler(app.urlNewHandler))).Name("url-new")
	urlrouter.Handle("/submit", auth.Then(errorHandler(app.urlSubmitHandler))).Methods("POST").Name("url-submit")
	urlrouter.Handle("/{id:[0-9]+}", all.Then(errorHandler(app.urlViewHandler))).Name("url-view")
	urlrouter.Handle("/{id:[0-9]+}/edit", auth.Then(errorHandler(app.urlEditHandler))).Name("url-edit")
	urlrouter.Handle("/{id:[0-9]+}/save", auth.Then(errorHandler(app.urlSaveHandler))).Methods("POST").Name("url-save")
	urlrouter.Handle("/{id:[0-9]+}/delete", auth.Then(errorHandler(app.urlDeleteHandler))).Name("url-delete")
	urlrouter.Handle("/{id:[0-9]+}/toggle-fav", auth.Then(errorHandler(app.urlFavHandler))).Methods("POST").Name("url-fav-toggle")

	tagrouter := router.PathPrefix("/tag").Subrouter()
	tagrouter.Handle("/", all.Then(errorHandler(app.tagIndexHandler))).Name("tag-index")
	tagrouter.Handle("/{id:[0-9]+}", all.Then(errorHandler(app.tagViewHandler))).Name("tag-view")

	router.Handle("/import/shitbucket", auth.Then(errorHandler(app.shitbucketImportHandler))).Methods("POST", "GET").Name("shitbucket-import")
	router.Handle("/settings", auth.Then(errorHandler(app.settingsHandler))).Methods("POST", "GET").Name("settings")
	router.Handle("/database-backup", auth.Then(errorHandler(app.db.BackupHandler))).Methods("GET").Name("database-backup")
	router.PathPrefix("/static").Handler(staticHandler)

	router.NotFoundHandler = errorHandler(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		return renderTemplate(w, "404", nil)
	})

	return app
}

func (s Sufr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := make(map[string]interface{})
	flashes := make(map[string][]interface{})
	session, err := store.Get(r, "flashes")
	if err != nil {
		log.Println(err)
	}
	flashes["danger"] = session.Flashes("danger")
	flashes["success"] = session.Flashes("success")
	flashes["warning"] = session.Flashes("warning")
	ctx["Flashes"] = flashes
	session.Save(r, w)

	context.Set(r, TemplateContext, ctx)

	// Have we configured?
	if !s.applicationConfigured() {
		if r.RequestURI != "/config" &&
			!strings.HasPrefix(r.RequestURI, "/static") {
			http.Redirect(w, r, reverse("config"), http.StatusSeeOther)
			return
		}
		router.ServeHTTP(w, r)
		return
	}

	// Is it a private only instance?
	if r.RequestURI != "/login" && s.instancePrivate() && !s.loggedIn(r) && !strings.HasPrefix(r.RequestURI, "/static") {
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
	url, err := router.GetRoute(name).URL(s...)
	if err != nil {
		log.Println(err)
	}
	return url.Path
}

// Returns the page title or an error. If there is an error, the url is returned as well.
func getPageTitle(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return url, err
	}

	req.Header.Set("User-Agent", SUFRUserAgent)

	res, err := client.Do(req)
	if err != nil {
		return url, err
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return url, err
	}

	title := doc.Find("title").Text()
	return title, nil
}

func parseTags(tagsstring string) []string {
	return strings.FieldsFunc(tagsstring, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	})
}

func parseTagsMap(tagsstring string) map[string]struct{} {
	var m = make(map[string]struct{})
	for _, t := range parseTags(tagsstring) {
		m[t] = struct{}{}
	}
	return m
}

func ui64toa(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func (a *Sufr) applicationConfigured() bool {
	settings, err := a.db.Get(uint64(1), config.BucketNameRoot)
	if err != nil {
		panic(err)
	}
	if settings != nil {
		return true
	}
	return false
}

func (a *Sufr) instancePrivate() bool {
	settingsbytes, err := a.db.Get(uint64(1), config.BucketNameRoot)
	if err != nil {
		panic(err)
	}
	settings := DeserializeSettings(settingsbytes)
	if settings.Visibility == VisibilityPrivate {
		return true
	}
	return false
}

func (a *Sufr) loggedIn(r *http.Request) bool {
	session, err := store.Get(r, "auth")
	if err != nil {
		panic(err)
	}

	val := session.Values["userID"]
	userID, ok := val.(uint64)

	if !ok || userID <= 0 {
		return false
	}

	user, err := a.db.Get(userID, config.BucketNameUser)
	if err != nil {
		panic(err)
	}

	if user == nil {
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
var imgurRE = regexp.MustCompile(`^(.{5})\.?[a-zA-z]*$`)

func imgurgal(gal string) string {
	u, _ := url.Parse(gal)

	parts := strings.Split(u.Path[1:], "/")
	if len(parts) > 1 {
		if parts[0] == "gallery" {
			parts[0] = "a"
		}
	} else {
		b := imgurRE.Find([]byte(parts[0]))
		if len(b) > 0 {
			parts[0] = string(b)
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
