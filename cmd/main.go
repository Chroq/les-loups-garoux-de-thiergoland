package main

import (
	"context"
	"log"

	"tiercelieux-llm-go/internal/application"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"
	"tiercelieux-llm-go/internal/interface/terminal"
	"tiercelieux-llm-go/internal/interface/websocket"
	"tiercelieux-llm-go/internal/service"
)

func main() {
	ctx := context.Background()
	config := service.NewConfig()

	var err error

	var playerSystem repository.PlayerSystem
	switch config.GameMode {
	case domain.GameModeLLM:
		log.Println("Testing connection to Ollama...")
		playerSystem, err = service.NewLLM(ctx, config)
		if err != nil {
			log.Fatalf("failed to create llm: %v", err)
		}
	case domain.GameModeRandom:
		playerSystem = service.NewRandomPlayerSystem()
	default:
		log.Fatalf("unknown game mode: %v", config.GameMode)
	}

	var displaySystem repository.DisplaySystem
	switch config.DisplayMode {
	case domain.DisplayModeTerminal:
		displaySystem = terminal.NewTerminalSystem()
	case domain.DisplayModeWebsocket:
		displaySystem = websocket.NewWSSystem()
	default:
		log.Fatalf("unknown display mode: %v", config.DisplayMode)
	}

	application.NewEngine(application.DefaultPlayerCount, playerSystem, displaySystem).Run(ctx)
}
