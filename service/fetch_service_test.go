package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"log/slog"
	"encoding/json"
	"io"

	"github.com/nuwanwimalasooriya/go-wa-api/models"
	"github.com/nuwanwimalasooriya/go-wa-api/service"
)


type MockFetcher struct{}

func (m *MockFetcher) ContentFetch(ctx context.Context, url string) (string, error) {
	return `<html><title>Test Page</title><h1>Hello World</h1><a href="https://example.com">Link</a></html>`, nil
}

func TestHandleFetchGet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))

	mockFetcher := &MockFetcher{}
	htmlAnalyzer := service.NewHTMLAnalyzer(logger)
	fetchSvc := service.NewFetchService(mockFetcher, htmlAnalyzer, logger)


	req := httptest.NewRequest(http.MethodGet, "/fetch?url=http://example.com", nil)
	w := httptest.NewRecorder()

	
	fetchSvc.HandleFetchGet(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Decode JSON response
	var fetchResp models.FetchResponse
	err := json.NewDecoder(resp.Body).Decode(&fetchResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check Assertions
	if fetchResp.Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", fetchResp.Title)
	}
	if fetchResp.Headings[0].Text != "Hello World" {
    t.Errorf("expected heading text 'Hello World', got %v", fetchResp.Headings[0].Text)
	}
	if fetchResp.Headings[0].Level != "h1" {
    t.Errorf("expected heading tag 'h1', got %v", fetchResp.Headings[0].Level)
}
	if len(fetchResp.Links) != 1 || fetchResp.Links[0] != "https://example.com" {
		t.Errorf("Expected link 'https://example.com', got %v", fetchResp.Links)
	}
	if fetchResp.LoginDetected {
		t.Errorf("Expected LoginDetected=false, got true")
	}
}
