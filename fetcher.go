package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Djancyp/goharvest"
)

type Fetcher struct {
	cfg *Config
}

func NewFetcher(cfg *Config) (*Fetcher, error) {
	return &Fetcher{cfg: cfg}, nil
}

type FetchResult struct {
	URL         string
	StatusCode  int
	ContentType string
	Body        string
}

func (f *Fetcher) Fetch(ctx context.Context, targetURL string, method string, headers map[string]string, reqBody string) (*FetchResult, error) {
	if method != "" && strings.ToUpper(method) != "GET" {
		return f.fetchHTTP(ctx, targetURL, method, headers, reqBody)
	}
	return f.fetchWithGoharvest(ctx, targetURL, headers)
}

func (f *Fetcher) fetchWithGoharvest(ctx context.Context, targetURL string, headers map[string]string) (*FetchResult, error) {
	cookies := make([]map[string]string, 0)
	for k, v := range headers {
		if strings.EqualFold(k, "cookie") {
			cookies = append(cookies, map[string]string{"name": "", "value": v})
		}
	}

	scraper := &goharvest.Scrapper[WebPage]{
		Urls:              []string{targetURL},
		Timeout:           f.cfg.Timeout,
		UserAgent:         f.cfg.UserAgent,
		ParseFunc:         extractPageStructure,
		RobotsTxtDisabled: true,
		LogDisabled:       true,
		KeepBrowserOpen:   false,
	}

	if len(cookies) > 0 {
		scraper.Cookies = cookies
	}

	results, err := scraper.Scrape()
	if err != nil {
		return f.fetchHTTP(ctx, targetURL, "GET", headers, "")
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no content extracted from %s", targetURL)
	}

	page := results[0]
	yamlText := formatAsYAML(page)

	return &FetchResult{
		URL:         targetURL,
		StatusCode:  200,
		ContentType: "text/yaml",
		Body:        yamlText,
	}, nil
}

func (f *Fetcher) fetchHTTP(ctx context.Context, targetURL string, method string, headers map[string]string, reqBody string) (*FetchResult, error) {
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	var bodyReader io.Reader
	if reqBody != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		bodyReader = strings.NewReader(reqBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", f.cfg.UserAgent)
	req.Header.Set("Accept", "application/json,text/html,application/xml,text/plain,*/*")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	transport := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	if f.cfg.ProxyURL != "" {
		proxyURL, err := url.Parse(f.cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   f.cfg.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, f.cfg.MaxRespSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	truncated := int64(len(data)) > f.cfg.MaxRespSize
	if truncated {
		data = data[:f.cfg.MaxRespSize]
	}

	contentType := resp.Header.Get("Content-Type")
	text := ExtractText(data, contentType)

	result := text
	if truncated {
		result += "\n\n[... response truncated at " + formatSize(f.cfg.MaxRespSize) + " ...]"
	}

	return &FetchResult{
		URL:         targetURL,
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Body:        result,
	}, nil
}
