package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-106/internal/console"
)

func consoleRoutes() http.Handler {
	router := chi.NewRouter()
	for _, page := range console.Pages() {
		router.Handle(page.Path, console.Handler(page))
	}
	return router
}
