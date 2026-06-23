package service

import (
	"context"
	"math/rand"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"
	"time"
)

type RandomPlayerSystem struct {
	rand *rand.Rand
}

func NewRandomPlayerSystem() repository.PlayerSystem {
	return RandomPlayerSystem{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r RandomPlayerSystem) ChooseWhoToVote(ctx context.Context, player domain.PlayerInterface, debate string, suspects []domain.PlayerInterface) (string, error) {
	idx := r.rand.Intn(len(suspects))
	return suspects[idx].Name(), nil
}

func (r RandomPlayerSystem) Talk(ctx context.Context, player domain.PlayerInterface, situation string) (string, error) {
	return "Bla bla bla", nil
}

func (r RandomPlayerSystem) ChooseWhoToEat(ctx context.Context, player *domain.Werewolf, villagers map[string]domain.Villager) *domain.Villager {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rnd.Intn(len(villagers))

	var i int
	for _, villager := range villagers {
		if idx == i {
			return &villager
		}
		i++
	}

	return nil
}
