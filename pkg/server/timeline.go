package server

import (
	"net/http"
	"strconv"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/store"
)

type timelineServer struct {
	db        store.Manager
	router    *http.ServeMux
	uifs      http.FileSystem
	templates *templates
}

func (s *timelineServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *timelineServer) route() {
	s.router.Handle("/", s.handleTimeline())
}

func (s *timelineServer) handleTimeline() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := ctx.Value(userContextKey{}).(*api.User)
		after := r.URL.Query().Get("after")

		a, err := strconv.ParseInt(after, 10, 64)
		if err != nil {
			a = 0
		}

		err = s.templates.withWriter("timeline/index", func(tw *templateWriter) error {
			all, err := s.db.UserURLs(user).GetAll(ctx, store.WithResultsAfter(a))
			if err != nil {
				return err
			}

			td := timelineData{
				templateData: templateData{
					User:  user,
					Title: "timeline",
				},
				URLs:  all,
				Count: len(all),
			}

			return tw.write(w, r, td)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}
