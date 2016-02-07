package app

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
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

// Sufr is the main application struct. It also implements http.Handler so it can
// be passed directly into ListenAndServe
type Sufr struct {
}

// New created a new pointer to Sufr
func New() *Sufr {
	log.Println("Creating new Sufr instance")
	app := &Sufr{}
	router.HandleFunc("/", app.urlIndexHandler).Name("url-index")
	router.HandleFunc("/url/new", app.urlNewHandler).Name("url-new")
	router.HandleFunc("/url/submit", app.urlSubmitHandler).Methods("POST").Name("url-submit")
	router.HandleFunc("/url/{id:[0-9]+}", app.urlViewHandler).Name("url-view")
	router.HandleFunc("/url/{id:[0-9]+}/edit", app.urlEditHandler).Name("url-edit")
	router.HandleFunc("/url/{id:[0-9]+}/save", app.urlSaveHandler).Methods("POST").Name("url-save")
	router.HandleFunc("/url/{id:[0-9]+}/delete", app.urlDeleteHandler).Name("url-delete")
	router.HandleFunc("/tag", app.tagIndexHandler).Name("tag-index")
	router.HandleFunc("/tag/{id:[0-9]+}", app.tagViewHandler).Name("tag-view")
	router.HandleFunc("/import/shitbucket", app.shitbucketImportHandler).Methods("POST", "GET").Name("shitbucket-import")
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
		panic(err)
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
