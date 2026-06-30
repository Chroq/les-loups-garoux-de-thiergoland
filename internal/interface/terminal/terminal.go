package terminal

import (
	"fmt"
	"tiercelieux-llm-go/internal/domain"
)

type TerminalSystem struct{}

func NewTerminalSystem() *TerminalSystem {
	return &TerminalSystem{}
}

func (t *TerminalSystem) DisplaySummary(game *domain.Game) {
	t.DisplayAllRoles(game)
}

func (t *TerminalSystem) Display(content string) {
	fmt.Println(content)
}

func (t *TerminalSystem) DisplayAllRoles(game *domain.Game) {
	fmt.Println("\n--- RÔLES DES JOUEURS ---")
	for _, p := range game.AllPlayers() {
		var role string
		if p.Role() == domain.RoleWerewolf {
			role = "Loup-Garou 🐺"
		} else {
			role = "Villageois 👨‍🌾"
		}
		fmt.Printf("- %s: %s %s\n", p.Name(), p.Temperament(), role)
	}
	fmt.Println("--------------------------")
}

func (t *TerminalSystem) DisplayVillagerVictory() {
	fmt.Println("\n🎉 VICTOIRE DU VILLAGE ! Le dernier Loup-Garou a été débusqué.")
}

func (t *TerminalSystem) DisplayWerewolfVictory() {
	fmt.Println("\n🩸 VICTOIRE DES LOUPS-GAROUS ! Ils ont dévoré le village.")
}

func (t *TerminalSystem) DisplayVictim(victim string, role domain.Role) {
	fmt.Printf("\n💀 Le verdict est tombé : %s est éliminé.\n Il avait le rôle de %s\n", victim, role)
}

func (t *TerminalSystem) DisplayVote(voter string, target string) {
	fmt.Printf("- %s vote contre %s\n", voter, target)
}
