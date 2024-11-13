package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func main() {
	// Initialize Chi router
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Message: "Welcome to the API",
			Status:  http.StatusOK,
		}
		json.NewEncoder(w).Encode(response)
	})

	// RESTful routes example
	r.Route("/api", func(r chi.Router) {
		// GET /api/health
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			response := Response{
				Message: "Service is healthy",
				Status:  http.StatusOK,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})

	})

	// Start server
	http.ListenAndServe(":3000", r)
}
