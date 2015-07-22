package ehttprouter

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
	"github.com/julienschmidt/httprouter"
)

func Example_simple() {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		return ehttp.NewErrorf(418, "fail")
	}
	router := httprouter.New()
	router.GET("/", MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Example_panic() {
	hdlr := func(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
		panic(ehttp.NewErrorf(418, "fail"))
	}
	router := httprouter.New()
	router.GET("/", MWErrorPanic(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}
