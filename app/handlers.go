package app

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/kyleterry/sufr/config"
	"golang.org/x/crypto/bcrypt"
)

var staticHandler = http.FileServer(
	&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo},
)

func urlIndexHandler(w http.ResponseWriter, r *http.Request) error {
	rawbytes, err := database.GetDesc(config.BucketNameURL)
	urls := DeserializeURLs(rawbytes...)
	if err != nil {
		return err
	}

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Count"] = len(urls)
	ctx["URLs"] = urls

	return renderTemplate(w, "url-index", ctx)
}

func urlNewHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Add a URL"
	return renderTemplate(w, "url-new", ctx)
}

type URLSchema struct {
	URL     string `schema:"url"`
	Title   string `schema:"title"`
	Tags    string `schema:"tags"`
	Private bool   `schema:"private"`
}

func urlSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	uschema := &URLSchema{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(uschema, r.PostForm); err != nil {
		return err
	}

	urlstring := uschema.URL
	tagsstring := uschema.Tags
	private := uschema.Private
	session, err := store.Get(r, "flashes")
	if err != nil {
		return err
	}
	if !govalidator.IsURL(urlstring) {
		errormessage := "URL is required"
		if urlstring != "" {
			errormessage = fmt.Sprintf("URL \"%s\" is not valid", urlstring)
		}
		session.AddFlash(errormessage, "danger")
		session.Save(r, w)
		http.Redirect(w, r, reverse("url-new"), http.StatusSeeOther)
		return nil
	}

	title, err := getPageTitle(urlstring)
	if err != nil {
		// <strike>Add flash about title not being fetchable</strike>
		// or alternatively add logic for detecting content type because it might be
		// an image or PDF
		session.AddFlash("Sorry! Could not fetch the page title!", "danger")
		session.Save(r, w)
	}

	url := &URL{
		URL:       urlstring,
		Title:     title,
		Private:   private,
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

	if len(rawbytes) <= 0 {
		w.WriteHeader(404)
		return renderTemplate(w, "404", nil)
	}

	url := DeserializeURL(rawbytes)

	if !loggedIn(r) && url.Private {
		w.WriteHeader(404)
		return renderTemplate(w, "404", nil)
	}

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["URL"] = url

	return renderTemplate(w, "url-view", ctx)
}

func urlFavHandler(w http.ResponseWriter, r *http.Request) error {
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

	if url.Favorite {
		url.Favorite = false
	} else {
		url.Favorite = true
	}

	if err := url.Save(); err != nil {
		return err
	}

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

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["URL"] = url
	ctx["Title"] = fmt.Sprintf("Editing %s", url.URL)

	return renderTemplate(w, "url-edit", ctx)
}

func urlSaveHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	uschema := &URLSchema{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(uschema, r.PostForm); err != nil {
		return err
	}

	title := uschema.Title
	tagsstring := uschema.Tags
	private := uschema.Private
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
	url.Title = title
	url.Private = private
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

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Tags"] = tags

	return renderTemplate(w, "tag-index", ctx)
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

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["URLs"] = tag.GetURLs()
	ctx["Count"] = len(tag.URLs)
	ctx["Title"] = fmt.Sprintf("Tagged under %s", tag.Name)
	ctx["IsTagView"] = true
	ctx["Tag"] = tag

	return renderTemplate(w, "url-index", ctx)
}

type ConfigSchema struct {
	Email       string `schema:"email"`
	Password    string `schema:"password"`
	Visibility  string `schema:"visibility"`
	EmbedPhotos bool   `schema:"embedphotos"`
	EmbedVideos bool   `schema:"embedvideos"`
}

func registrationHandler(w http.ResponseWriter, r *http.Request) error {
	if applicationConfigured() {
		http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	}

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Setup"
	if r.Method == "GET" {
		return renderTemplate(w, "registration", ctx)
	}

	session, err := store.Get(r, "flashes")
	if err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	cschema := &ConfigSchema{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(cschema, r.PostForm); err != nil {
		return err
	}

	formErrors := []string{}

	if !govalidator.IsEmail(cschema.Email) {
		if cschema.Email == "" {
			formErrors = append(formErrors, "Email cannot be blank")
		} else {
			formErrors = append(formErrors, fmt.Sprintf("%s is not a valid email address", cschema.Email))
		}
	}
	if cschema.Password == "" {
		formErrors = append(formErrors, "Password cannot be blank")
	}
	if len(formErrors) > 0 {
		for _, msg := range formErrors {
			session.AddFlash(msg, "danger")
		}
		session.Save(r, w)
		http.Redirect(w, r, reverse("config"), http.StatusSeeOther)
	}

	user := User{}
	user.Email = cschema.Email
	passwordCrypt, err := bcrypt.GenerateFromPassword([]byte(cschema.Password), 0)
	if err != nil {
		return err
	}
	user.Password = string(passwordCrypt)

	settings := Settings{}
	settings.Visibility = cschema.Visibility
	settings.EmbedPhotos = cschema.EmbedPhotos
	settings.EmbedVideos = cschema.EmbedVideos

	user.Save()
	settings.Save()

	authsession, err := store.Get(r, "auth")
	authsession.Values["userID"] = user.ID
	authsession.Save(r, w)

	// Otherwise things are good
	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

type LoginSchema struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Login"
	if r.Method == "GET" {
		return renderTemplate(w, "login", ctx)
	}

	session, err := store.Get(r, "flashes")
	if err != nil {
		return err
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	lschema := &LoginSchema{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(lschema, r.PostForm); err != nil {
		return err
	}

	formErrors := []string{}

	if !govalidator.IsEmail(lschema.Email) {
		if lschema.Email == "" {
			formErrors = append(formErrors, "Email cannot be blank")
		} else {
			formErrors = append(formErrors, fmt.Sprintf("%s is not a valid email address", lschema.Email))
		}
	}
	if lschema.Password == "" {
		formErrors = append(formErrors, "Password cannot be blank")
	}
	if len(formErrors) > 0 {
		for _, msg := range formErrors {
			session.AddFlash(msg, "danger")
		}
		session.Save(r, w)
		http.Redirect(w, r, reverse("login"), http.StatusSeeOther)
		return nil
	}

	userbytes, err := database.Get(uint64(1), config.BucketNameUser)
	if err != nil {
		formErrors = append(formErrors, "Email and password did not match")
	}

	user := DeserializeUser(userbytes)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(lschema.Password)); err != nil {
		formErrors = append(formErrors, "Email and password did not match")
	}

	if len(formErrors) > 0 {
		for _, msg := range formErrors {
			session.AddFlash(msg, "danger")
		}
		session.Save(r, w)
		http.Redirect(w, r, reverse("login"), http.StatusSeeOther)
		return nil
	}

	var authsession *sessions.Session
	authsession, err = store.Get(r, "auth")
	if err != nil {
		return err
	}

	authsession.Values["userID"] = user.ID
	authsession.Save(r, w)

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "auth")
	if err != nil {
		return err
	}
	session.Values["userID"] = 0
	session.Save(r, w)

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func settingsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Settings"
	val, err := database.Get(uint64(1), config.BucketNameRoot)
	if err != nil {
		return err
	}
	settings := DeserializeSettings(val)
	if r.Method == "GET" {
		ctx["SettingsObject"] = settings
		return renderTemplate(w, "settings", ctx)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	cschema := &ConfigSchema{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(cschema, r.PostForm); err != nil {
		return err
	}

	settings.Visibility = cschema.Visibility
	settings.EmbedPhotos = cschema.EmbedPhotos
	settings.EmbedVideos = cschema.EmbedVideos

	settings.Save()

	session, err := store.Get(r, "flashes")
	if err != nil {
		return err
	}

	session.AddFlash("Settings have been saved", "success")
	session.Save(r, w)

	http.Redirect(w, r, reverse("settings"), http.StatusSeeOther)
	return nil
}
