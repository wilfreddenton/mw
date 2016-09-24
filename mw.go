package mw

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func isInInputs(inputs []string, input string) bool {
	for _, i := range inputs {
		if i == input {
			return true
		}
	}
	return false
}

func isValidHandler(inputs []string, h Handler) bool {
	for _, i := range h.Inputs {
		if !isInInputs(inputs, i) {
			log.Printf("%s ware is not receiving input %s", h.Name, i)
			return false
		}
	}
	return true
}

func isValidWare(inputs []string, w Ware) bool {
	for _, i := range w.Inputs {
		if !isInInputs(inputs, i) {
			log.Printf("%s ware is not receiving input %s", w.Name, i)
			return false
		}
	}
	return true
}

func Compose(ms []Ware, h Handler, inputs []string) http.Handler {
	if len(ms) == 0 {
		if !isValidHandler(inputs, h) {
			log.Fatalf("exiting because %s handler is not chained the correct blocks", h.Name)
		}
		return h.Fn
	} else {
		w := ms[0]
		if !isValidWare(inputs, w) {
			log.Fatalf("exiting because %s ware is not chained the correct blocks", w.Name)
		}
		return w.Fn(Compose(ms[1:], h, append(inputs, w.Outputs...)))
	}
}

func CreateEndpoints(r *mux.Router, es []Endpoint, prefix string) {
	for _, e := range es {
		r.Handle(prefix+e.Path, Compose(append(e.Middlewares, e.Blockwares...), e.Handler, []string{})).Methods(e.Methods...)
	}
}

type Ware struct {
	Name    string
	Inputs  []string
	Outputs []string
	Fn      func(http.Handler) http.Handler
}

type Handler struct {
	Name   string
	Inputs []string
	Fn     http.Handler
}

type Endpoint struct {
	Path        string
	Middlewares []Ware
	Blockwares  []Ware
	Handler     Handler
	Methods     []string
}
