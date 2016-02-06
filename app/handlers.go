package app

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/kyleterry/sufr/config"
)

var staticHandler = http.StripPrefix(
	"/static/",
	http.FileServer(http.Dir("static")))

func (s *Sufr) urlIndexHandler(w http.ResponseWriter, r *http.Request) {
	rawbytes, err := database.GetAll(config.BucketNameURL)
	urls := DeserializeURLs(rawbytes...)
	if err != nil {
		http.Error(w, "Internal Error: Sufr is down!", http.StatusInternalServerError)
	}
	renderTemplate(w, "url-index", map[string]interface{}{
		"ActiveTab": "urls",
		"Title":     "URLs",
		"Count":     len(urls),
		"URLs":      urls,
	})
}

func (s *Sufr) urlNewHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "url-new", map[string]interface{}{
		"ActiveTab": "urls",
	})
}

func (s *Sufr) urlSubmitHandler(w http.ResponseWriter, r *http.Request) {
	urlstring := r.FormValue("url")
	tagsstring := r.FormValue("tags")
	// validate URL here
	if !govalidator.IsURL(urlstring) {
		// flash add error
		// redirect back to url-new
	}
	title, err := getPageTitle(urlstring)
	if err != nil {
		// Add flash about title not being fetchable
		// or alternatively add logic for detecting content type because it might be
		// an image or PDF
	}

	url := &URL{
		URL:       urlstring,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = url.SaveWithTags(tagsstring)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving URL: %s", err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
}

func (s *Sufr) urlViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}
	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	url := DeserializeURL(rawbytes)

	renderTemplate(w, "url-view", map[string]interface{}{
		"ActiveTab": "urls",
		"Url":       url,
	})
}

func (s *Sufr) urlEditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}
	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	url := DeserializeURL(rawbytes)

	renderTemplate(w, "url-edit", map[string]interface{}{
		"ActiveTab": "urls",
		"Url":       url,
	})

}

func (s *Sufr) urlSaveHandler(w http.ResponseWriter, r *http.Request) {
	titlestring := r.FormValue("title")
	tagsstring := r.FormValue("tags")
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	url := DeserializeURL(rawbytes)
	url.Title = titlestring
	url.UpdatedAt = time.Now()
	err = url.SaveWithTags(tagsstring)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving URL: %s", err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
}

func (s *Sufr) urlDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	url := DeserializeURL(rawbytes)
	err = url.Delete()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting URL: %s", err), http.StatusInternalServerError)
	}

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
}

func (s *Sufr) tagIndexHandler(w http.ResponseWriter, r *http.Request) {
	rawbytes, err := database.GetAll(config.BucketNameTag)
	tags := DeserializeTags(rawbytes...)
	if err != nil {
		http.Error(w, "Internal Error: Sufr is down!", http.StatusInternalServerError)
	}
	renderTemplate(w, "tag-index", map[string]interface{}{
		"ActiveTab": "tags",
		"Tags":      tags,
	})
}

func (s *Sufr) tagViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}
	rawbytes, err := database.Get(id, config.BucketNameTag)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	tag := DeserializeTag(rawbytes)

	renderTemplate(w, "url-index", map[string]interface{}{
		"ActiveTab": "tags",
		"Title":     fmt.Sprintf("URLs tagged under %s", tag.Name),
		"URLs":      tag.GetURLs(),
		"Count":     len(tag.URLs),
	})
}
