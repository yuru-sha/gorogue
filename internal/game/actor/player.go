package actor

import (
	"math/rand"
	"time"

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
	rng         *rand.Rand
}

var experienceThresholds = []int{0, 10, 20, 40, 80, 160, 320, 640, 1300, 2600, 5200, 10000, 20000, 40000, 80000}

func newRandomSource() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewPlayer creates a new player at the given position
func NewPlayer(x, y int) *Player {
	return NewPlayerWithRand(x, y, newRandomSource())
}

// NewPlayerWithSeed creates a player with deterministic item appearances and effects.
func NewPlayerWithSeed(x, y int, seed int64) *Player {
	return NewPlayerWithRand(x, y, rand.New(rand.NewSource(seed)))
}

// NewPlayerWithRand creates a player using the supplied random source.
func NewPlayerWithRand(x, y int, rng *rand.Rand) *Player {
	if rng == nil {
		rng = newRandomSource()
	}
	player := &Player{
		Actor:       NewActor(x, y, '@', 0xFFFFFF, 20, 5, 2), // White color - オリジナルローグ風
		Level:       1,
		Hunger:      100,
		Exp:         0,
		Gold:        0,
		Inventory:   inventory.NewInventory(),
		Equipment:   inventory.NewEquipment(),
		IdentifyMgr: identification.NewIdentificationManagerWithRand(rng),
		rng:         rng,
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

// SetRandomSource assigns the game's random source to the player and combat.
func (p *Player) SetRandomSource(rng *rand.Rand) {
	if rng != nil {
		p.rng = rng
	}
}

// RandomSource returns the game's random source for player effects.
func (p *Player) RandomSource() *rand.Rand {
	return p.random()
}

func (p *Player) random() *rand.Rand {
	if p.rng == nil {
		p.rng = newRandomSource()
	}
	return p.rng
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
	for p.Level < len(experienceThresholds) && p.Exp >= experienceThresholds[p.Level] {
		p.Level++
		p.MaxHP += 5
		p.HP = p.MaxHP
		p.Attack++
	}
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

// EatFood restores hunger without exceeding the full hunger value.
func (p *Player) EatFood(nutrition int) {
	p.Hunger += nutrition
	if p.Hunger > 100 {
		p.Hunger = 100
	}
}

// GetExpToNextLevel returns experience needed to reach next level
func (p *Player) GetExpToNextLevel() int {
	if p.Level >= len(experienceThresholds) {
		return 0
	}
	nextLevelExp := experienceThresholds[p.Level]
	return nextLevelExp - p.Exp
}
