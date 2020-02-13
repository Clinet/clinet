package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type APIError struct {
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

func errAPI(err ...interface{}) *APIError {
	if len(err) > 1 {
		return &APIError{Error: err[0].(string), Details: fmt.Sprintf("%v", err[1].(error))}
	}
	return &APIError{Error: err[0].(string)}
}

func StartAPI(host string) {
	router := APIRouter()

	walkFunc := func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		DebugAPI.Printf("%s %s\n", method, route) //Walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		ErrorAPI.Printf("Error walking routes: %v", err)
		return
	}

	if err := http.ListenAndServe(host, router); err != nil {
		ErrorAPI.Printf("Error running HTTP server: %v", err)
	}
}

func APIRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), //Set Content-Type to application/json
		middleware.Logger,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	)

	router.Route("/api", func(r chi.Router) {
		r.Mount("/v0", APIv0())
	})

	return router
}
