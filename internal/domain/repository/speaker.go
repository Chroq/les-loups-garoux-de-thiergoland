package repository

import (
	"context"
	"tiercelieux-llm-go/internal/domain"
)

type Speaker interface {
	Talk(ctx context.Context, player *domain.Player, situation string) (string, error)
	ChooseWhoToVote(ctx context.Context, player *domain.Player, suspects []string) string
}
