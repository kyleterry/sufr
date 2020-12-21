package server

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/kyleterry/sufr/pkg/store"
	"github.com/kyleterry/sufr/pkg/ui"
)

const defaultBindAddr = "127.0.0.1:8090"

type serverOptions struct {
	db             store.Manager
	bindAddr       string
	sessionAuthKey []byte
	sessionEncKey  []byte
}

type serverOptionFunc struct {
	f func(*serverOptions)
}

func (s *serverOptionFunc) apply(opts *serverOptions) {
	s.f(opts)
}

type ServerOption interface {
	apply(*serverOptions)
}

func WithStore(db store.Manager) ServerOption {
	return &serverOptionFunc{
		f: func(opts *serverOptions) {
			opts.db = db
		},
	}
}

func WithBindAddr(addr string) ServerOption {
	return &serverOptionFunc{
		f: func(opts *serverOptions) {
			opts.bindAddr = addr
		},
	}
}

func WithSessionKeyPair(auth, enc []byte) ServerOption {
	return &serverOptionFunc{
		f: func(opts *serverOptions) {
			opts.sessionAuthKey = auth
			opts.sessionEncKey = enc
		},
	}
}

type server struct {
	db     store.Manager
	router *http.ServeMux

	bindAddr       string
	sessionAuthKey []byte
	sessionEncKey  []byte
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) Run(ctx context.Context) error {
	return listenAndServe(ctx, s)
}

func (s *server) route() {
	s.router.HandleFunc("/", s.handleUI())
	s.router.HandleFunc("/api/", s.handleAPI())
}

func (s *server) handleUI() http.HandlerFunc {
	srv := uiServer{
		db:     s.db,
		router: http.NewServeMux(),
		uifs:   ui.NewFileSystem(),
		sessionStore: sessions.NewCookieStore(
			s.sessionAuthKey,
			s.sessionEncKey,
		),
	}

	srv.setupTemplates()
	srv.route()

	return func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)
	}
}

func (s *server) handleAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
}

func New(opts ...ServerOption) *server {
	so := serverOptions{
		bindAddr: defaultBindAddr,
	}

	for _, opt := range opts {
		opt.apply(&so)
	}

	srv := &server{
		db:             so.db,
		bindAddr:       so.bindAddr,
		router:         http.NewServeMux(),
		sessionAuthKey: so.sessionAuthKey,
		sessionEncKey:  so.sessionEncKey,
	}

	srv.route()

	return srv
}
