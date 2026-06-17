package application

import (
	"context"
	"fmt"
	"math/rand"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"
	"time"
)

type Game struct {
	Players        []domain.Player
	Turn           int
	SpeakingSystem repository.Speaker
}

func NewGame(iaNames []string, speakingSystem repository.Speaker) *Game {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	roles := []domain.Role{
		domain.RoleWerewolf,
		domain.RoleWerewolf,
		domain.RoleVillager,
		domain.RoleVillager,
		domain.RoleVillager,
		domain.RoleVillager,
	}
	temperaments := []domain.Temperament{
		domain.TempAgressive,
		domain.TempFearful,
		domain.TempCalculator,
		domain.TempStrategic,
		domain.TempUndecided,
		domain.TempIronic,
	}

	r.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})

	r.Shuffle(len(temperaments), func(i, j int) {
		temperaments[i], temperaments[j] = temperaments[j], temperaments[i]
	})

	var players []domain.Player

	for i, name := range iaNames {
		players = append(players, domain.Player{
			Name:        name,
			Role:        roles[i+1],
			Temperament: temperaments[i+1],
			IsAlive:     true,
		})
	}

	return &Game{Players: players, Turn: 1, SpeakingSystem: speakingSystem}
}

func (g *Game) DisplayVillage() {
	fmt.Printf("\n--- État du Village (Tour %d) ---\n", g.Turn)
	for p := range g.Players {
		status := "Vivant"
		if !g.Players[p].IsAlive {
			status = "Mort 💀"
		}
		fmt.Printf("- %s (%s) (%s) - %s\n", g.Players[p].Name, string(g.Players[p].Role), string(g.Players[p].Temperament), status)
	}
}

func (g *Game) Run(ctx context.Context) {
	chatChannel := make(chan string, len(g.Players))

	var aliveIaCount int

	situation := "Le village se réveille paisiblement."

	for _, p := range g.Players {
		if !p.IsAlive && p.Role == domain.RoleVillager {
			continue
		}

		aliveIaCount++

		playerToTalk := p

		go func() {
			reply, err := g.SpeakingSystem.Talk(ctx, &playerToTalk, situation)
			if err != nil {
				chatChannel <- fmt.Sprintf("[%s] : Error: %v", playerToTalk.Name, err)
			} else {
				chatChannel <- fmt.Sprintf("[%s] : %s", playerToTalk.Name, reply)
			}
		}()
	}

	for i := 0; i < aliveIaCount; i++ {
		<-chatChannel
	}
	fmt.Println("\n\n[Fin du débat du jour]")
	g.ExecuteVotes(ctx)
}

func (g *Game) ExecuteVotes(ctx context.Context) {
	votesTable := make(map[string]int)
	var suspects []string

	for i := range g.Players {
		if g.Players[i].IsAlive {
			suspects = append(suspects, g.Players[i].Name)
		}
	}

	for i, p := range g.Players {
		if !p.IsAlive {
			continue
		}

		var target string
		target = g.SpeakingSystem.ChooseWhoToVote(ctx, &g.Players[i], suspects)
		fmt.Printf("[VOTE] %s a voté contre %s\n", g.Players[i].Name, target)
		votesTable[target]++
	}
	var mostVoted string
	maxVotes := -1
	for name, count := range votesTable {
		if count > maxVotes {
			maxVotes = count
			mostVoted = name
		}
	}

	for i, p := range g.Players {
		if p.Name == mostVoted {
			g.Players[i].IsAlive = false
			fmt.Printf("\n💀 Le village a décidé d'éliminer %s ! Son rôle était : %s\n", p.Name, string(p.Role))
		}
	}
	g.Turn++
}
