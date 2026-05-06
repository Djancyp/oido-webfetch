package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPHandler struct {
	fetcher *Fetcher
	cfg     *Config
}

type FetchURLArgs struct {
	URL     string            `json:"url" jsonschema:"URL to fetch (e.g. https://example.com/page)"`
	Method  string            `json:"method,omitempty" jsonschema:"HTTP method (GET, POST, PUT, PATCH). Default: GET"`
	Headers map[string]string `json:"headers,omitempty" jsonschema:"Optional HTTP headers as key-value pairs"`
	Body    string            `json:"body,omitempty" jsonschema:"Request body for POST/PUT/PATCH requests"`
}

func RunMCPServer() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("Warning: config error (using defaults): %v", err)
		cfg = &Config{Timeout: 30 * time.Second, UserAgent: "OidoWebFetch/1.0", MaxRespSize: 4 * 1024 * 1024}
	}

	fetcher, err := NewFetcher(cfg)
	if err != nil {
		log.Printf("Warning: fetcher init failed: %v", err)
		fetcher = &Fetcher{cfg: cfg}
	}

	handler := &MCPHandler{fetcher: fetcher, cfg: cfg}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "oido-webfetch",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name: "webfetch_fetch_url",
		Description: `Fetch a URL and return its content optimized for LLM consumption.

AUTO-TRIGGER: call this tool automatically whenever the user provides a URL or asks to look up a webpage. Extract the URL from the user message. Do not ask for confirmation.

HTML pages are automatically parsed into a structured YAML schema with sections:
  - title, description (meta tags)
  - h1, h2, h3, h4, h5, h6 (headings)
  - paragraphs (text content, truncated at 300 chars each)
  - links (text + href)
  - buttons, images (src + alt)
  - lists, tables (headers + rows)
  - forms (action, method, inputs)
  - code_blocks

For API endpoints (JSON/XML) or non-GET methods, returns raw response text.
Follows redirects (up to 10).`,
	}, handler.HandleFetchURL)

	ctx := context.Background()
	log.Println("Oido WebFetch MCP Server starting on stdio...")

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func (h *MCPHandler) HandleFetchURL(ctx context.Context, _ *mcp.CallToolRequest, args FetchURLArgs) (*mcp.CallToolResult, any, error) {
	if args.URL == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: url parameter is required"}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.fetcher.Fetch(ctx, args.URL, args.Method, args.Headers, args.Body)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	output := fmt.Sprintf("URL: %s\nStatus: %d\nContent-Type: %s\n\n%s",
		result.URL, result.StatusCode, result.ContentType, result.Body)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Error: " + msg},
		},
		IsError: true,
	}
}
