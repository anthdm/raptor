package main

import (
	"fmt"
	"net/http"

	raptor "github.com/anthdm/raptor/sdk"
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
	// router := chi.NewMux()
	// router.Get("/dashboard", handleDashboard)
	// router.Get("/login", handleLogin)
	fmt.Println("user log")
	raptor.Handle(http.HandlerFunc(handleLogin))
}
