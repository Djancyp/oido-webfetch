# Oido WebFetch MCP Extension

Fetch URLs and inject optimized, structured content into the LLM context.

## Features

- **Structured HTML Extraction**: Parses HTML into YAML schema with headings, paragraphs, links, buttons, images, lists, tables, forms, and code blocks
- **Headless Browser**: Uses goharvest + Chromium for JS-heavy sites and anti-detection
- **API Support**: POST/PUT/PATCH with custom headers and body via plain HTTP
- **Proxy Support**: Route requests through an HTTP proxy
- **Auto-trigger**: Fetches URLs automatically when user provides one

## Installation

### Option 1: Upload via Plugins UI (Recommended)

1. Download the latest release zip from [GitHub Releases](../../releases)
2. Open Oido Studio → Plugins UI
3. Upload the zip file
4. Configure settings in the plugin settings panel

### Option 2: Build from Source

```bash
git clone <repo-url>
cd oido-webfetch
make build
```

## Requirements

- Go 1.23+
- Chromium browser (for goharvest HTML rendering)
  ```bash
  apt-get install chromium-browser   # Linux
  brew install chromium               # macOS
  ```

## Setup

| Variable | Description | Default |
|----------|-------------|---------|
| `OIDO_WEBFETCH_TIMEOUT` | Request timeout in seconds | `30` |
| `OIDO_WEBFETCH_USER_AGENT` | User-Agent header | `OidoWebFetch/1.0` |
| `OIDO_WEBFETCH_PROXY` | HTTP proxy URL | *(none)* |
| `OIDO_WEBFETCH_MAX_RESPONSE_SIZE` | Max response bytes (raw fetches) | `4194304` |

## Tools

### `webfetch_fetch_url`

Fetches a URL. HTML pages are parsed into structured YAML. API calls return raw text.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | yes | Full URL including scheme |
| `method` | string | no | HTTP method (GET, POST, PUT, PATCH). Default: GET |
| `headers` | object | no | Custom HTTP headers as key-value pairs |
| `body` | string | no | Request body for POST/PUT/PATCH |

## Architecture

```
┌─────────────────┐    stdio     ┌──────────────────────┐
│  Oido Studio    │ ◄──────────► │  oido-webfetch-mcp   │
│                 │              │                      │
│                 │              │  ┌────────────────┐  │
│                 │              │  │ goharvest      │──►── Chromium → Internet
│                 │              │  │ (HTML→YAML)    │  │
│                 │              │  └────────────────┘  │
│                 │              │  ┌────────────────┐  │
│                 │              │  │ net/http       │──►── API endpoints
│                 │              │  │ (raw text)     │  │
│                 │              │  └────────────────┘  │
└─────────────────┘              └──────────────────────┘
```

## License

MIT
