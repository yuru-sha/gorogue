package actor

import (
	"math"
	"math/rand"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// MonsterType represents different types of monsters
type MonsterType struct {
	Symbol  rune
	Name    string
	HP      int
	Attack  int
	Defense int
	Color   gruid.Color
	Speed   int // Turn frequency (lower is faster)
}

// Predefined monster types - PyRogue風の色設定
var MonsterTypes = map[rune]MonsterType{
	'B': {Symbol: 'B', Name: "コウモリ", HP: 10, Attack: 3, Defense: 1, Color: 0x8B4513, Speed: 1},    // Brown
	'D': {Symbol: 'D', Name: "ドラゴン", HP: 100, Attack: 20, Defense: 10, Color: 0xFF0000, Speed: 3}, // Red
	'E': {Symbol: 'E', Name: "目玉", HP: 15, Attack: 5, Defense: 2, Color: 0x00FF00, Speed: 2},      // Green
	'F': {Symbol: 'F', Name: "ファンガス", HP: 8, Attack: 2, Defense: 1, Color: 0x90EE90, Speed: 4},    // LightGreen
	'G': {Symbol: 'G', Name: "ゴブリン", HP: 20, Attack: 6, Defense: 3, Color: 0x32CD32, Speed: 2},    // LimeGreen
	'O': {Symbol: 'O', Name: "オーク", HP: 25, Attack: 8, Defense: 4, Color: 0x696969, Speed: 2},     // DimGray
	'S': {Symbol: 'S', Name: "スケルトン", HP: 18, Attack: 7, Defense: 3, Color: 0xF5F5DC, Speed: 2},   // Beige
	'T': {Symbol: 'T', Name: "トロル", HP: 50, Attack: 12, Defense: 6, Color: 0x8B4513, Speed: 3},    // Brown
}

// Monster represents a monster in the game
type Monster struct {
	*Actor
	Type      MonsterType
	TurnCount int  // Turn management
	IsActive  bool // Active state
}

// NewMonster creates a new monster of the given type at the specified position
func NewMonster(x, y int, monsterType rune) *Monster {
	mType := MonsterTypes[monsterType]
	monster := &Monster{
		Actor:     NewActor(x, y, mType.Symbol, mType.Color, mType.HP, mType.Attack, mType.Defense),
		Type:      mType,
		TurnCount: 0,
		IsActive:  true,
	}

	logger.Debug("Created new monster",
		"type", mType.Name,
		"x", x,
		"y", y,
		"hp", mType.HP,
		"attack", mType.Attack,
		"defense", mType.Defense,
	)

	return monster
}

// LevelCollisionChecker is an interface for collision detection
type LevelCollisionChecker interface {
	IsInBounds(x, y int) bool
	IsWalkable(x, y int) bool
	GetMonsterAt(x, y int) *Monster
}

// Update handles monster AI logic
func (m *Monster) Update(player *Player, level LevelCollisionChecker) {
	if !m.IsActive || !m.IsAlive() {
		return
	}

	// Turn management
	m.TurnCount++
	if m.TurnCount < m.Type.Speed {
		return
	}
	m.TurnCount = 0

	// Calculate distance to player
	distance := m.DistanceToPlayer(player)

	// Attack if adjacent
	if distance <= 1.5 {
		m.AttackPlayer(player)
		return
	}

	// Move towards player
	m.MoveTowardsPlayer(player, level)
}

// DistanceToPlayer calculates the distance to the player
func (m *Monster) DistanceToPlayer(player *Player) float64 {
	dx := float64(m.Position.X - player.Position.X)
	dy := float64(m.Position.Y - player.Position.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// AttackPlayer performs an attack on the player
func (m *Monster) AttackPlayer(player *Player) {
	damage := m.CalculateDamage(player.Defense)
	player.TakeDamage(damage)

	logger.Info("Monster attacked player",
		"monster", m.Type.Name,
		"damage", damage,
		"player_hp", player.HP,
	)
}

// MoveTowardsPlayer moves the monster towards the player
func (m *Monster) MoveTowardsPlayer(player *Player, level LevelCollisionChecker) {
	dx := 0
	dy := 0

	if m.Position.X < player.Position.X {
		dx = 1
	} else if m.Position.X > player.Position.X {
		dx = -1
	}

	if m.Position.Y < player.Position.Y {
		dy = 1
	} else if m.Position.Y > player.Position.Y {
		dy = -1
	}

	// Add random element (25% chance to move in different direction)
	if rand.Float32() < 0.25 {
		directions := []struct{ dx, dy int }{
			{-1, -1}, {-1, 0}, {-1, 1},
			{0, -1}, {0, 1},
			{1, -1}, {1, 0}, {1, 1},
		}
		if len(directions) > 0 {
			dir := directions[rand.Intn(len(directions))]
			dx = dir.dx
			dy = dir.dy
		}
	}

	if dx != 0 || dy != 0 {
		// Calculate new position
		newX := m.Position.X + dx
		newY := m.Position.Y + dy

		// Check if movement is possible
		if m.CanMoveTo(newX, newY, level) {
			m.Position.Move(dx, dy)
			logger.Debug("Monster moved",
				"monster", m.Type.Name,
				"new_x", m.Position.X,
				"new_y", m.Position.Y,
			)
		} else {
			logger.Debug("Monster movement blocked",
				"monster", m.Type.Name,
				"blocked_x", newX,
				"blocked_y", newY,
			)
		}
	}
}

// CanMoveTo checks if the monster can move to the given position
func (m *Monster) CanMoveTo(x, y int, level LevelCollisionChecker) bool {
	// Boundary check
	if !level.IsInBounds(x, y) {
		return false
	}

	// Check if tile is walkable
	if !level.IsWalkable(x, y) {
		return false
	}

	// Check if another monster is present
	if level.GetMonsterAt(x, y) != nil {
		return false
	}

	return true
}
