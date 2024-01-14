package main

import (
	"net/http"

	raptor "github.com/anthdm/raptor/sdk"
	"github.com/go-chi/chi"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello from the login handler YADA"))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello from the dashboard handler"))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("login page: <a href=\"/login\" /><br />Dashboard page: <a href=\"/dashboard\" />"))
}

func main() {
	router := chi.NewMux()
	router.Get("/dashboard", handleDashboard)
	router.Get("/login", handleLogin)
	router.Get("/", handleIndex)
	raptor.Handle(router)
}
