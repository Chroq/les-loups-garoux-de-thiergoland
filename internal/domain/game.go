package domain

import (
	"math/rand"
	"time"
)

type GameMode uint8

const (
	GameModeNone GameMode = iota
	GameModeLLM
	GameModeRandom
)

type GameState uint8

const (
	GameStateNight GameState = iota
	GameStateDay
)

type Game struct {
	Villagers  map[string]Villager
	Werewolves map[string]Werewolf
	Deceased   map[string]PlayerInterface
	Turn       int
	GameState  GameState
}

func NewGame(playerNumber int) *Game {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	playerNames := []string{"Hervé", "Bernard", "Chantale", "Josiane", "Christophe", "Jeanne"}
	numWerewolves := playerNumber / 3

	temperaments := []Temperament{
		TempAgressive,
		TempFearful,
		TempCalculator,
		TempStrategic,
		TempUndecided,
		TempIronic,
	}

	r.Shuffle(len(temperaments), func(i, j int) {
		temperaments[i], temperaments[j] = temperaments[j], temperaments[i]
	})

	werewolves := make(map[string]Werewolf, numWerewolves)
	villagers := make(map[string]Villager, playerNumber-numWerewolves)
	for i, name := range playerNames {
		if i < numWerewolves {
			werewolves[name] = Werewolf{
				Player: Player{
					name:        name,
					temperament: temperaments[i%len(temperaments)],
				},
			}
		} else {
			villagers[name] = Villager{
				Player: Player{
					name:        name,
					temperament: temperaments[i%len(temperaments)],
				},
			}
		}
	}

	return &Game{
		Villagers:  villagers,
		Werewolves: werewolves,
		Deceased:   make(map[string]PlayerInterface, playerNumber),
		Turn:       1,
		GameState:  GameStateDay,
	}
}

func (g *Game) EliminatePlayer(name string) {
	for _, p := range g.Villagers {
		if p.name == name {
			delete(g.Villagers, name)
			g.Deceased[name] = p
			return
		}
	}
	for _, p := range g.Werewolves {
		if p.name == name {
			delete(g.Werewolves, name)
			g.Deceased[name] = p
			return
		}
	}
}

func (g *Game) AllPlayers() []PlayerInterface {
	players := make([]PlayerInterface, 0, len(g.Villagers)+len(g.Werewolves))
	for i := range g.Villagers {
		players = append(players, g.Villagers[i])
	}
	for i := range g.Werewolves {
		players = append(players, g.Werewolves[i])
	}
	return players
}

func (g *Game) AllPlayersExcept(excepted string) []PlayerInterface {
	players := make([]PlayerInterface, 0, len(g.Villagers)+len(g.Werewolves))
	for i := range g.Villagers {
		if g.Villagers[i].name != excepted {
			players = append(players, g.Villagers[i])
		}
	}
	for i := range g.Werewolves {
		if g.Werewolves[i].name != excepted {
			players = append(players, g.Werewolves[i])
		}
	}
	return players
}
