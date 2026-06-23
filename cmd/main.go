package main

import (
	"context"
	"fmt"
	"log"

	"tiercelieux-llm-go/internal/application"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"
	"tiercelieux-llm-go/internal/service"
)

func main() {
	ctx := context.Background()
	config := service.NewConfig()

	var err error
	var playerSystem repository.PlayerSystem
	switch config.GameMode {
	case domain.GameModeLLM:
		fmt.Println("Testing connection to Ollama...")
		playerSystem, err = service.NewLLM(ctx, config)
		if err != nil {
			log.Fatalf("failed to create llm: %v", err)
		}
	case domain.GameModeRandom:
		playerSystem = service.NewRandomPlayerSystem()
	default:
		log.Fatalf("unknown game mode: %v", config.GameMode)
	}

	game := application.NewEngine(application.DefaultPlayerCount, playerSystem)

	fmt.Println("============ DÉBUT DE LA SIMULATION ============")
	game.Game.DisplayAllRoles()
	game.Run(ctx)
}
