package service

import (
	"bufio"
	"log"
	"os"
	"strings"
	"tiercelieux-llm-go/internal/domain"
)

const (
	envFile = ".env"

	envLLMApiKey      = "LLM_API_KEY"
	envLLMApiProvider = "LLM_API_PROVIDER"

	envOllamaUrl   = "OLLAMA_URL"
	envOllamaModel = "OLLAMA_MODEL"

	envGameMode    = "GAME_MODE"
	envDisplayMode = "DISPLAY_MODE"

	commentPrefix = "#"
	separator     = "="
)

type Config struct {
	LlmApiKey      string
	LlmApiProvider string

	OllamaUrl   string
	OllamaModel string
	GameMode    domain.GameMode
	DisplayMode domain.DisplayMode
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
			case envLLMApiKey:
				config.LlmApiKey = value
			case envLLMApiProvider:
				config.LlmApiProvider = value
			case envOllamaUrl:
				config.OllamaUrl = value
			case envOllamaModel:
				config.OllamaModel = value
			case envGameMode:
				switch value {
				case "random":
					config.GameMode = domain.GameModeRandom
				case "llm":
					config.GameMode = domain.GameModeLLM
				default:
					log.Fatalf("invalid game mode: %s", value)
				}
			case envDisplayMode:
				switch value {
				case "terminal":
					config.DisplayMode = domain.DisplayModeTerminal
				case "websocket":
					config.DisplayMode = domain.DisplayModeWebsocket
				default:
					log.Fatalf("invalid display mode: %s", value)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to scan %s file: %v", envFile, err)
	}

	return config
}
