package main

import (
	"strings"

	"golang.org/x/net/html"
)

const maxChars = 4_000_000

var knownCodeExts = map[string]bool{
	".txt": true, ".md": true, ".csv": true, ".log": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true,
	".xml": true, ".html": true, ".htm": true,
	".js": true, ".ts": true, ".go": true, ".py": true,
	".rs": true, ".java": true, ".c": true, ".cpp": true,
	".h": true, ".sh": true, ".bash": true, ".sql": true,
	".css": true, ".scss": true, ".less": true,
	".proto": true, ".graphql": true, ".tf": true,
	".dockerfile": true,
}

func ExtractText(data []byte, contentType string) string {
	ct := strings.ToLower(contentType)

	switch {
	case strings.Contains(ct, "text/html"), strings.Contains(ct, "application/xhtml"):
		return extractHTML(data)
	case strings.Contains(ct, "application/json"):
		return tryFormatJSON(string(data))
	case strings.Contains(ct, "text/"), strings.Contains(ct, "application/xml"), strings.Contains(ct, "application/javascript"):
		return truncate(string(data))
	default:
		if looksLikeText(data) {
			return truncate(string(data))
		}
		return truncate(string(data))
	}
}

func extractHTML(data []byte) string {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return truncate(string(data))
	}

	var sb strings.Builder
	extractTextNode(doc, &sb)
	return truncate(strings.TrimSpace(sb.String()))
}

func extractTextNode(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(text)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "script" || c.Data == "style" || c.Data == "noscript" {
			continue
		}
		if c.Type == html.ElementNode {
			switch c.Data {
			case "p", "div", "br", "h1", "h2", "h3", "h4", "h5", "h6",
				"li", "tr", "td", "th", "blockquote", "pre", "section", "article":
				sb.WriteByte('\n')
			case "a":
			}
		}
		extractTextNode(c, sb)
	}
}

func tryFormatJSON(s string) string {
	return truncate(s)
}

func looksLikeText(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	printable := 0
	check := data
	if len(check) > 512 {
		check = check[:512]
	}
	for _, b := range check {
		if b >= 32 || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}
	return float64(printable)/float64(len(check)) > 0.8
}

func truncate(s string) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "\n\n[... truncated at 4M characters ...]"
}
