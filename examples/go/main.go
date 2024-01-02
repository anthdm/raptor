package main

import (
	"fmt"
	"math/rand"
	"net/http"

	ffaas "github.com/anthdm/ffaas/sdk"
	"github.com/go-chi/chi"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("x-request-id")
	w.Write([]byte(requestID))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	num := rand.Intn(100)
	fmt.Println(r.URL)
	w.Write([]byte(fmt.Sprintf("from /login => %d", num)))
}

func main() {
	router := chi.NewMux()
	router.Get("/", handleHome)
	router.Get("/login", handleLogin)
	ffaas.Handle(router)
}
