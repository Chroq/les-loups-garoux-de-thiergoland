package application

import (
	"context"
	"fmt"
	"log"
	"strings"
	"tiercelieux-llm-go/internal/domain"
	"tiercelieux-llm-go/internal/domain/repository"
	"time"
)

const DefaultPlayerCount = 6

type Engine struct {
	Game         *domain.Game
	PlayerSystem repository.PlayerSystem
}

func NewEngine(playerCount int, playerSystem repository.PlayerSystem) *Engine {
	if playerCount <= 0 {
		playerCount = DefaultPlayerCount
	}
	game := domain.NewGame(playerCount)
	return &Engine{Game: game, PlayerSystem: playerSystem}
}

func (e *Engine) Run(ctx context.Context) {
	for !e.IsGameOver() {
		fmt.Printf("\n=================== JOUR %d ===================\n", e.Game.Turn)

		lastNightAction := ""
		lastNightAction = e.ExecuteNightAction(ctx)

		if e.IsGameOver() {
			break
		}

		situation := fmt.Sprintf(
			`Le jour se lève sur le village. Hier, %s a été éliminé par les loups-garous. Débattez pour démasquer le coupable. Chaque joueur restant peut exprimer son 
			opinion en fonction de son tempérament. Les joueurs restants sont %v`, lastNightAction, e.Game.AllPlayers())
		debate := e.RunDebate(ctx, situation)
		e.ExecuteVotes(ctx, debate)

		time.Sleep(time.Second)
	}
}

func (e *Engine) ExecuteNightAction(ctx context.Context) string {
	fmt.Println("\n🌙 [Nuit] Les Loups-Garous se réveillent...")

	votesTable := make(map[string]int, len(e.Game.Werewolves))
	maxVotes := 0
	for _, p := range e.Game.Werewolves {
		victim := e.PlayerSystem.ChooseWhoToEat(ctx, &p, e.Game.Villagers)
		votesTable[victim.Name()]++
		if votesTable[victim.Name()] > maxVotes {
			maxVotes = votesTable[victim.Name()]
		}
	}

	// Get the first even if there is a tie
	var victim string
	for name, count := range votesTable {
		if count == maxVotes {
			victim = name
			break
		}
	}

	log.Default().Printf("Les loups-garous ont mangé %s\n", victim)
	e.Game.EliminatePlayer(victim)
	return victim
}

func (e *Engine) RunDebate(ctx context.Context, situation string) string {
	fmt.Println("\n💬 [Début du débat public] :")

	strBuilder := strings.Builder{}

	for _, p := range e.Game.AllPlayers() {
		strBuilder.WriteString(fmt.Sprintf("[%s] (%s) : ", p.Name, p.Temperament))
		reply, err := e.PlayerSystem.Talk(ctx, p, situation)
		if err != nil {
			fmt.Printf("Erreur: %v\n", err)
		} else {
			fmt.Println(reply)
		}
	}
	fmt.Println("\n[Fin du débat du jour]")

	return strBuilder.String()
}

func (e *Engine) ExecuteVotes(ctx context.Context, debate string) {
	fmt.Println("\n🗳️ [Phase de scrutin] :")

	votesTable := make(map[string]int)
	for _, p := range e.Game.AllPlayers() {
		target, err := e.PlayerSystem.ChooseWhoToVote(ctx, p, debate, e.Game.AllPlayersExcept(p.Name()))
		if err != nil {
			fmt.Printf("Erreur: %v\n", err)
		} else {
			fmt.Printf("- %s vote contre %s\n", p.Name(), target)
			votesTable[target]++
		}
	}

	var victim string
	maxVotes := -1
	for name, count := range votesTable {
		if count > maxVotes && name != "" {
			maxVotes = count
			victim = name
		}
	}

	// Application de la sentence
	if victim != "" {
		e.Game.EliminatePlayer(victim)
		fmt.Printf("\n💀 Le verdict est tombé : %s est éliminé.\n Il avait le rôle de %s\n", victim, e.Game.Deceased[victim].Role())
		log.Printf("Le village compte désormais %d villageois et %d loups-garous\n", len(e.Game.Villagers), len(e.Game.Werewolves))
	}
	e.Game.Turn++
}

func (e *Engine) IsGameOver() bool {
	if len(e.Game.Werewolves) == 0 {
		fmt.Println("\n🎉 VICTOIRE DU VILLAGE ! Le dernier Loup-Garou a été débusqué.")
		return true
	} else if len(e.Game.Villagers) == 0 {
		fmt.Println("\n🩸 VICTOIRE DES LOUPS-GAROUS ! Ils ont dévoré le village.")
		return true
	}
	return false
}
