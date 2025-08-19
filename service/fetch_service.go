package service

import (
	//"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"regexp"
	"log/slog"

	models "github.com/nuwanwimalasooriya/go-wa-api/models"
)

type FetchService struct {
	fetcher  Fetcher
	analyzer Analyzer
	logger   *slog.Logger
}

func NewFetchService(fetcher Fetcher, analyzer Analyzer, logger *slog.Logger) *FetchService {
	return &FetchService{
		fetcher:  fetcher,
		analyzer: analyzer,
		logger:   logger,
	}
}

 // GET /fetch?url=
func (fs *FetchService) HandleFetchGet(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimSpace(r.URL.Query().Get("url"))
	if url == "" {
		fs.logger.Error("Missing URL parameter")
		http.Error(w, "url parameter required", http.StatusBadRequest)
		return
	}

	content, err := fs.fetcher.ContentFetch(r.Context(), url)
	if err != nil {
		fs.logger.Error("Fetch failed", "url", url, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := fs.analyzer.Analyze(content,url)
	writeJSON(w, resp, http.StatusOK)
}

// Helpers
func validateRequest(r *http.Request) (models.FetchRequest, error) {
	var req models.FetchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	var urlRegex = regexp.MustCompile(`^(https?:\/\/|www\.)[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(:[0-9]{1,5})?(\/.*)?$`)

	if err != nil || strings.TrimSpace(req.URL) == "" {
		return req, errors.New("invalid request payload")
	} else if !urlRegex.MatchString(req.URL) {
    return req, errors.New("Invalid URL")
	}

	return req, nil
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
