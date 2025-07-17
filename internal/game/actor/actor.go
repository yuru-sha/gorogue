// Package actor プレイヤーとモンスターの管理を提供
// 戦闘可能なエンティティの基底クラスActorと、Player、Monster型を定義
package actor

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/entity"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// Actor represents a common base for all living entities that can fight
type Actor struct {
	*entity.Entity
	HP      int
	MaxHP   int
	Attack  int
	Defense int
}

// NewActor creates a new actor with the given stats
func NewActor(x, y int, symbol rune, color gruid.Color, hp, attack, defense int) *Actor {
	return &Actor{
		Entity:  entity.NewEntity(x, y, symbol, color),
		HP:      hp,
		MaxHP:   hp,
		Attack:  attack,
		Defense: defense,
	}
}

// IsAlive returns true if the actor is alive
func (a *Actor) IsAlive() bool {
	return a.HP > 0
}

// TakeDamage reduces the actor's HP by the given amount
func (a *Actor) TakeDamage(damage int) {
	oldHP := a.HP
	a.HP -= damage
	if a.HP < 0 {
		a.HP = 0
	}
	logger.Debug("Actor took damage",
		"damage", damage,
		"hp_before", oldHP,
		"hp_after", a.HP,
	)
	if a.HP == 0 {
		logger.Debug("Actor died")
	}
}

// Heal restores the actor's HP by the given amount, up to MaxHP
func (a *Actor) Heal(amount int) {
	oldHP := a.HP
	a.HP += amount
	if a.HP > a.MaxHP {
		a.HP = a.MaxHP
	}
	logger.Debug("Actor healed",
		"amount", amount,
		"hp_before", oldHP,
		"hp_after", a.HP,
	)
}

// CalculateDamage calculates damage dealt to a target based on this actor's attack and target's defense
func (a *Actor) CalculateDamage(targetDefense int) int {
	damage := a.Attack - targetDefense
	if damage < 1 {
		damage = 1 // Minimum 1 damage
	}
	return damage
}
