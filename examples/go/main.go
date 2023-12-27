package main

import (
	"fmt"
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func init() {
	ffaas.Handle(myHandler)
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("this is my response"))
}

func main() {
	fmt.Println("hello")
}
