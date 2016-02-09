package app

import (
	"fmt"
	"log"
	"net/http"
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
	router   = mux.NewRouter()
	store    = sessions.NewCookieStore([]byte("I gotta glock in my rari")) // TODO(kt): generate secret key instead of using Fetty Wap lyrics
	database *db.SufrDB
)

const (
	VisibilityPrivate = "private"
	VisibilityPublic  = "public"
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
}

// New created a new pointer to Sufr
func New() *Sufr {
	log.Println("Creating new Sufr instance")
	app := &Sufr{}

	// This route is used to initially configure the instance
	router.Handle("/config", errorHandler(registrationHandler)).Methods("POST", "GET").Name("config")
	router.Handle("/login", errorHandler(loginHandler)).Methods("POST", "GET").Name("login")
	router.Handle("/logout", errorHandler(logoutHandler)).Methods("POST", "GET").Name("logout")

	all := alice.New(SetLoggedInHandler, SetActiveTabHandler)
	auth := alice.New(AuthHandler)
	auth = auth.Extend(all)

	router.Handle("/", all.Then(errorHandler(urlIndexHandler))).Name("url-index")
	router.Handle("/url/new", auth.Then(errorHandler(urlNewHandler))).Name("url-new")
	router.Handle("/url/submit", auth.Then(errorHandler(urlSubmitHandler))).Methods("POST").Name("url-submit")
	router.Handle("/url/{id:[0-9]+}", all.Then(errorHandler(urlViewHandler))).Name("url-view")
	router.Handle("/url/{id:[0-9]+}/edit", auth.Then(errorHandler(urlEditHandler))).Name("url-edit")
	router.Handle("/url/{id:[0-9]+}/save", auth.Then(errorHandler(urlSaveHandler))).Methods("POST").Name("url-save")
	router.Handle("/url/{id:[0-9]+}/delete", auth.Then(errorHandler(urlDeleteHandler))).Name("url-delete")
	router.Handle("/tag", all.Then(errorHandler(tagIndexHandler))).Name("tag-index")
	router.Handle("/tag/{id:[0-9]+}", all.Then(errorHandler(tagViewHandler))).Name("tag-view")
	router.Handle("/import/shitbucket", auth.Then(errorHandler(shitbucketImportHandler))).Methods("POST", "GET").Name("shitbucket-import")
	router.PathPrefix("/").Handler(staticHandler)

	database = db.New(config.DatabaseFile)
	err := database.Open()
	// Panic if we can't open the database
	if err != nil {
		log.Panic(err)
	}

	return app
}

func (s Sufr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Have we configured?
	if r.RequestURI != "/config" &&
		!strings.HasPrefix(r.RequestURI, "/static") &&
		!applicationConfigured() {
		http.Redirect(w, r, reverse("config"), http.StatusSeeOther)
		return
	}

	// Is it a private only instance?
	if r.RequestURI != "/login" && instancePrivate() && !loggedIn(r) {
		http.Redirect(w, r, reverse("login"), http.StatusSeeOther)
		return
	}

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

func getPageTitle(url string) (string, error) {
	doc, err := goquery.NewDocument(url)

	title := doc.Find("title").Text()
	return title, err
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

func applicationConfigured() bool {
	settings, err := database.Get(uint64(1), config.BucketNameRoot)
	if err != nil {
		panic(err)
	}
	if settings != nil {
		return true
	}
	return false
}

func instancePrivate() bool {
	settingsbytes, err := database.Get(uint64(1), config.BucketNameRoot)
	if err != nil {
		panic(err)
	}
	settings := DeserializeSettings(settingsbytes)
	if settings.Visibility == VisibilityPrivate {
		return true
	}
	return false
}

func loggedIn(r *http.Request) bool {
	session, err := store.Get(r, "auth")
	if err != nil {
		panic(err)
	}

	val := session.Values["userID"]
	userID, ok := val.(uint64)

	if !ok || userID <= 0 {
		return false
	}

	user, err := database.Get(userID, config.BucketNameUser)
	if err != nil {
		panic(err)
	}

	if user == nil {
		return false
	}

	return true
}
