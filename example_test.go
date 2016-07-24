package ehttp_test

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
	"github.com/gorilla/mux"
)

func Example_http() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return ehttp.NewErrorf(418, "fail")
	}
	http.HandleFunc("/", ehttp.MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_gorilla() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	}
	router := mux.NewRouter()
	router.HandleFunc("/", ehttp.MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Example_httpPanic() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		panic(ehttp.NewErrorf(http.StatusTeapot, "big fail"))
	}
	http.HandleFunc("/", ehttp.MWErrorPanic(hdlr))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_gorillaPanic() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		panic(ehttp.NewErrorf(http.StatusTeapot, "big fail"))
	}
	router := mux.NewRouter()
	router.HandleFunc("/", ehttp.MWErrorPanic(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}
