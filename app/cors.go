package app

import "github.com/go-chi/cors"

func corsConfig() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // to be set to something more specific
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Content-Encoding"},
		ExposedHeaders:   []string{"Date", "Backend-Version"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}
