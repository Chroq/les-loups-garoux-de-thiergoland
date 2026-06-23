package domain

type Role uint8
type Temperament uint8

type PlayerInterface interface {
	Name() string
	Role() Role
	Temperament() Temperament
}

const (
	RoleWerewolf Role = iota
	RoleVillager
)

const (
	TempAgressive Temperament = iota
	TempFearful
	TempCalculator
	TempUndecided
	TempStrategic
	TempIronic
)

func (r Role) String() string {
	switch r {
	case RoleWerewolf:
		return "Loup-Garou"
	case RoleVillager:
		return "Villageois"
	default:
		return "Inconnu"
	}
}

func (t Temperament) String() string {
	switch t {
	case TempAgressive:
		return "Agressif"
	case TempFearful:
		return "Peureux"
	case TempCalculator:
		return "Calculateur"
	case TempUndecided:
		return "Indécis"
	case TempStrategic:
		return "Stratégique"
	case TempIronic:
		return "Ironique"
	default:
		return "Inconnu"
	}
}

type Player struct {
	name        string
	temperament Temperament
}

type Werewolf struct {
	Player
}

type Villager struct {
	Player
}

func (p Player) Name() string {
	return p.name
}

func (p Player) Temperament() Temperament {
	return p.temperament
}

func (w Werewolf) Role() Role {
	return RoleWerewolf
}

func (v Villager) Role() Role {
	return RoleVillager
}
