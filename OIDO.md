# Oido WebFetch Extension

Fetch URLs and inject optimized, structured content into the LLM context.

## Available Tools

### `webfetch_fetch_url`
Fetch a URL and return its content optimized for LLM consumption.

HTML pages are automatically parsed into a structured YAML schema:
- **title**, **description** — page metadata
- **h1-h6** — heading hierarchy
- **paragraphs** — text content (truncated at 300 chars each)
- **links** — text + href pairs
- **buttons**, **images** — with src and alt
- **lists**, **tables** — with headers and rows
- **forms** — action, method, and inputs
- **code_blocks** — extracted from `<pre>`/`<code>` tags

For API endpoints (JSON/XML) or non-GET methods, raw response text is returned.

**Parameters:**
- `url` (string, required): Full URL including scheme (e.g. `https://example.com/page`)
- `method` (string, optional): HTTP method. Default: GET
- `headers` (object, optional): Custom HTTP headers as key-value pairs
- `body` (string, optional): Request body for POST/PUT/PATCH

## When to Auto-Trigger

Call `webfetch_fetch_url` automatically when:
- User provides a URL and asks what's on the page
- User asks to look up documentation, articles, or web content
- User wants to fetch data from an API endpoint
- User pastes a URL without context (assume they want its content)

**Do NOT ask for confirmation** — just fetch and return the content.

## Example Output (HTML page)

```yaml
url: https://example.com
title: "Example Domain"
description: "This domain is for use in illustrative examples"
h1:
  - "Example Domain"
paragraphs:
  - "This domain is for use in illustrative examples in documents..."
links:
  - text: "More information..."
    href: "https://www.iana.org/domains/example"
```

## When to Use Each Mode

| Mode | Trigger | Backend |
|------|---------|---------|
| **Structured YAML** | GET request to HTML pages | goharvest + chromedp |
| **Raw text** | POST/PUT/PATCH, or JSON/XML API endpoints | net/http |

## Requirements

- Chromium browser installed (for HTML page rendering via goharvest)
  ```bash
  # Ubuntu/Debian
  apt-get install chromium-browser
  # macOS
  brew install chromium
  ```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OIDO_WEBFETCH_TIMEOUT` | Request timeout in seconds | `30` |
| `OIDO_WEBFETCH_USER_AGENT` | User-Agent header | `OidoWebFetch/1.0` |
| `OIDO_WEBFETCH_PROXY` | HTTP proxy URL | *(none)* |
| `OIDO_WEBFETCH_MAX_RESPONSE_SIZE` | Max response bytes for raw fetches | `4194304` |
