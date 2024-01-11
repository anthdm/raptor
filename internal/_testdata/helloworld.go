package main

import (
	"net/http"

	raptor "github.com/anthdm/raptor/sdk"
)

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world!"))
}

func main() {
	raptor.Handle(http.HandlerFunc(handle))
}
