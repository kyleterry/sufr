package app

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/kyleterry/sufr/config"
)

var staticHandler = http.StripPrefix(
	"/static/",
	http.FileServer(http.Dir("static")))

func urlIndexHandler(w http.ResponseWriter, r *http.Request) error {
	rawbytes, err := database.GetAll(config.BucketNameURL)
	urls := DeserializeURLs(rawbytes...)
	if err != nil {
		return err
	}

	renderTemplate(w, "url-index", map[string]interface{}{
		"ActiveTab": "urls",
		"Title":     "URLs",
		"Count":     len(urls),
		"URLs":      urls,
	})
	return nil
}

func urlNewHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["ActiveTab"] = "urls"
	fmt.Println(ctx)
	renderTemplate(w, "url-new", ctx)
	return nil
}

func urlSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	urlstring := r.FormValue("url")
	tagsstring := r.FormValue("tags")
	if !govalidator.IsURL(urlstring) {
		errormessage := "URL is required"
		if urlstring != "" {
			errormessage = fmt.Sprintf("URL \"%s\" is not valid", urlstring)
		}
		session, err := store.Get(r, "flashes")
		if err != nil {
			return err
		}
		session.AddFlash(errormessage, "danger")
		session.Save(r, w)
		http.Redirect(w, r, reverse("url-new"), http.StatusSeeOther)
		return nil
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
		return err
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
	return nil
}

func urlViewHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		return err
	}
	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		return err
	}

	url := DeserializeURL(rawbytes)

	renderTemplate(w, "url-view", map[string]interface{}{
		"ActiveTab": "urls",
		"Url":       url,
	})
	return nil
}

func urlEditHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		return err
	}
	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		return err
	}

	url := DeserializeURL(rawbytes)

	renderTemplate(w, "url-edit", map[string]interface{}{
		"ActiveTab": "urls",
		"Url":       url,
	})

	return nil
}

func urlSaveHandler(w http.ResponseWriter, r *http.Request) error {
	titlestring := r.FormValue("title")
	tagsstring := r.FormValue("tags")
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching data: %s", err), http.StatusInternalServerError)
	}

	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		return err
	}

	url := DeserializeURL(rawbytes)
	url.Title = titlestring
	url.UpdatedAt = time.Now()
	err = url.SaveWithTags(tagsstring)
	if err != nil {
		return err
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
	return nil
}

func urlDeleteHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		return err
	}

	rawbytes, err := database.Get(id, config.BucketNameURL)
	if err != nil {
		return err
	}

	url := DeserializeURL(rawbytes)
	err = url.Delete()
	if err != nil {
		return err
	}

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func tagIndexHandler(w http.ResponseWriter, r *http.Request) error {
	rawbytes, err := database.GetAll(config.BucketNameTag)
	tags := DeserializeTags(rawbytes...)
	if err != nil {
		return err
	}
	renderTemplate(w, "tag-index", map[string]interface{}{
		"ActiveTab": "tags",
		"Tags":      tags,
	})
	return nil
}

func tagViewHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint((vars["id"]), 10, 64)
	if err != nil {
		return err
	}
	rawbytes, err := database.Get(id, config.BucketNameTag)
	if err != nil {
		return err
	}

	tag := DeserializeTag(rawbytes)

	renderTemplate(w, "url-index", map[string]interface{}{
		"ActiveTab": "tags",
		"Title":     fmt.Sprintf("URLs tagged under %s", tag.Name),
		"URLs":      tag.GetURLs(),
		"Count":     len(tag.URLs),
	})
	return nil
}
