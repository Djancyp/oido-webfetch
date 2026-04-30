package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Timeout       time.Duration
	UserAgent     string
	ProxyURL      string
	MaxRespSize   int64
}

func LoadConfig() (*Config, error) {
	timeoutSec := 30
	if v := os.Getenv("OIDO_WEBFETCH_TIMEOUT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("OIDO_WEBFETCH_TIMEOUT must be a number (seconds): %w", err)
		}
		timeoutSec = n
	}

	ua := os.Getenv("OIDO_WEBFETCH_USER_AGENT")
	if ua == "" {
		ua = "OidoWebFetch/1.0"
	}

	proxyURL := os.Getenv("OIDO_WEBFETCH_PROXY")

	maxSize := int64(4 * 1024 * 1024)
	if v := os.Getenv("OIDO_WEBFETCH_MAX_RESPONSE_SIZE"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("OIDO_WEBFETCH_MAX_RESPONSE_SIZE must be a number (bytes): %w", err)
		}
		maxSize = n
	}

	return &Config{
		Timeout:     time.Duration(timeoutSec) * time.Second,
		UserAgent:   ua,
		ProxyURL:     proxyURL,
		MaxRespSize: maxSize,
	}, nil
}
