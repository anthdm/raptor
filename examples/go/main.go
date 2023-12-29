package main

import (
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello from my handler"))
}

func main() {
	ffaas.HandleFunc(myHandler)
}
