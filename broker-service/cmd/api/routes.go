package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "ContentType", "X-CSRF-Token", "Referrer-Policy", "X-Content-Type-Options", "Strict-Transport-Security"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Post("/", app.Broker)
	mux.Get("/health", app.Health)
	mux.Post("/handle", app.HandleSubmission)
	mux.Post("/log-grpc", app.GRPCLog)
	return mux
}
