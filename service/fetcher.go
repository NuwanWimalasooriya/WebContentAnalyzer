package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"log/slog"
)


type Fetcher interface {
	ContentFetch(ctx context.Context, url string) (string, error)
}


type ContentFetcher struct {
	client *http.Client
	logger *slog.Logger
}

func NewContentPFetcher(timeout time.Duration, logger *slog.Logger) *ContentFetcher {
	return &ContentFetcher{
		client: &http.Client{Timeout: timeout},
		logger: logger,
	}
}

func (hf *ContentFetcher) ContentFetch(ctx context.Context, url string) (string, error) {
	hf.logger.Info("Fetching URL", "url", url)
	start := time.Now()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	hf.logger.Info("HandleFetchGet execution time", "duration", time.Since(start))
	if err != nil {
		hf.logger.Error("Failed to create request", "url", url, "err", err)
		return "", fmt.Errorf("creating request failed: %w", err)
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
		
	res, err := hf.client.Do(request)
	if err != nil {
		hf.logger.Error("HTTP request failed", "url", url, "err", err)
		return "", fmt.Errorf("fetching url failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		hf.logger.Warn("Unexpected HTTP status", "url", url, "status", res.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		hf.logger.Error("Failed to read body", "url", url, "err", err)
		return "", fmt.Errorf("reading response failed: %w", err)
	}

	hf.logger.Info("Fetch successful", "url", url, "length", len(body))
	return string(body), nil
}
