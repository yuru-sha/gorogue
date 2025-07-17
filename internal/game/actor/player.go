package actor

import (
	"github.com/yuru-sha/gorogue/internal/core/entity"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// Player represents the player character
type Player struct {
	*entity.Entity
	Level     int
	HP        int
	MaxHP     int
	Attack    int
	Defense   int
	Hunger    int
	Exp       int
	Gold      int
	Inventory []interface{} // TODO: Replace with proper item types
}

// NewPlayer creates a new player at the given position
func NewPlayer(x, y int) *Player {
	player := &Player{
		Entity:    entity.NewEntity(x, y, '@', [3]uint8{255, 255, 255}), // White color
		Level:     1,
		HP:        20,
		MaxHP:     20,
		Attack:    5,
		Defense:   2,
		Hunger:    100,
		Exp:       0,
		Gold:      0,
		Inventory: make([]interface{}, 0),
	}
	logger.Debug("Created new player",
		"position_x", x,
		"position_y", y,
		"level", player.Level,
		"hp", player.HP,
		"attack", player.Attack,
		"defense", player.Defense,
	)
	return player
}

// TakeDamage applies damage to the player
func (p *Player) TakeDamage(damage int) {
	oldHP := p.HP
	p.HP -= damage
	if p.HP < 0 {
		p.HP = 0
	}
	logger.Debug("Player took damage",
		"damage", damage,
		"hp_before", oldHP,
		"hp_after", p.HP,
	)
	if p.HP == 0 {
		logger.Info("Player died")
	}
}

// Heal recovers the player's HP
func (p *Player) Heal(amount int) {
	oldHP := p.HP
	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	logger.Debug("Player healed",
		"amount", amount,
		"hp_before", oldHP,
		"hp_after", p.HP,
	)
}

// IsAlive checks if the player is still alive
func (p *Player) IsAlive() bool {
	return p.HP > 0
}

// AddGold adds gold to the player's inventory
func (p *Player) AddGold(amount int) {
	oldGold := p.Gold
	p.Gold += amount
	logger.Debug("Player collected gold",
		"amount", amount,
		"gold_before", oldGold,
		"gold_after", p.Gold,
	)
}

// AddExp adds experience points and handles level up
func (p *Player) AddExp(amount int) {
	oldExp := p.Exp
	oldLevel := p.Level
	p.Exp += amount
	// TODO: Implement level up logic based on original Rogue
	logger.Debug("Player gained experience",
		"amount", amount,
		"exp_before", oldExp,
		"exp_after", p.Exp,
		"level_before", oldLevel,
		"level_after", p.Level,
	)
}

// GainExp is an alias for AddExp
func (p *Player) GainExp(amount int) {
	p.AddExp(amount)
}

// CalculateDamage calculates damage dealt to a target
func (p *Player) CalculateDamage(targetDefense int) int {
	// 基本攻撃力から防御力を引く
	damage := p.Attack - targetDefense
	if damage < 1 {
		damage = 1
	}
	return damage
}

// UpdateHunger decreases hunger and handles starvation
func (p *Player) UpdateHunger() {
	oldHunger := p.Hunger
	p.Hunger--
	logger.Debug("Player hunger updated",
		"hunger_before", oldHunger,
		"hunger_after", p.Hunger,
	)
	if p.Hunger <= 0 {
		logger.Debug("Player is starving",
			"damage", 1,
			"hunger", p.Hunger,
		)
		p.TakeDamage(1) // Starvation damage
	}
}

// GetExpToNextLevel returns experience needed to reach next level
func (p *Player) GetExpToNextLevel() int {
	// Simple formula: level * 100 experience per level
	nextLevelExp := p.Level * 100
	return nextLevelExp - p.Exp
}
