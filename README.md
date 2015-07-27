# ehttp

[![GoDoc](https://godoc.org/github.com/creack/ehttp?status.svg)](https://godoc.org/github.com/creack/ehttp) [![Build Status](https://travis-ci.org/creack/ehttp.svg?branch=master)](https://travis-ci.org/creack/ehttp) [![Coverage Status](https://coveralls.io/repos/creack/ehttp/badge.svg?branch=master&service=github)](https://coveralls.io/github/creack/ehttp?branch=master)

This package allows you to write http handlers returning an error.

## HTTP Status Code

`ehttp.NewError` and `ehttp.NewErrorf` can be called to create an error with a custom http status.

If the error is not nil and not an `ehttp.Error`, then 500 InternalServerError is sent.

If the status is 0, it implies 500.

The same idea applies to panic as well as returned errors.

## Error after sending headers

Due to http limitation, we can send the headers only once. If some data has been sent prior to
the error, then nothing gets send to the client, the error gets logged on the server side.

## Panic

The default `ehttp.MWError` handles errors, but do not handle panics.
In order to send the panic error to the user (or log it after headers are sent), you can use `ehttp.MWErrorPanic`
which wraps `ehttp.MWError` and use the recovered value as an error.

If the panic value is an `ehttp.Error`, the proper http status code will be sent to the client when possible.

## Support

The package have been tested with:

- [net/http](http://godoc.org/net/http)
- [github.com/gorilla/mux](http://www.gorillatoolkit.org/pkg/mux)
- [github.com/julienschmidt/httprouter](http://godoc.org/github.com/julienschmidt/httprouter)

# Examples

## net/http

```go
package main

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
)

func hdlr(w http.ResponseWriter, req *http.Request) error {
	return ehttp.NewErrorf(418, "fail")
}

func main() {
	http.HandleFunc("/", ehttp.MWError(hdlr))
	http.Handle("/", ehttp.HandlerFunc(hdlr))
	ehttp.HandleFunc("/", hdlr)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## gorilla/mux

```go
package main

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
	"github.com/gorilla/mux"
)

func hdlr(w http.ResponseWriter, req *http.Request) error {
	return ehttp.NewErrorf(418, "fail")
}

func main() {
	router := mux.NewRouter()
	router.Handle("/", ehttp.HandlerFunc(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

## httprouter

```go
package main

import (
	"log"
	"net/http"

	"github.com/creack/ehttp"
	"github.com/creack/ehttp/ehttprouter"
	"github.com/julienschmidt/httprouter"
)

func hdlr(w http.ResponseWriter, req *http.Request, p httprouter.Params) error {
	return ehttp.NewErrorf(418, "fail")
}

func main() {
	router := httprouter.New()
	router.GET("/", ehttprouter.MWError(hdlr))
	log.Fatal(http.ListenAndServe(":8080", router))
}
```
