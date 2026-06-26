package websocket

import "tiercelieux-llm-go/internal/domain"

type WSSystem struct{}

func NewWSSystem() *WSSystem {
	return &WSSystem{}
}

func (w *WSSystem) DisplaySummary(game *domain.Game) {
	// TODO: Implement WebSocket display
}

func (w *WSSystem) Display(content string) {
	// TODO: Implement WebSocket display
}

func (w *WSSystem) DisplayVillagerVictory() {
	// TODO: Implement WebSocket display
}

func (w *WSSystem) DisplayWerewolfVictory() {
	// TODO: Implement WebSocket display
}

func (w *WSSystem) DisplayVictim(victim string, role domain.Role) {
	// TODO: Implement WebSocket display
}
