package mw

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	keyText  = "text"
	textVal1 = "Hello World!"
	textVal2 = "h e l l o w o r l d"
)

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
var printer = Ware{
	Name:   "printer",
	Inputs: []string{keyText},
	Fn: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if v, ok := context.GetOk(r, keyText); ok {
				fmt.Println(v)
			}
			next.ServeHTTP(w, r)
		})
	},
}

var writer1 = Ware{
	Name:    "writer1",
	Outputs: []string{keyText},
	Fn: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, keyText, textVal1)
			next.ServeHTTP(w, r)
		})
	},
}

var writer2 = Ware{
	Name:    "writer2",
	Outputs: []string{keyText},
	Fn: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, keyText, textVal2)
			next.ServeHTTP(w, r)
		})
	},
}

// handlers
var handler1 = Handler{
	Name:   "handler1",
	Inputs: []string{keyText},
	Fn: appHandler(func(w http.ResponseWriter, r *http.Request) *Error {
		t := context.Get(r, keyText).(string)
		w.Write([]byte(t))
		return nil
	}),
}

var handler2 = Handler{
	Name:   "handler2",
	Inputs: []string{keyText},
	Fn: appHandler(func(w http.ResponseWriter, r *http.Request) *Error {
		t := context.Get(r, keyText).(string)
		w.Write([]byte(t))
		return nil
	}),
}

func TestIsInInputs(t *testing.T) {
	i1 := []string{"hello", "world"}
	i2 := []string{"hello", "you", "there"}
	if !isInInputs(i1, "hello") {
		t.Errorf("should be in inputs")
	}
	if isInInputs(i2, "world") {
		t.Errorf("should not be in inputs")
	}
}

func TestIsValidHandler(t *testing.T) {
	inputs := []string{"hello", "world"}
	h := Handler{Name: "h1", Inputs: inputs}
	if !isValidHandler(inputs, h) {
		t.Error("should be a valid handler")
	}
	h = Handler{Name: "h2", Inputs: []string{"foo"}}
	if isValidHandler(inputs, h) {
		t.Error("should be an invalid handler")
	}
}

func TestIsValidWare(t *testing.T) {
	inputs := []string{"hello", "world"}
	w := Ware{Name: "w1", Inputs: inputs}
	if !isValidWare(inputs, w) {
		t.Error("should be a valid ware")
	}
	w = Ware{Name: "w2", Inputs: []string{"foo"}}
	if isValidWare(inputs, w) {
		t.Error("should be an invalid ware")
	}
}

func TestCompose(t *testing.T) {
	key1 := "key1"
	key2 := "key2"
	key3 := "key3"
	key4 := "key4"
	val1 := "val1"
	val2 := "val2"
	val3 := "val3"
	val4 := "val4"
	ws := []Ware{
		Ware{
			Name:    "w1",
			Outputs: []string{key1, key2},
			Fn: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					context.Set(r, key1, val1)
					context.Set(r, key2, val2)
					next.ServeHTTP(w, r)
				})
			},
		},
		Ware{
			Name:    "w2",
			Inputs:  []string{key1, key2},
			Outputs: []string{key3},
			Fn: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if _, ok := context.GetOk(r, key1); !ok {
						t.Errorf("%s not set", key1)
					}
					if _, ok := context.GetOk(r, key2); !ok {
						t.Errorf("%s not set", key2)
					}
					context.Set(r, key3, val3)
					next.ServeHTTP(w, r)
				})
			},
		},
		Ware{
			Name:    "w3",
			Inputs:  []string{key1, key2, key3},
			Outputs: []string{key4},
			Fn: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if _, ok := context.GetOk(r, key1); !ok {
						t.Errorf("%s not set", key1)
					}
					if _, ok := context.GetOk(r, key2); !ok {
						t.Errorf("%s not set", key2)
					}
					if _, ok := context.GetOk(r, key3); !ok {
						t.Errorf("%s not set", key3)
					}
					context.Set(r, key4, val4)
					next.ServeHTTP(w, r)
				})
			},
		},
	}
	h := Handler{
		Name:   "h1",
		Inputs: []string{key4},
		Fn: appHandler(func(w http.ResponseWriter, r *http.Request) *Error {
			if _, ok := context.GetOk(r, key4); !ok {
				t.Errorf("%s not set", key4)
			}
			w.Write([]byte(val4))
			return nil
		}),
	}
	chain := Compose(ws, h, []string{})
	path := "/testcompose"
	router := mux.NewRouter()
	router.Handle(path, chain)
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatal("could not create request")
	}
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("got code %v; want %v", w.Code, http.StatusOK)
	}
}

func TestCreateEndpoint(t *testing.T) {
	es := []Endpoint{
		Endpoint{
			"/e1",
			[]Ware{writer1},
			[]Ware{printer},
			handler1,
			[]string{"GET"},
		},
		Endpoint{
			"/e2",
			[]Ware{writer2},
			[]Ware{printer},
			handler2,
			[]string{"GET"},
		},
	}
	prefix := "/api/v1"
	router := mux.NewRouter()
	CreateEndpoints(router, es, prefix)
	// TEST e1
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, prefix+"/e1", nil)
	if err != nil {
		t.Fatal("could not create request")
	}
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("got code %v; want %v", w.Code, http.StatusOK)
	}
	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error("body could not be read")
	}
	val := string(body)
	if val != textVal1 {
		t.Errorf("got body %v; want %v", val, textVal1)
	}
	// TEST e2
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, prefix+"/e2", nil)
	if err != nil {
		t.Fatal("could not create request")
	}
	router.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("got code %v; want %v", w.Code, http.StatusOK)
	}
	body, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Error("body could not be read")
	}
	val = string(body)
	if val != textVal2 {
		t.Errorf("got body %v; want %v", val, textVal2)
	}
}
