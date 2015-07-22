package ehttp

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Example_http() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return NewErrorf(418, "fail")
	}
	http.HandleFunc("/", MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_gorilla() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return NewErrorf(418, "fail")
	}
	router := mux.NewRouter()
	router.HandleFunc("/", MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Example_httpPanic() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		panic(NewErrorf(418, "big fail"))
	}
	http.HandleFunc("/", MWErrorPanic(hdlr))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_gorillaPanic() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		panic(NewErrorf(418, "big fail"))
	}
	router := mux.NewRouter()
	router.HandleFunc("/", MWErrorPanic(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}
