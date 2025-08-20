package service

import (
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"sync"

	models "github.com/nuwanwimalasooriya/go-wa-api/models"
	"github.com/PuerkitoBio/goquery"
)

type Analyzer interface {
	Analyze(content string, baseURL string) models.FetchResponse
}

type HTMLAnalyzer struct {
	logger *slog.Logger
}

func NewHTMLAnalyzer(logger *slog.Logger) *HTMLAnalyzer {
	return &HTMLAnalyzer{logger: logger}
}

func (ha *HTMLAnalyzer) Analyze(content string, baseURL string) models.FetchResponse {
	start := time.Now()
	response := models.FetchResponse{
		Headings:        []models.Heading{},
		Links:           []string{},
		LoginDetected:   false,
		LoginIndicators: []string{},
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	ha.logger.Info("NewDocumentFromReader execution time", "duration", time.Since(start))
	if err != nil {
		response.Error = err.Error()
		ha.logger.Error("Failed to parse HTML", "err", err)
		return response
	}

	response.Title = strings.TrimSpace(doc.Find("title").First().Text())

	response.HtmlVersion = findHtmlVersion(content)
	ha.logger.Info("findHtmlVersion execution time", "duration", time.Since(start))
	headingsSet := map[string]struct{}{}
	for i := 1; i <= 6; i++ {
		selector := fmt.Sprintf("h%d", i)
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text == "" {
				return
			}
			text = normalizeSpace(text)
			if text == "" {
				return
			}
			key := fmt.Sprintf("%s:%s", selector, text)
			if _, exists := headingsSet[key]; !exists {
				headingsSet[key] = struct{}{}
				response.Headings = append(response.Headings, models.Heading{
					Level: selector,
					Text:  text,
				})
			}
		})
	}
	ha.logger.Info("headingsSet execution time", "duration", time.Since(start))
	linksSet := map[string]struct{}{}
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		href = strings.TrimSpace(href)
		if href != "" && !strings.HasPrefix(href, "javascript:") && !strings.HasPrefix(href, "#") {
			if _, exists := linksSet[href]; !exists {
				linksSet[href] = struct{}{}
				response.Links = append(response.Links, href)
			}
		}
	})
	ha.logger.Info("linksSet execution time", "duration", time.Since(start))
	base, _ := url.Parse(baseURL)
	internalLinks := map[string]struct{}{}
	externalLinks := map[string]struct{}{}
	inaccessibleLinks := map[string]struct{}{}
	start1:=time.Now()
	

		client := &http.Client{Timeout: 2 * time.Second}
		var wg sync.WaitGroup
		var mu sync.Mutex

			doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				href = strings.TrimSpace(href)
				if href == "" || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "#") {
					return
				}

				u, err := url.Parse(href)
				if err != nil {
					return
				}

				absURL := u.String()
				if !u.IsAbs() && base != nil {
					absURL = base.ResolveReference(u).String()
				}

				if base != nil && u.Host == base.Host {
					internalLinks[absURL] = struct{}{}
				} else {
					externalLinks[absURL] = struct{}{}
				}

				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					response, err := client.Head(url)
					if err != nil || response.StatusCode < 200 || response.StatusCode >= 400 {
						mu.Lock()
						inaccessibleLinks[url] = struct{}{}
						mu.Unlock()
					}
				}(absURL)
			})

			wg.Wait()
	ha.logger.Info("linktypesanalyze execution time", "duration", time.Since(start1))
	// Fill response
	response.InternalLinks = len(internalLinks)
	response.ExternalLinks = len(externalLinks)
	response.InaccessibleLinks = len(inaccessibleLinks)

	if doc.Find("input[type='password']").Length() > 0 {
		response.LoginDetected = true
		response.LoginIndicators = append(response.LoginIndicators, "password_input")
	}

	pageText := strings.ToLower(normalizeSpace(doc.Text()))
	if strings.Contains(pageText, "login") || strings.Contains(pageText, "sign in") {
		response.LoginDetected = true
		response.LoginIndicators = append(response.LoginIndicators, "page_text_contains_login")
	}

	response.LoginIndicators = uniqueStrings(response.LoginIndicators)
	ha.logger.Info("HTML analysis completed", "title", response.Title, "headings", len(response.Headings), "links", len(response.Links), "login_detected", response.LoginDetected)
	ha.logger.Info("login check execution time", "duration", time.Since(start1))
	return response
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func uniqueStrings(in []string) []string {
	out := []string{}
	set := map[string]struct{}{} 
	for _, v := range in {
		if _, exists := set[v]; !exists { 
			set[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

func findHtmlVersion(htmlDoc string)string{
		reader:=strings.NewReader(htmlDoc)	
		tokenizer:=html.NewTokenizer(reader)

for {
        token := tokenizer.Next()

        switch token {
        case html.ErrorToken:
            return "Unknown" // reached EOF without finding doctype

        case html.DoctypeToken:
            t := tokenizer.Token()
            doc := strings.ToLower(t.Data)

            // Basic checks
            if doc == "html" && len(t.Attr) == 0 {
                return "HTML5"
            }

            // Check for known doctypes in attributes
            doctypeStr := t.Data
            for _, attr := range t.Attr {
                doctypeStr += " " + attr.Val
            }

            if strings.Contains(doctypeStr, "xhtml") {
                return "XHTML"
            }
            if strings.Contains(doctypeStr, "4.01") {
                return "HTML 4.01"
            }
            if strings.Contains(doctypeStr, "transitional") {
                return "HTML 4.01 Transitional"
            }
            if strings.Contains(doctypeStr, "strict") {
                return "HTML 4.01 Strict"
            }

            return "Unknown"
        }
    }
}
