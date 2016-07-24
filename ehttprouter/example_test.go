package ehttprouter_test

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
	"github.com/creack/ehttp/ehttprouter"
	"github.com/julienschmidt/httprouter"
)

func Example_simple() {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		return ehttp.NewErrorf(http.StatusTeapot, "fail")
	}
	router := httprouter.New()
	router.GET("/", ehttprouter.MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Example_panic() {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(ehttp.NewErrorf(http.StatusTeapot, "fail"))
	}
	router := httprouter.New()
	router.GET("/", ehttprouter.MWErrorPanic(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}
