package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type configJSON struct {
	OIDO_WEBFETCH_TIMEOUT          string `json:"OIDO_WEBFETCH_TIMEOUT"`
	OIDO_WEBFETCH_USER_AGENT      string `json:"OIDO_WEBFETCH_USER_AGENT"`
	OIDO_WEBFETCH_PROXY           string `json:"OIDO_WEBFETCH_PROXY"`
	OIDO_WEBFETCH_MAX_RESPONSE_SIZE string `json:"OIDO_WEBFETCH_MAX_RESPONSE_SIZE"`
}

func main() {
	configStr := flag.String("config", "", "JSON string with WebFetch settings (overrides .env, overridden by env vars)")
	flag.Parse()

	if *configStr != "" {
		var cfg configJSON
		if err := json.Unmarshal([]byte(*configStr), &cfg); err != nil {
			log.Fatalf("Failed to parse --config JSON: %v", err)
		}
		setIfNotEnv := func(key, val string) {
			if val != "" && os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
		setIfNotEnv("OIDO_WEBFETCH_TIMEOUT", cfg.OIDO_WEBFETCH_TIMEOUT)
		setIfNotEnv("OIDO_WEBFETCH_USER_AGENT", cfg.OIDO_WEBFETCH_USER_AGENT)
		setIfNotEnv("OIDO_WEBFETCH_PROXY", cfg.OIDO_WEBFETCH_PROXY)
		setIfNotEnv("OIDO_WEBFETCH_MAX_RESPONSE_SIZE", cfg.OIDO_WEBFETCH_MAX_RESPONSE_SIZE)
	}

	log.Println("Starting Oido WebFetch MCP Server v1.0.0...")
	RunMCPServer()
}
