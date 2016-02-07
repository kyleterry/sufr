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

func urlIndexHandler(w http.ResponseWriter, r *http.Request) {
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

func urlNewHandler(w http.ResponseWriter, r *http.Request) {
	flashes := make(map[string][]interface{})
	session, err := store.Get(r, "flashes")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flashes["danger"] = session.Flashes("error")
	session.Save(r, w)
	renderTemplate(w, "url-new", map[string]interface{}{
		"ActiveTab": "urls",
		"Flashes":   flashes,
	})
}

func urlSubmitHandler(w http.ResponseWriter, r *http.Request) {
	urlstring := r.FormValue("url")
	tagsstring := r.FormValue("tags")
	if !govalidator.IsURL(urlstring) {
		session, err := store.Get(r, "flashes")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.AddFlash(fmt.Sprintf("URL \"%s\" is not valid", urlstring), "error")
		session.Save(r, w)
		http.Redirect(w, r, reverse("url-new"), http.StatusSeeOther)
		return
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
		return
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
}

func urlViewHandler(w http.ResponseWriter, r *http.Request) {
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

func urlEditHandler(w http.ResponseWriter, r *http.Request) {
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

func urlSaveHandler(w http.ResponseWriter, r *http.Request) {
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

func urlDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

func tagIndexHandler(w http.ResponseWriter, r *http.Request) {
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

func tagViewHandler(w http.ResponseWriter, r *http.Request) {
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
