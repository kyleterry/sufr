package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/kyleterry/sufr/data"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var staticHandler = http.FileServer(
	&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo},
)

func (a *Sufr) urlIndexHandler(w http.ResponseWriter, r *http.Request) error {
	urls, err := data.GetURLs()
	if err != nil {
		return err
	}

	// TODO: do pagination for urls
	// urlCount, err := a.db.BucketLength(config.BucketNameURL)
	// if err != nil {
	// 	return err
	// }

	// pagestr := r.URL.Query().Get("page")
	// if pagestr == "" {
	// 	pagestr = "1"
	// }
	// page, err := strconv.ParseInt(pagestr, 10, 64)
	// if err != nil {
	// 	return err
	// }

	// paginator := NewPaginator(urlCount, int(page), config.DefaultPerPage)
	// rawbytes, err := paginator.GetObjects(config.BucketNameURL)
	// if err != nil {
	// 	return err
	// }

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Count"] = len(urls)
	ctx["URLs"] = urls
	// ctx["Paginator"] = paginator

	return renderTemplate(w, "url-index", ctx)
}

func (a *Sufr) urlNewHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Add a URL"
	return renderTemplate(w, "url-new", ctx)
}

func (a *Sufr) urlSubmitHandler(w http.ResponseWriter, r *http.Request) error {
	decoder := schema.NewDecoder()

	if err := r.ParseForm(); err != nil {
		return err
	}

	createURLOptions := data.CreateURLOptions{}

	if err := decoder.Decode(&createURLOptions, r.PostForm); err != nil {
		return err
	}

	// session, err := store.Get(r, "flashes")
	// if err != nil {
	// 	return err
	// }

	// if !govalidator.IsURL(createURLOptions.URL) {
	// 	errormessage := "URL is required"
	// 	if urlstring != "" {
	// 		errormessage = fmt.Sprintf("URL \"%s\" is not valid", urlstring)
	// 	}
	// 	//TODO: wrap AddFlash in session.Flash()
	// 	session.AddFlash(errormessage, "danger")
	// 	session.Save(r, w)
	// 	http.Redirect(w, r, reverse("url-new"), http.StatusSeeOther)
	// 	return nil
	// }

	// TODO: make and check for a validation error
	// session.FlashValidationError(err)
	url, err := data.CreateURL(createURLOptions)
	if err != nil {
		return err
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
	return nil
}

func (a *Sufr) urlViewHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	url, err := data.GetURL(id)
	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return renderTemplate(w, "404", nil)
		}

		return errors.Wrap(err, "failed to get url")
	}

	if !loggedIn(r) && url.Private {
		w.WriteHeader(404)
		return renderTemplate(w, "404", nil)
	}

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["URL"] = url

	return renderTemplate(w, "url-view", ctx)
}

func (a *Sufr) urlFavHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	url, err := data.GetURL(id)

	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return renderTemplate(w, "404", nil)
		}

		return errors.Wrap(err, "failed to get url")
	}

	err = url.ToggleFavorite()
	if err != nil {
		return errors.Wrap(err, "failed to toggle favorite flag")
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(struct {
		State bool `json:"state"`
	}{url.Favorite})

	if err != nil {
		return err
	}

	w.Write(response)

	return nil
}

func (a *Sufr) urlEditHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	url, err := data.GetURL(id)
	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return renderTemplate(w, "404", nil)
		}

		return errors.Wrap(err, "failed to get url")
	}

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["URL"] = url
	ctx["Title"] = fmt.Sprintf("Editing %s", url.URL)

	return renderTemplate(w, "url-edit", ctx)
}

func (a *Sufr) urlSaveHandler(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	updateURLOptions := data.UpdateURLOptions{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(&updateURLOptions, r.PostForm); err != nil {
		return err
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	updateURLOptions.ID = id

	url, err := data.UpdateURL(updateURLOptions)
	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return renderTemplate(w, "404", nil)
		}

		return errors.Wrap(err, "failed to get url")
	}

	http.Redirect(w, r, reverse("url-view", "id", url.ID), http.StatusSeeOther)
	return nil
}

