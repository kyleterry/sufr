package server

import (
	"context"
	"net/http"
)

func listenAndServe(ctx context.Context, s *server) error {
	hs := http.Server{Addr: s.bindAddr, Handler: s}

	return hs.ListenAndServe()
}
