package service

import (
	"context"
	"encoding/json"
	"fmt"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LLM struct {
	model llms.LLM
}

func NewLLM(ctx context.Context, config Config) (repository.Speaker, error) {
	llm, err := ollama.New(
		ollama.WithModel(config.OllamaModel),
		ollama.WithServerURL(config.OllamaUrl),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama client: %v", err)
	}

	prompt := "Dis-moi si tu reçois ce message, par oui ou par non"

	_, err = llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	fmt.Println("LLM response:")
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %v", err)
	}
	return &LLM{model: llm}, nil
}

func (l *LLM) Talk(ctx context.Context, player *domain.Player, situation string) (string, error) {
	if !player.IsAlive {
		return "", fmt.Errorf("%s is not alive", player.Name)
	}

	systemPrompt := fmt.Sprintf(
		`Tu es %s, un joueur de Loups-Garous. Ton rôle SECRET est : %s. Ton tempérament est : %s.
		Tu dois réagir à la situation donnée. Ne révèle JAMAIS ton rôle explicitement si tu es Loup-Garou.
		Réponds par une seule phrase courte et percutante.`,
		player.Name, string(player.Role), string(player.Temperament),
	)
	fmt.Println(systemPrompt)
	response, err := l.model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, situation),
	})

	if err != nil {
		return "", fmt.Errorf("Erreur de réponse : %v", err)
	}

	return response.Choices[0].Content, nil
}

type VoteArgument struct {
	TargetName string `json:"target_name" description:"Le nom exact du joueur suspecté à éliminer"`
}

func (l *LLM) ChooseWhoToVote(ctx context.Context, player *domain.Player, suspects []string) string {
	voteTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "SubmitVote",
			Description: "Permet de voter officiellement contre un suspect pour l'éliminer du village.",
			Parameters:  VoteArgument{},
		},
	}

	prompt := fmt.Sprintf("Parmi les joueurs suivants : %v, choisis qui tu penses être le Loup-Garou et vote contre lui.", suspects)
	resp, err := l.model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf("Tu es %s (%s). Vote contre un suspect.", player.Name, player.Role)),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, llms.WithTools([]llms.Tool{voteTool}))

	if err != nil {
		return ""
	}

	if len(resp.Choices[0].ToolCalls) > 0 {
		toolCall := resp.Choices[0].ToolCalls[0]
		if toolCall.FunctionCall.Name == "SubmitVote" {
			var args VoteArgument
			json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args)
			return args.TargetName
		}
	}

	return ""
}