func (a *Sufr) urlDeleteHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	url, err := data.GetURL(id)
	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return renderTemplate(w, "404", nil)
		}

		return errors.Wrap(err, "failed to get url")
	}

	err = data.DeleteURL(url)
	if err != nil {
		return errors.Wrap(err, "failed to delete url")
	}

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func (a *Sufr) tagViewHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	tag, err := data.GetTag(id)
	if err != nil {
		return err
	}

	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["URLs"] = tag.URLs
	ctx["Count"] = len(tag.URLs)
	ctx["Title"] = tag.Name
	ctx["IsTagView"] = true
	ctx["Tag"] = tag

	return renderTemplate(w, "url-index", ctx)
}

func (a *Sufr) registrationHandler(w http.ResponseWriter, r *http.Request) error {
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

	opts := data.InitializeInstanceOptions{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(&opts, r.PostForm); err != nil {
		return err
	}

	formErrors := []string{}

	if !govalidator.IsEmail(opts.Email) {
		if opts.Email == "" {
			formErrors = append(formErrors, "Email cannot be blank")
		} else {
			formErrors = append(formErrors, fmt.Sprintf("%s is not a valid email address", opts.Email))
		}
	}
	if opts.Password == "" {
		formErrors = append(formErrors, "Password cannot be blank")
	}
	if len(formErrors) > 0 {
		for _, msg := range formErrors {
			session.AddFlash(msg, "danger")
		}
		session.Save(r, w)
		http.Redirect(w, r, reverse("config"), http.StatusSeeOther)
	}

	passwordCrypt, err := bcrypt.GenerateFromPassword([]byte(opts.Password), 0)
	if err != nil {
		return err
	}
	opts.Password = string(passwordCrypt)

	userOpts := data.UserOptions{
		Email:    opts.Email,
		Password: opts.Password,
	}

	user, err := data.CreateUser(userOpts)
	if err != nil {
		return err
	}

	settingsOpts := data.SettingsOptions{
		Private:     opts.Private,
		EmbedVideos: opts.EmbedVideos,
		EmbedPhotos: opts.EmbedPhotos,
		PerPage:     opts.PerPage,
	}

	_, err = data.SaveSettings(settingsOpts)
	if err != nil {
		return err
	}

	authsession, err := store.Get(r, "auth")
	authsession.Values["userID"] = user.ID
	authsession.Save(r, w)

	// Otherwise things are good
	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func (a *Sufr) loginHandler(w http.ResponseWriter, r *http.Request) error {
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

	opts := data.UserOptions{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(&opts, r.PostForm); err != nil {
		return err
	}

	formErrors := []string{}

	if !govalidator.IsEmail(opts.Email) {
		if opts.Email == "" {
			formErrors = append(formErrors, "Email cannot be blank")
		} else {
			formErrors = append(formErrors, fmt.Sprintf("%s is not a valid email address", opts.Email))
		}
	}

	if opts.Password == "" {
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

	user, err := data.GetUser()
	if err != nil {
		return err
	}

	if user.Email != opts.Email {
		formErrors = append(formErrors, "Email and password did not match")
	} else if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(opts.Password)); err != nil {
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
	err = authsession.Save(r, w)
	if err != nil {
		return errors.Wrap(err, "failed to save auth session")
	}

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func (a *Sufr) logoutHandler(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "auth")
	if err != nil {
		return err
	}

	delete(session.Values, "userID")
	session.Save(r, w)

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)

	return nil
}

func (a *Sufr) settingsHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Settings"

	settings, err := data.GetSettings()
	if err != nil {
		return err
	}

	if r.Method == "GET" {
		ctx["SettingsObject"] = settings
		return renderTemplate(w, "settings", ctx)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	opts := data.SettingsOptions{}
	decoder := schema.NewDecoder()

	if err := decoder.Decode(&opts, r.PostForm); err != nil {
		return err
	}

	session, err := store.Get(r, "flashes")
	if err != nil {
		return err
	}

	settings, err = data.SaveSettings(opts)
	if err != nil {
		session.AddFlash("There was an error saving your settings", "danger")
		session.Save(r, w)

		http.Redirect(w, r, reverse("settings"), http.StatusSeeOther)

		return nil
	}

	session.AddFlash("Settings have been saved", "success")
	session.Save(r, w)

	http.Redirect(w, r, reverse("settings"), http.StatusSeeOther)

	return nil
}
