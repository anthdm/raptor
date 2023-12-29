package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

var (
	ErrDecodeRequestBody = errors.New("could not decode the request body")
)

type errorResponse struct {
	Error string `json:"error"`
}

func ErrorResponse(err error) errorResponse {
	return errorResponse{
		Error: err.Error(),
	}
}

type apiHandler func(w http.ResponseWriter, r *http.Request) error

func makeAPIHandler(h apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			// todo
			slog.Error("api handler error", "err", err)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
