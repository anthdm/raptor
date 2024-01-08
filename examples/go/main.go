package main

import (
	"log"
	"net/http"

	run "github.com/anthdm/raptor/sdk"
	"github.com/go-chi/chi"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello from the login handler"))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello from the dashboard handler"))
}

func main() {
	_, err := http.Get("http://google.com")
	if err != nil {
		log.Fatal(err)
	}
	router := chi.NewMux()
	router.Get("/dashboard", handleDashboard)
	router.Get("/login", handleLogin)
	run.Handle(router)
}
