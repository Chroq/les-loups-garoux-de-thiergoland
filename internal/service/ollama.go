package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LLM struct {
	model llms.Model
}

func NewLLM(ctx context.Context, config Config) (repository.PlayerSystem, error) {
	llm, err := ollama.New(
		ollama.WithModel(config.OllamaModel),
		ollama.WithServerURL(config.OllamaUrl),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama client: %v", err)
	}
	return &LLM{model: llm}, nil
}

func (l *LLM) Talk(ctx context.Context, player domain.PlayerInterface, situation string) (string, error) {
	systemPrompt := fmt.Sprintf(
		`Tu es %s, un joueur de Loups-Garous. Ton rôle SECRET est : %s. Ton tempérament est : %s.
		Tu dois réagir à la situation donnée. Ne révèle JAMAIS ton rôle explicitement si tu es Loup-Garou.
		Ne produis AUCUN raisonnement ou pensée intermédiaire (ne commence pas par <think>). Réponds directement par une seule phrase courte et percutante.`,
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
	TargetName string `json:"target_name" description:"Le nom exact du joueur suspecté à éliminer"`
}

func (l *LLM) ChooseWhoToVote(ctx context.Context, player domain.PlayerInterface, debate string, suspects []domain.PlayerInterface) (string, error) {
	voteTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "SubmitVote",
			Description: "Permet de voter officiellement contre un suspect pour l'éliminer du village.",
			Parameters:  VoteArgument{},
		},
	}

	prompt := fmt.Sprintf(
		`Tu es %s et voici le débat : %v. Vote pour le joueur que tu penses être le Loup-Garou, par contre tu ne peux pas voter contre toi même.`, //
		player.Name(), debate)

	resp, err := l.model.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf("Tu es %s (%s). Ne produis aucun raisonnement ou pensée intermédiaire. Vote directement contre un suspect.", player.Name(), player.Role())),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, llms.WithTools([]llms.Tool{voteTool}), llms.WithMaxTokens(60))

	log.Default().Println("Response : \n", resp)
	log.Default().Printf("Error : %v\n", err)

	if err == nil && len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if len(choice.ToolCalls) > 0 {
			toolCall := choice.ToolCalls[0]
			if toolCall.FunctionCall.Name == "SubmitVote" {
				var args VoteArgument
				json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args)
				target := strings.TrimSpace(args.TargetName)
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
