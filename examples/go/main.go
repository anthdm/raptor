package main

import (
	"fmt"
	"net/http"

	run "github.com/anthdm/run/sdk"
	"github.com/go-chi/chi/v5"
	_ "github.com/stealthrocket/net/http"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("on the login page"))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello from the dashboard"))
}

func main() {
	fmt.Println("this worked")
	router := chi.NewMux()
	router.Get("/dashboard", handleDashboard)
	router.Get("/login", handleLogin)
	run.Handle(router)
}
