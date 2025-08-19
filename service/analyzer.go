package service

import (
	"fmt"
	"strings"
	"golang.org/x/net/html"
	"log/slog"
	"net/http"
	"net/url"
	"time"

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
	resp := models.FetchResponse{
		Headings:        []models.Heading{},
		Links:           []string{},
		LoginDetected:   false,
		LoginIndicators: []string{},
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		resp.Error = err.Error()
		ha.logger.Error("Failed to parse HTML", "err", err)
		return resp
	}

	resp.Title = strings.TrimSpace(doc.Find("title").First().Text())

	resp.HtmlVersion = findHtmlVersion(content)

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
				resp.Headings = append(resp.Headings, models.Heading{
					Level: selector,
					Text:  text,
				})
			}
		})
	}
	linksSet := map[string]struct{}{}
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		href = strings.TrimSpace(href)
		if href != "" && !strings.HasPrefix(href, "javascript:") && !strings.HasPrefix(href, "#") {
			if _, exists := linksSet[href]; !exists {
				linksSet[href] = struct{}{}
				resp.Links = append(resp.Links, href)
			}
		}
	})
	base, _ := url.Parse(baseURL)
	internalLinks := map[string]struct{}{}
	externalLinks := map[string]struct{}{}
	inaccessibleLinks := map[string]struct{}{}

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

		// Classify internal vs external
		if base != nil && u.Host == base.Host {
			internalLinks[absURL] = struct{}{}
		} else {
			externalLinks[absURL] = struct{}{}
		}

		// Check accessibility (simple HEAD request with timeout)
		client := &http.Client{Timeout: 1 * time.Second}
		respHead, err := client.Head(absURL)
		if err != nil || respHead.StatusCode < 200 || respHead.StatusCode >= 400 {
			inaccessibleLinks[absURL] = struct{}{}
		}
	})

	// Fill response
	resp.InternalLinks = len(internalLinks)
	resp.ExternalLinks = len(externalLinks)
	resp.InaccessibleLinks = len(inaccessibleLinks)

	if doc.Find("input[type='password']").Length() > 0 {
		resp.LoginDetected = true
		resp.LoginIndicators = append(resp.LoginIndicators, "password_input")
	}

	pageText := strings.ToLower(normalizeSpace(doc.Text()))
	if strings.Contains(pageText, "login") || strings.Contains(pageText, "sign in") {
		resp.LoginDetected = true
		resp.LoginIndicators = append(resp.LoginIndicators, "page_text_contains_login")
	}

	resp.LoginIndicators = uniqueStrings(resp.LoginIndicators)
	ha.logger.Info("HTML analysis completed", "title", resp.Title, "headings", len(resp.Headings), "links", len(resp.Links), "login_detected", resp.LoginDetected)
	return resp
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
