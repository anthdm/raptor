package main

import (
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func main() {
	ffaas.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good!"))
	}))
}
