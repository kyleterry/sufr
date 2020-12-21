package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kyleterry/sufr/pkg/store"
	"github.com/kyleterry/sufr/pkg/ui"
	"github.com/shurcooL/httpfs/html/vfstemplate"
)

const userAuthSessionKey = "user-auth"

type uiServer struct {
	db           store.Manager
	router       *http.ServeMux
	uifs         http.FileSystem
	sessionStore sessions.Store
	templates    *templates
}

func (s *uiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *uiServer) route() {
	auth := NewSessionAuthenticationMiddleware(s.sessionStore, s.db)

	s.router.HandleFunc("/", s.handleRootRedirect())
	s.router.Handle("/timeline", auth(s.handleTimeline()))
	s.router.Handle("/url", auth(s.handleURL()))
	s.router.Handle("/login", s.handleLogin())
	s.router.Handle("/logout", s.handleLogout())
	s.router.Handle("/static/", s.handleStatic())
}

func (s *uiServer) setupTemplates() {
	tm := make(map[string]*template.Template)

	f := template.FuncMap{
		"dict":            dict,
		"formatTimestamp": formatTimestamp,
		"tagNames":        tagNames,
		"reverse":         func(name string, p ...interface{}) string { return "" },
		"isyoutube":       func(name string, p ...interface{}) string { return "" },
		"youtubevid":      func(name string, p ...interface{}) string { return "" },
		"updatePage":      func(name string, p ...interface{}) string { return "" },
	}

	tm["timeline/index"] = template.Must(
		vfstemplate.ParseFiles(s.uifs, template.New("base").Funcs(f),
			"templates/base.html", "templates/url-index.html"))
	tm["urls/new"] = template.Must(
		vfstemplate.ParseFiles(s.uifs, template.New("base").Funcs(f),
			"templates/base.html", "templates/url-new.html"))
	tm["users/login"] = template.Must(
		vfstemplate.ParseFiles(s.uifs, template.New("base").Funcs(f),
			"templates/base.html", "templates/login.html"))
	tm["errors/404"] = template.Must(
		vfstemplate.ParseFiles(s.uifs, template.New("base").Funcs(f),
			"templates/base.html", "templates/404.html"))

	s.templates = &templates{templates: tm}
}

func (s *uiServer) handleRootRedirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
	}
}

func (s *uiServer) handleTimeline() http.HandlerFunc {
	srv := timelineServer{
		db:        s.db,
		router:    http.NewServeMux(),
		uifs:      s.uifs,
		templates: s.templates,
	}

	srv.route()

	return func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)
	}
}

func (s *uiServer) handleURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	}
}

func (s *uiServer) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodGet:
			err := s.templates.withWriter("users/login", func(tw *templateWriter) error {
				return tw.write(w, r, templateData{Title: "login"})
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			return
		case http.MethodPost:
			email := r.PostFormValue("email")
			password := r.PostFormValue("password")

			user, err := s.db.Users().GetByEmailAndPassword(ctx, email, password)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)

				return
			}

			session, err := s.sessionStore.Get(r, userAuthSessionKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			session.Values["userID"] = user.Id

			if err := session.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		default:
			http.NotFound(w, r)

			return
		}
	}
}

func (s *uiServer) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, userAuthSessionKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		delete(session.Values, "userID")

		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *uiServer) handleStatic() http.Handler {
	return http.FileServer(s.uifs)
}

func newUIServer(db store.Manager) *uiServer {
	s := &uiServer{
		db:           db,
		router:       http.NewServeMux(),
		uifs:         ui.NewFileSystem(),
		sessionStore: sessions.NewCookieStore(),
	}

	s.route()
	s.setupTemplates()

	return s
}
