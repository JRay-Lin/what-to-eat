package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"what-to-eat/pkg/api"
	"what-to-eat/pkg/vertex"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		type Response struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		}

		response := Response{
			Message: "Service is healthy",
			Status:  http.StatusOK,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Restaurant picker
		r.Route("/picker", func(r chi.Router) {
			r.Get("/random", api.RandomRestaurantHandler)
			r.Route("/ai", func(r chi.Router) {
				r.Post("/filter-categories", vertex.FilteredCategories)
				r.Post("/suggestion", vertex.RestaurantSuggestion)
			})
		})

		// Restaurant routes
		r.Route("/restaurants", func(r chi.Router) {
			r.Get("/menu", api.GetMenuHandler)
		})

		// Cuisines route
		r.Get("/cuisines", api.GetCuisinesHandler)
	})

	fmt.Println("Server running on port 3000")
	http.ListenAndServe(":3000", r)
}
