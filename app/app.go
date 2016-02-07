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
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/db"
)

var (
	router   = mux.NewRouter()
	store    = sessions.NewCookieStore([]byte("I gotta glock in my rari")) // TODO(kt): generate secret key instead of using Fetty Wap lyrics
	database *db.SufrDB
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

	router.Handle("/", errorHandler(urlIndexHandler)).Name("url-index")
	router.Handle("/url/new", errorHandler(urlNewHandler)).Name("url-new")
	router.Handle("/url/submit", errorHandler(urlSubmitHandler)).Methods("POST").Name("url-submit")
	router.Handle("/url/{id:[0-9]+}", errorHandler(urlViewHandler)).Name("url-view")
	router.Handle("/url/{id:[0-9]+}/edit", errorHandler(urlEditHandler)).Name("url-edit")
	router.Handle("/url/{id:[0-9]+}/save", errorHandler(urlSaveHandler)).Methods("POST").Name("url-save")
	router.Handle("/url/{id:[0-9]+}/delete", errorHandler(urlDeleteHandler)).Name("url-delete")
	router.Handle("/tag", errorHandler(tagIndexHandler)).Name("tag-index")
	router.Handle("/tag/{id:[0-9]+}", errorHandler(tagViewHandler)).Name("tag-view")
	router.Handle("/import/shitbucket", errorHandler(shitbucketImportHandler)).Methods("POST", "GET").Name("shitbucket-import")
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
	cxt := make(map[string]interface{})
	flashes := make(map[string][]interface{})
	session, err := store.Get(r, "flashes")
	if err != nil {
		log.Println(err)
	}
	flashes["danger"] = session.Flashes("danger")
	flashes["success"] = session.Flashes("success")
	flashes["warning"] = session.Flashes("warning")
	cxt["Flashes"] = flashes
	session.Save(r, w)

	context.Set(r, TemplateContext, cxt)

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
