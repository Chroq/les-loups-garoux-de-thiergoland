package main

import (
	"context"
	"fmt"
	"log"

	"tiercelieux-llm-go/internal/application"
	"tiercelieux-llm-go/internal/service"
)

func main() {
	ctx := context.Background()
	config := service.NewConfig()

	fmt.Println("Testing connection to Ollama...")
	llm, err := service.NewLLM(ctx, config)
	if err != nil {
		log.Fatalf("failed to create llm: %v", err)
	}

	fmt.Println("Connection successful!")

	iaNames := []string{"Bob", "Alice", "Charlie", "Eva", "Chris", "Jeanne"}
	game := application.NewGame(iaNames, llm)
	game.DisplayVillage()
	game.Run(ctx)

}
