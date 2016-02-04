package app

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/db"
)

// Sufr is the main application struct. It also implements http.Handler so it can
// be passed directly into ListenAndServe
type Sufr struct {
	Router *mux.Router
	DB     *db.SufrDB
}

// New created a new pointer to Sufr
func New() *Sufr {
	log.Println("Creating new Sufr instance")
	app := &Sufr{Router: mux.NewRouter()}
	app.Router.HandleFunc("/", app.urlIndexHandler)
	app.Router.HandleFunc("/url/new", app.urlNewHandler)
	app.Router.PathPrefix("/").Handler(staticHandler)
	app.DB = db.New(config.DatabaseFile)
	err := app.DB.Open()
	// Panic if we can't open the database
	if err != nil {
		log.Panic(err)
	}
	return app
}

func (s Sufr) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
