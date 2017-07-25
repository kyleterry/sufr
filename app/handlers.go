package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/data"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var staticHandler = http.FileServer(
	&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo},
)

func page(r *http.Request) int {
	pagestr := r.URL.Query().Get("page")
	if pagestr == "" {
		pagestr = "1"
	}

	page, err := strconv.ParseInt(pagestr, 10, 64)
	if err != nil {
		return 0
	}

	return int(page)
}

func perPage(r *http.Request) int {
	settings, ok := r.Context().Value(settingsKey).(map[string]interface{})
	if !ok {
		return config.DefaultPerPage
	}

	per, ok := settings["PerPage"].(int)
	if !ok {
		return config.DefaultPerPage
	}

	return per
}

func (a *Sufr) urlIndexHandler(w http.ResponseWriter, r *http.Request) error {
	paginator, err := data.NewURLPaginator(page(r), perPage(r), 3, data.AllURLGetter{})
	if err != nil {
		return errors.Wrap(err, "failed to get paginator")
	}

	ctx := r.Context()

	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["Count"] = len(paginator.URLs)
	templateData["URLs"] = paginator.URLs
	templateData["Paginator"] = paginator

	ctx = context.WithValue(ctx, templateDataKey, templateData)

	return renderTemplate(w, r.WithContext(ctx), "url-index")
}

func (a *Sufr) urlFavoritesHandler(w http.ResponseWriter, r *http.Request) error {
	paginator, err := data.NewURLPaginator(page(r), perPage(r), 3, data.FavURLGetter{})
	if err != nil {
		return errors.Wrap(err, "failed to get paginator")
	}

	ctx := r.Context()

	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["Count"] = len(paginator.URLs)
	templateData["URLs"] = paginator.URLs
	templateData["Paginator"] = paginator
	templateData["Title"] = "Favorites"

	ctx = context.WithValue(ctx, templateDataKey, templateData)

	return renderTemplate(w, r.WithContext(ctx), "url-index")
}

func (a *Sufr) urlNewHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["Title"] = "Add a URL"
	ctx = context.WithValue(ctx, templateDataKey, templateData)

	return renderTemplate(w, r.WithContext(ctx), "url-new")
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

	session, err := store.Get(r, "flashes")
	if err != nil {
		return err
	}

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
	url, err := data.CreateURL(createURLOptions, data.HTTPMetadataFetcher{})
	if err != nil {
		if errors.Cause(err) == data.ErrDuplicateKey {
			session.AddFlash("URL already exists", "danger")
			if err := session.Save(r, w); err != nil {
				return err
			}
			http.Redirect(w, r, reverse("url-new"), http.StatusSeeOther)
			return nil
		}
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
			return renderTemplate(w, r, "404")
		}

		return errors.Wrap(err, "failed to get url")
	}

	if !loggedIn(r) && url.Private {
		w.WriteHeader(404)
		return renderTemplate(w, r, "404")
	}

	ctx := r.Context()
	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["URL"] = url

	ctx = context.WithValue(ctx, templateDataKey, templateData)

	return renderTemplate(w, r.WithContext(ctx), "url-view")
}

func (a *Sufr) urlToggleFavoriteHandler(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	id, err := uuid.Parse(vars["id"])
	if err != nil {
		return err
	}

	url, err := data.GetURL(id)

	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return renderTemplate(w, r, "404")
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
			return renderTemplate(w, r, "404")
		}

		return errors.Wrap(err, "failed to get url")
	}

	ctx := r.Context()
	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["URL"] = url
	templateData["Title"] = fmt.Sprintf("Editing %s", url.URL)
	ctx = context.WithValue(ctx, templateDataKey, templateData)

	return renderTemplate(w, r.WithContext(ctx), "url-edit")
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
			return renderTemplate(w, r, "404")
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
			return renderTemplate(w, r, "404")
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

	paginator, err := data.NewURLPaginator(page(r), perPage(r), 3, tag)
	if err != nil {
		return errors.Wrap(err, "failed to get paginator")
	}

	ctx := r.Context()

	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["URLs"] = paginator.URLs
	templateData["Count"] = len(paginator.URLs)
	templateData["Paginator"] = paginator
	templateData["Title"] = tag.Name
	templateData["IsTagView"] = true
	templateData["Tag"] = tag

	ctx = context.WithValue(ctx, templateDataKey, templateData)

	return renderTemplate(w, r.WithContext(ctx), "url-index")
}

func (a *Sufr) PinTagHandler(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (a *Sufr) registrationHandler(w http.ResponseWriter, r *http.Request) error {
	if applicationConfigured() {
		http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	}

	ctx := r.Context()
	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["Title"] = "Setup"
	ctx = context.WithValue(ctx, templateDataKey, templateData)

	if r.Method == "GET" {
		return renderTemplate(w, r.WithContext(ctx), "registration")
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
	if err != nil {
		return err
	}

	id, _ := user.ID.MarshalText()
	authsession.Values["userID"] = id

	err = authsession.Save(r, w)
	if err != nil {
		return errors.Wrap(err, "failed to save auth session")
	}

	// Otherwise things are good
	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)
	return nil
}

func (a *Sufr) loginHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	templateData := ctx.Value(templateDataKey).(map[string]interface{})
	templateData["Title"] = "Login"

	ctx = context.WithValue(ctx, templateDataKey, templateData)

	if r.Method == "GET" {
		return renderTemplate(w, r.WithContext(ctx), "login")
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

	id, _ := user.ID.MarshalText()
	authsession.Values["userID"] = id

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

	err = session.Save(r, w)
	if err != nil {
		return errors.Wrap(err, "failed to save auth session")
	}

	http.Redirect(w, r, reverse("url-index"), http.StatusSeeOther)

	return nil
}

func (a *Sufr) settingsHandler(w http.ResponseWriter, r *http.Request) error {
	settings, err := data.GetSettings()
	if err != nil {
		return err
	}

	if r.Method == "GET" {
		ctx := r.Context()
		templateData := ctx.Value(templateDataKey).(map[string]interface{})
		templateData["Title"] = "Settings"
		templateData["SettingsObject"] = settings
		ctx = context.WithValue(ctx, templateDataKey, templateData)
		return renderTemplate(w, r.WithContext(ctx), "settings")
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
