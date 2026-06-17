package service

import (
	"bufio"
	"log"
	"os"
	"strings"
)

const (
	envFile = ".env"

	envOllamaUrl   = "OLLAMA_URL"
	envOllamaModel = "OLLAMA_MODEL"

	commentPrefix = "#"
	separator     = "="
)

type Config struct {
	OllamaUrl   string
	OllamaModel string
}

func NewConfig() Config {
	var config Config

	file, err := os.Open(envFile)
	if err != nil {
		log.Fatalf("failed to open %s file: %v", envFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, commentPrefix) {
			continue
		}
		parts := strings.SplitN(line, separator, 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch key {
			case envOllamaUrl:
				config.OllamaUrl = value
			case envOllamaModel:
				config.OllamaModel = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to scan %s file: %v", envFile, err)
	}

	return config
}
