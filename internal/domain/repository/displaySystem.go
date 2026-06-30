package repository

import "tiercelieux-llm-go/internal/domain"

type DisplayMode uint8

const (
	DisplayModeConsole DisplayMode = iota
	DisplayModeWeb
)

type DisplaySystem interface {
	DisplaySummary(game *domain.Game)
	Display(content string)
	DisplayVillagerVictory()
	DisplayWerewolfVictory()
	DisplayVictim(victim string, role domain.Role)
	DisplayVote(voter string, target string)
}
