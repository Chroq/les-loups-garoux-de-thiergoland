package repository

import (
	"context"
	"tiercelieux-llm-go/internal/domain"
)

type PlayerSystem interface {
	Talk(ctx context.Context, player domain.PlayerInterface, situation string) (string, error)
	ChooseWhoToVote(ctx context.Context, player domain.PlayerInterface, debate string, suspects []domain.PlayerInterface) (string, error)
	ChooseWhoToEat(ctx context.Context, player *domain.Werewolf, villagers map[string]domain.Villager) *domain.Villager
}
