package actor

import (
	"github.com/yuru-sha/gorogue/internal/game/identification"
	"github.com/yuru-sha/gorogue/internal/game/inventory"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// Player represents the player character
type Player struct {
	*Actor
	Level       int
	Hunger      int
	Exp         int
	Gold        int
	Inventory   *inventory.Inventory
	Equipment   *inventory.Equipment
	IdentifyMgr *identification.IdentificationManager
}

// NewPlayer creates a new player at the given position
func NewPlayer(x, y int) *Player {
	player := &Player{
		Actor:       NewActor(x, y, '@', 0xFFFFFF, 20, 5, 2), // White color - オリジナルローグ風
		Level:       1,
		Hunger:      100,
		Exp:         0,
		Gold:        0,
		Inventory:   inventory.NewInventory(),
		Equipment:   inventory.NewEquipment(),
		IdentifyMgr: identification.NewIdentificationManager(),
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
	// Base attack + equipment bonus - enemy defense
	totalAttack := p.Attack + p.Equipment.GetAttackBonus()
	damage := totalAttack - targetDefense
	if damage < 1 {
		damage = 1
	}
	return damage
}

// GetTotalDefense returns total defense including equipment bonuses
func (p *Player) GetTotalDefense() int {
	return p.Defense + p.Equipment.GetDefenseBonus()
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
