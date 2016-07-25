package ehttp_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/creack/ehttp"
	"github.com/gorilla/mux"
)

func ExampleServeMux() {
	mux := ehttp.NewServeMux(nil, "application/text; charset=utf-8", false, nil)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		return ehttp.InternalServerError
	})
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func ExampleServeMux_panic() {
	mux := ehttp.NewServeMux(nil, "application/text; charset=utf-8", true, nil)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		panic(ehttp.InternalServerError)
	})
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func ExampleServeMux_customLogger() {
	// When returning an error after sending the errors, we can't set the http Status.
	// As data as already been sent, we don't want to corrupt it so we log the error server side.
	logger := log.New(os.Stderr, "", log.LstdFlags)
	mux := ehttp.NewServeMux(nil, "application/text; charset=utf-8", true, logger)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		panic(ehttp.InternalServerError)
	})
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func ExampleServeMux_customized() {
	// Define our error format and how to expose it to the client.
	type customError struct {
		Error    string `json:"error"`
		HTTPCode int    `json:"http_code"`
	}
	errorHandler := func(w ehttp.ResponseWriter, req *http.Request, err error) {
		_ = json.NewEncoder(w).Encode(customError{
			Error:    err.Error(),
			HTTPCode: w.Code(),
		})
	}

	// Define a cutom logger for unexpected events (double header send).
	logger := log.New(os.Stderr, "", log.LstdFlags)

	// Create the mux.
	mux := ehttp.NewServeMux(errorHandler, "application/text; charset=utf-8", false, logger)

	// Register the handler.
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		// Return an error.
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	})

	// Start serve the mux.
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func Example_http() {
	hdlr := func(w http.ResponseWriter, req *http.Request) error {
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	}
	http.HandleFunc("/", ehttp.MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_customErrorHandler() {
	// Define how to send errors to the user.
	errorHandler := func(w ehttp.ResponseWriter, req *http.Request, err error) {
		fmt.Fprintf(w, "<<<<<<%s>>>>>>", err)
	}
	mux := ehttp.NewServeMux(errorHandler, "application/text; charset=utf-8", false, nil)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) error {
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	})
	log.Fatal(http.ListenAndServe(":8080", mux))
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
