// Package actor プレイヤーとモンスターの管理を提供
// 戦闘可能なエンティティの基底クラスActorと、Player、Monster型を定義
package actor

import (
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

// NewActor creates an actor with source combat stats.
func NewActor(x, y, hp, attack, defense int) *Actor {
	return &Actor{
		Entity:  entity.NewEntity(x, y),
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
