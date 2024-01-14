package main

import (
	"log"
	"net/http"

	raptor "github.com/anthdm/raptor/sdk"
	"github.com/go-chi/chi/v5"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("hello from the login handler YADA"))
	if err != nil {
		checkError(w, "handleLogin")
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("hello from the dashboard handler"))
	if err != nil {
		checkError(w, "handleDashboard")
	}
}

func main() {
	router := chi.NewMux()
	router.Get("/dashboard", handleDashboard)
	router.Get("/login", handleLogin)
	raptor.Handle(router)
}

func checkError(w http.ResponseWriter, handlerName string) {
	http.Error(w, "Failed to write response", http.StatusInternalServerError)
	log.Printf("Error in %s\n", handlerName)
	return
}
