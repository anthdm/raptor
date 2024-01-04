package main

import (
	"net/http"

	run "github.com/anthdm/run/sdk"
)

func main() {
	run.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good!"))
	}))
}
