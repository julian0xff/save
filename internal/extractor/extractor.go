package extractor

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	readability "github.com/go-shiori/go-readability"
	htmltomd "github.com/JohannesKaufmann/html-to-markdown/v2"
)

type Result struct {
	URL         string
	Title       string
	Author      string
	Content     string // markdown
	TextContent string // plain text
	Excerpt     string
	SiteName    string
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func Extract(rawURL string) (*Result, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: HTTP %d", rawURL, resp.StatusCode)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return nil, fmt.Errorf("extracting article: %w", err)
	}

	markdown, err := htmltomd.ConvertString(article.Content)
	if err != nil {
		// Fall back to plain text if markdown conversion fails
		markdown = article.TextContent
	}

	return &Result{
		URL:         rawURL,
		Title:       article.Title,
		Author:      article.Byline,
		Content:     markdown,
		TextContent: article.TextContent,
		Excerpt:     article.Excerpt,
		SiteName:    article.SiteName,
	}, nil
}
