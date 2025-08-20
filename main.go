package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	middlewarex "github.com/nuwanwimalasooriya/go-wa-api/middleware"
	service "github.com/nuwanwimalasooriya/go-wa-api/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{AddSource: true}))
	slog.SetDefault(logger)

	httpFetcher := service.NewContentPFetcher(15*time.Second, logger)
	htmlAnalyzer := service.NewHTMLAnalyzer(logger)
	fetchService := service.ContentFetchService(httpFetcher, htmlAnalyzer, logger)

	r := NewRouter(logger, fetchService)

	slog.Info("Content Fetching Server started", "addr", ":8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		slog.Error("Server failed", "err", err)
	}
}

func NewRouter(logger *slog.Logger, fetchService *service.FetchService) http.Handler {
	r := chi.NewRouter()
	middlewarex.Register(r)

	r.Route("/api", func(api chi.Router) {
		api.Get("/fetch", fetchService.HandleFetchGet)
	})

	return r
}
