package main

import (
	"fmt"
	"math/rand"
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	num := rand.Intn(100)
	w.Write([]byte(fmt.Sprintf("from my application: %d", num)))
}

func main() {
	ffaas.HandleFunc(myHandler)
}
