package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/iamYole/go-movies/internal/env"
)

func (app *application) routes() http.Handler {
	// create a router mux
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{env.GetString("FRONTEND_URL", "http://localhost:3000")}, // Use this to allow specific origin hosts
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	mux.Get("/", app.Home)
	mux.Get("/movies", app.AllMovies)
	mux.Get("/authenticate", app.authenticate)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh",app.refreshToken)
	mux.Get("/logout",app.logout)

	mux.Post("/register", app.Register)

	mux.Route("/admin",func(r chi.Router) {
		r.Use(app.authRequired)

		r.Get("/movies", app.MovieCatalog)
	})

	return mux
}
