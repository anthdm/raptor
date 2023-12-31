package cors

import (
	"net/http"
)

// Cors configuration.
type CorsConfig struct {
	Api struct {
		Origin         string
		AllowedMethods string
		AllowedHeaders string
	}
	Wasm struct {
		Origin         string
		AllowedMethods string
		AllowedHeaders string
	}
}

// Cors holds the CORS configuration.
type Cors struct {
	origin         string
	methods        string
	allowedHeaders string
}

// NewCors returns a new Cors struct.
func NewCors(origin string, methods string, allowedHeaders string) *Cors {
	return &Cors{
		origin:         origin,
		methods:        methods,
		allowedHeaders: allowedHeaders,
	}
}

// ApplyCORS applies CORS to the given handler.
func (c *Cors) ApplyCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", c.origin)
		w.Header().Set("Access-Control-Allow-Methods", c.allowedHeaders)
		w.Header().Set("Access-Control-Allow-Headers", c.methods)
		next.ServeHTTP(w, r)
	})
}
