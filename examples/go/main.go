package main

import (
	"net/http"

	run "github.com/anthdm/run/sdk"
	"github.com/go-chi/chi/v5"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("on the login page"))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello from the dashboard"))
}

func main() {
	router := chi.NewMux()
	router.Get("/login", handleLogin)
	router.Get("/dashboard", handleDashboard)
	run.Handle(router)
}
