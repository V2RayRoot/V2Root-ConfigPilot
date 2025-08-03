package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
	"github.com/v2root/configpilot/parser"
	"github.com/v2root/configpilot/scorer"
	"github.com/v2root/configpilot/tester"
	"golang.org/x/exp/slog"
)

type Config struct {
	Protocol string
	URL      string
	Score    float64
	Results  tester.TestResults
}

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Read configs from FetchConfig.py output
	configsJSON, err := os.ReadFile("configs.json")
	if err != nil {
		logger.Error("Failed to read configs.json", err)
		os.Exit(1)
	}

	var rawConfigs map[string][]string
	if err := json.Unmarshal(configsJSON, &rawConfigs); err != nil {
		logger.Error("Failed to parse configs.json", err)
		os.Exit(1)
	}

	// Parse configs
	var configs []Config
	for protocol, urls := range rawConfigs {
		for _, url := range urls {
			if parser.ValidateURL(protocol, url) {
				configs = append(configs, Config{Protocol: protocol, URL: url})
			} else {
				logger.Warn("Invalid URL", "protocol", protocol, "url", url)
			}
		}
	}

	logger.Info("Parsed configs", "count", len(configs))

	// Test configs concurrently
	maxConcurrency := 10 // Adjustable via env var
	concurrency := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	results := make([]Config, len(configs))
	copy(results, configs)

	for i := range results {
		wg.Add(1)
		concurrency <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-concurrency }()
			config := &results[idx]
			config.Results = tester.TestConfig(config.Protocol, config.URL, 10*time.Second)
			config.Score = scorer.CalculateScore(config.Results)
			logger.Info("Tested config", "url", config.URL, "score", config.Score)
		}(i)
	}
	wg.Wait()

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Select top 10
	if len(results) > 10 {
		results = results[:10]
	}

	// Save outputs
	if err := os.MkdirAll("output", 0755); err != nil {
		logger.Error("Failed to create output directory", err)
		os.Exit(1)
	}

	// Save BestConfigs.txt
	txtPath := filepath.Join("output", "BestConfigs.txt")
	txtFile, err := os.Create(txtPath)
	if err != nil {
		logger.Error("Failed to create BestConfigs.txt", err)
		os.Exit(1)
	}
	defer txtFile.Close()
	for _, config := range results {
		fmt.Fprintf(txtFile, "%s\n", config.URL)
	}
	logger.Info("Saved BestConfigs.txt", "path", txtPath)

	// Save BestConfigs_scored.json
	jsonPath := filepath.Join("output", "BestConfigs_scored.json")
	scoredData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal scored configs", err)
		os.Exit(1)
	}
	if err := os.WriteFile(jsonPath, scoredData, 0644); err != nil {
		logger.Error("Failed to write BestConfigs_scored.json", err)
		os.Exit(1)
	}
	logger.Info("Saved BestConfigs_scored.json", "path", jsonPath)

}
