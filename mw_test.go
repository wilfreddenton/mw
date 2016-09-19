package mw

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"testing"
)

const keyText = "text"

type Error struct {
	Message string
	Code    int
	Error   error
}

type appHandler func(w http.ResponseWriter, r *http.Request) *Error

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil {
		http.Error(w, e.Message, e.Code)
	}
}

// middlewares
func printer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v, ok := context.GetOk(r, "text"); ok {
			fmt.Println(v)
		}
		next.ServeHTTP(w, r)
	})
}

func writer1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, keyText, "Hello World!")
		next.ServeHTTP(w, r)
	})
}

func writer2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, keyText, "h e l l o w o r l d")
		next.ServeHTTP(w, r)
	})
}

// handlers
func endpoint1Handler(w http.ResponseWriter, r *http.Request) *Error {
	w.Write([]byte("endpoint 1"))
	return nil
}

func endpoint2Handler(w http.ResponseWriter, r *http.Request) *Error {
	w.Write([]byte("endpoint 2"))
	return nil
}

func Test(t *testing.T) {
	e1 := Endpoint{
		"/e1",
		[]Ware{writer1, printer},
		endpoint1Handler,
		[]string{"GET"},
	}
	e2 := Endpoint{
		"/e2",
		[]Ware{writer2, printer},
		endpoint2Handler,
		[]string{"GET"},
	}
	es := []Endpoint{e1, e2}
	r := mux.NewRouter().StrictSlash(true)
	CreateEndpoints(r, es)
	http.ListenAndServe(":9999", r)
}
