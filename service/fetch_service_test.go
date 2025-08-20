package service_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuwanwimalasooriya/go-wa-api/models"
	"github.com/nuwanwimalasooriya/go-wa-api/service"
)


type MockFetcher struct{
	content string 
	err error
}

func (m *MockFetcher) ContentFetch(ctx context.Context, url string) (string, error) {
	return m.content,m.err
}


func setupMockServiceWithContent(content string) *service.FetchService {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{AddSource: false}))
	mockFetcher := &MockFetcher{content: content}
	htmlAnalyzer := service.NewHTMLAnalyzer(logger)
	return service.ContentFetchService(mockFetcher, htmlAnalyzer, logger)
}

//  Test case for Title extraction
func TestHandleFetchGet_Title(t *testing.T) {
	htmlContent := `<html><title>My Test Title</title></html>`
	fetchService := setupMockServiceWithContent(htmlContent)

	request := httptest.NewRequest(http.MethodGet, "/fetch?url=http://example.com", nil)
	w := httptest.NewRecorder()
	fetchService.HandleFetchGet(w, request)

	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", response.StatusCode)
	}

	var fetchResponse models.FetchResponse
	if err := json.NewDecoder(response.Body).Decode(&fetchResponse); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if fetchResponse.Title != "My Test Title" {
		t.Errorf("Expected title 'My Test Title', got %s", fetchResponse.Title)
	}
}

//  Test case for Headings extraction
func TestHandleFetchGet_Headings(t *testing.T) {
	htmlContent := `<html><h1>Main Heading</h1><h2>Sub Heading</h2></html>`
	fetchService := setupMockServiceWithContent(htmlContent)

	request := httptest.NewRequest(http.MethodGet, "/fetch?url=http://example.com", nil)
	w := httptest.NewRecorder()
	fetchService.HandleFetchGet(w, request)

	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", response.StatusCode)
	}

	var fetchResponse models.FetchResponse
	if err := json.NewDecoder(response.Body).Decode(&fetchResponse); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(fetchResponse.Headings) != 2 {
		t.Fatalf("Expected 2 headings, got %d", len(fetchResponse.Headings))
	}
	if fetchResponse.Headings[0].Text != "Main Heading" || fetchResponse.Headings[0].Level != "h1" {
		t.Errorf("Expected h1='Main Heading', got %+v", fetchResponse.Headings[0])
	}
	if fetchResponse.Headings[1].Text != "Sub Heading" || fetchResponse.Headings[1].Level != "h2" {
		t.Errorf("Expected h2='Sub Heading', got %+v", fetchResponse.Headings[1])
	}
}

// Test case for Links extraction
func TestHandleFetchGet_Links(t *testing.T) {
	htmlContent := `<html><a href="https://site1.com">One</a><a href="https://site2.com">Two</a></html>`
	fetchService := setupMockServiceWithContent(htmlContent)

	request := httptest.NewRequest(http.MethodGet, "/fetch?url=http://example.com", nil)
	w := httptest.NewRecorder()
	fetchService.HandleFetchGet(w, request)

	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", response.StatusCode)
	}

	var fetchResponse models.FetchResponse
	if err := json.NewDecoder(response.Body).Decode(&fetchResponse); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	expected := []string{"https://site1.com", "https://site2.com"}
	if len(fetchResponse.Links) != len(expected) {
		t.Fatalf("Expected %d links, got %d", len(expected), len(fetchResponse.Links))
	}
	for i, link := range expected {
		if fetchResponse.Links[i] != link {
			t.Errorf("Expected link %s, got %s", link, fetchResponse.Links[i])
		}
	}
}

func TestHandleFetchGet_HTMLVersion(t *testing.T) {
	testHtmlArr := []struct {
		name            string
		htmlContent     string
		expectedVersion string
	}{
		{
			name:            "HTML5 doctype",
			htmlContent:     `<!DOCTYPE html><html><head><title>Doc</title></head><body></body></html>`,
			expectedVersion: "HTML5",
		},
		{
			name:            "HTML 4.01 doctype",
			htmlContent:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN"><html><head><title>Doc</title></head><body></body></html>`,
			expectedVersion: "HTML 4.01",
		},
		{
			name:            "No doctype",
			htmlContent:     `<html><head><title>Doc</title></head><body></body></html>`,
			expectedVersion: "Unknown",
		},
	}

	for _, testHtml := range testHtmlArr {
		t.Run(testHtml.name, func(t *testing.T) {
			fetchService := setupMockServiceWithContent(testHtml.htmlContent) 

			request := httptest.NewRequest(http.MethodGet, "/fetch?url=http://example.com", nil)
			w := httptest.NewRecorder()
			fetchService.HandleFetchGet(w, request)

			response := w.Result()
			if response.StatusCode != http.StatusOK {
				t.Fatalf("Expected 200, got %d", response.StatusCode)
			}

			var fetchResponse models.FetchResponse
			if err := json.NewDecoder(response.Body).Decode(&fetchResponse); err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if fetchResponse.HtmlVersion != testHtml.expectedVersion {
				t.Errorf("Expected HTML version %s, got %s", testHtml.expectedVersion, fetchResponse.HtmlVersion)
			}
		})
	}
}
