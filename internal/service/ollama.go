package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"
	"tiercelieux-llm-go/internal/logger"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	ToolCallingVote = "Vote"
)

type LLM struct {
	model llms.Model
}

func NewLLM(ctx context.Context, config Config) (repository.PlayerSystem, error) {
	var llm llms.Model
	var err error
	if config.LlmApiProvider == "gemini" {
		llm, err = googleai.New(ctx, googleai.WithAPIKey(config.LlmApiKey))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Use OpenAI-compatible client pointing to Ollama's local URL.
		// Ollama's OpenAI API is at config.OllamaUrl + "/v1" (e.g. http://localhost:11434/v1).
		// This provides native tool calling support which the langchaingo ollama client lacks.
		llm, err = openai.New(
			openai.WithBaseURL(config.OllamaUrl+"/v1"),
			openai.WithModel(config.OllamaModel),
			openai.WithToken("ollama"), // Ollama does not require a token but openai client needs one configured
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create ollama client: %v", err)
		}
	}

	return &LLM{model: llm}, nil
}

func (l *LLM) Talk(ctx context.Context, player domain.PlayerInterface, situation string) (string, error) {
	systemPrompt := fmt.Sprintf(
		`Tu es %s, un joueur de Loups-Garous. 
		Ton rôle SECRET est : %s. 
		Ton tempérament est : %s.
		Tu dois réagir en tenant compte de la situation donnée.
		Le joueur éliminé est forcément un villageois. 
		Même en temps que loup garou, tu dois jouer le jeu et accuser un villageois d'être un loup garou.
		Ne révèle JAMAIS ton rôle explicitement si tu es Loup-Garou.
		Ne produis AUCUN raisonnement ou pensée intermédiaire (ne commence pas par <think>). 
		Réponds directement par une seule phrase courte et percutante.`,
		player.Name(), player.Role(), player.Temperament(),
	)

	response, err := l.model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, situation),
	}, llms.WithMaxTokens(60))

	if err != nil {
		return "", fmt.Errorf("Erreur de réponse : %v", err)
	}

	reply := response.Choices[0].Content
	if idx := strings.Index(reply, "</think>"); idx != -1 {
		reply = reply[idx+8:]
	}
	reply = strings.TrimSpace(reply)

	return reply, nil
}

type VoteArgument struct {
	Name string `json:"name" description:"Le nom exact du joueur suspecté à éliminer"`
}

func (l *LLM) ChooseWhoToVote(ctx context.Context, player domain.PlayerInterface, debate string, suspects []domain.PlayerInterface) (string, error) {
	var suspectNames []string
	for _, s := range suspects {
		suspectNames = append(suspectNames, s.Name())
	}
	suspectsListStr := strings.Join(suspectNames, ", ")

	voteTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        ToolCallingVote,
			Description: "Permet de voter officiellement contre un suspect pour l'éliminer du village.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Le nom exact du joueur suspecté à éliminer",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	prompt := fmt.Sprintf(
		`Tu es %s et voici le débat : %v. Vote pour le joueur que tu penses être le Loup-Garou parmi la liste des suspects suivants : %s`,
		player.Name(), debate, suspectsListStr)

	resp, err := l.model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem,
			`Ne produis aucun raisonnement ou pensée intermédiaire, vote en te basant sur le débat.
			Tu ne peux pas voter contre toi même. 
			Réponds uniquement avec un seul Tool Call de type "Vote".`),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, llms.WithTools([]llms.Tool{voteTool}))
	if err != nil {
		logger.Errorf("Error: %v\n", err)
		return "", err
	}

	logger.Debugf("resp.Choices: %v\n", resp.Choices)
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		logger.Debugf("choice.ToolCalls: %v\n", choice.ToolCalls)
		if len(choice.ToolCalls) > 0 {
			toolCall := choice.ToolCalls[0]
			logger.Debugf("Tool Call: %v\n", toolCall)
			if toolCall.FunctionCall.Name == ToolCallingVote {
				var args VoteArgument
				json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args)
				target := strings.TrimSpace(args.Name)
				logger.Debugf("Target: %s\n", target)
				for _, suspect := range suspects {
					if strings.EqualFold(suspect.Name(), target) {
						return suspect.Name(), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("No target found, picking randomly")
}

func (l *LLM) ChooseWhoToEat(ctx context.Context, player *domain.Werewolf, villagers map[string]domain.Villager) *domain.Villager {
	var victim *domain.Villager
	for _, v := range villagers {
		victim = &v
		break
	}

	return victim
}
