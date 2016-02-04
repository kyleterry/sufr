package app

import (
	"log"
	"net/http"
)

type MiddlewareHandler func(http.Handler) http.Handler

type MiddlewareChain struct {
	Handlers []MiddlewareHandler
}

func NewMiddlewareChain(handlers ...MiddlewareHandler) MiddlewareChain {
	chain := MiddlewareChain{}
	chain.Handlers = append(chain.Handlers, handlers...)
	return chain
}

func (mc MiddlewareChain) SetHandler(h http.Handler) http.Handler {
	//Since it's a stack we need to go --
	for i := len(mc.Handlers) - 1; i >= 0; i-- {
		h = mc.Handlers[i](h)
	}

	return h
}

func (mc MiddlewareChain) SetHandlerFunc(fn http.HandlerFunc) http.Handler {
	return mc.SetHandler(http.HandlerFunc(fn))
}

func TestM1(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("entering m1")
		h.ServeHTTP(w, r)
		log.Println("exiting m1")
	})
}

func TestM2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("entering m2")
		h.ServeHTTP(w, r)
		log.Println("exiting m2")
	})
}
