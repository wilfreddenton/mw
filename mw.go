package mw

import (
	"github.com/gorilla/mux"
	"net/http"
)

func compose(ms []Ware, h http.Handler) http.Handler {
	if len(ms) == 0 {
		return h
	} else {
		return ms[0](compose(ms[1:], h))
	}
}

func CreateEndpoints(r *mux.Router, es []Endpoint) {
	for _, e := range es {
		r.Handle(e.Path, compose(append(e.Middlewares, e.Blockwares...), e.Handler)).Methods(e.Methods...)
	}
}

type Ware func(http.Handler) http.Handler

type Endpoint struct {
	Path        string
	Middlewares []Ware
	Blockwares  []Ware
	Handler     appHandler
	Methods     []string
}
