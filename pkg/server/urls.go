package server

import (
	"net/http"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/store"
)

type urlServer struct {
	db        store.Manager
	router    *http.ServeMux
	uifs      http.FileSystem
	templates *templates
}

func (s *urlServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *urlServer) route() {
	s.router.HandleFunc("/url/new", s.handleURLNew())
}

func (s *urlServer) handleURLNew() http.HandlerFunc {
	// decoder := schema.NewDecoder()

	type urlForm struct {
		URL  string
		Tags []string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value(userContextKey{}).(*api.User)

		switch r.Method {
		case http.MethodGet:
			err := s.templates.withWriter("urls/new", func(tw *templateWriter) error {
				td := templateData{
					User:  user,
					Title: "New URL",
				}

				return tw.write(w, r, td)
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			// form := urlForm{}

			// if err := decoder.Decode(&form, r.PostForm); err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// }

			// if _, err := url.Parse(form.URL); err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// }

			// tl := api.TagList{}

			// for _, tag := range form.Tags {
			// 	tl.Items = append(tl.Items, *api.Tag{Name: tag})
			// }

			// if err := s.db.Tags().CreateAll(ctx, &tl); err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// }

			// uu := api.UserURL{}

			// if err := s.db.UserURLs(user).Create(ctx, &uu); err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// }
		default:
			http.NotFound(w, r)
		}
	}
}
