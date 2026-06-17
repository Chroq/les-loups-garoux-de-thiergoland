package domain

type Role uint8
type Temperament uint8

const (
	RoleWerewolf = iota
	RoleVillager

	TempAgressive = iota
	TempFearful
	TempCalculator
	TempUndecided
	TempStrategic
	TempIronic
)

type Player struct {
	Name        string
	Role        Role
	Temperament Temperament
	IsAlive     bool
}
