package app

import (
	"net/http"

	"github.com/kyleterry/sufr/config"
)

var staticHandler = http.StripPrefix(
	"/static/",
	http.FileServer(http.Dir("static")))

func (s *Sufr) urlIndexHandler(w http.ResponseWriter, r *http.Request) {
	urls, err := s.DB.GetAll(config.BucketNameURL)
	if err != nil {
		http.Error(w, "Internal Error: Sufr is down!", http.StatusInternalServerError)
	}
	renderTemplate(w, "url-index", map[string]interface{}{
		"ActiveTab": "urls",
		"Urls":      urls,
	})
}

func (s *Sufr) urlNewHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "url-new", map[string]interface{}{
		"ActiveTab": "urls",
	})
}
