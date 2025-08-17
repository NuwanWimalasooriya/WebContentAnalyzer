package service

import (
	"fmt"
	"strings"

	"log/slog"

	models "github.com/nuwanwimalasooriya/go-wa-api/models"
	"github.com/PuerkitoBio/goquery"
)

type Analyzer interface {
	Analyze(content string) models.FetchResponse
}

type HTMLAnalyzer struct {
	logger *slog.Logger
}

func NewHTMLAnalyzer(logger *slog.Logger) *HTMLAnalyzer {
	return &HTMLAnalyzer{logger: logger}
}

func (ha *HTMLAnalyzer) Analyze(content string) models.FetchResponse {
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
