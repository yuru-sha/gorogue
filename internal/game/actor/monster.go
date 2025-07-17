package actor

import (
	"math"
	"math/rand"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/entity"
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

// Predefined monster types - PyRogue風の色設定 (A-Z全26種類)
var MonsterTypes = map[rune]MonsterType{
	'A': {Symbol: 'A', Name: "アント", HP: 12, Attack: 4, Defense: 2, Color: 0x800000, Speed: 1},        // Maroon
	'B': {Symbol: 'B', Name: "コウモリ", HP: 10, Attack: 3, Defense: 1, Color: 0x8B4513, Speed: 1},       // Brown
	'C': {Symbol: 'C', Name: "ケンタウロス", HP: 35, Attack: 10, Defense: 5, Color: 0xCD853F, Speed: 2},   // Peru
	'D': {Symbol: 'D', Name: "ドラゴン", HP: 100, Attack: 20, Defense: 10, Color: 0xFF0000, Speed: 3},   // Red
	'E': {Symbol: 'E', Name: "目玉", HP: 15, Attack: 5, Defense: 2, Color: 0x00FF00, Speed: 2},         // Green
	'F': {Symbol: 'F', Name: "ファンガス", HP: 8, Attack: 2, Defense: 1, Color: 0x90EE90, Speed: 4},     // LightGreen
	'G': {Symbol: 'G', Name: "ゴブリン", HP: 20, Attack: 6, Defense: 3, Color: 0x32CD32, Speed: 2},     // LimeGreen
	'H': {Symbol: 'H', Name: "ホブゴブリン", HP: 30, Attack: 8, Defense: 4, Color: 0xFF8C00, Speed: 2},   // DarkOrange
	'I': {Symbol: 'I', Name: "インプ", HP: 18, Attack: 7, Defense: 3, Color: 0xFF1493, Speed: 2},       // DeepPink
	'J': {Symbol: 'J', Name: "ジェリー", HP: 25, Attack: 6, Defense: 2, Color: 0x40E0D0, Speed: 3},     // Turquoise
	'K': {Symbol: 'K', Name: "コボルト", HP: 14, Attack: 5, Defense: 2, Color: 0x8B008B, Speed: 2},     // DarkMagenta
	'L': {Symbol: 'L', Name: "レプラコーン", HP: 22, Attack: 6, Defense: 3, Color: 0x9ACD32, Speed: 1},  // YellowGreen
	'M': {Symbol: 'M', Name: "ミノタウロス", HP: 60, Attack: 15, Defense: 8, Color: 0xA0522D, Speed: 3},  // Sienna
	'N': {Symbol: 'N', Name: "ニンフ", HP: 16, Attack: 4, Defense: 2, Color: 0x98FB98, Speed: 2},       // PaleGreen
	'O': {Symbol: 'O', Name: "オーク", HP: 25, Attack: 8, Defense: 4, Color: 0x696969, Speed: 2},       // DimGray
	'P': {Symbol: 'P', Name: "ファントム", HP: 40, Attack: 12, Defense: 6, Color: 0x778899, Speed: 3},   // LightSlateGray
	'Q': {Symbol: 'Q', Name: "クエーサー", HP: 80, Attack: 18, Defense: 9, Color: 0x4B0082, Speed: 4},   // Indigo
	'R': {Symbol: 'R', Name: "ラットルスネーク", HP: 28, Attack: 9, Defense: 4, Color: 0x9932CC, Speed: 2}, // DarkOrchid
	'S': {Symbol: 'S', Name: "スケルトン", HP: 18, Attack: 7, Defense: 3, Color: 0xF5F5DC, Speed: 2},    // Beige
	'T': {Symbol: 'T', Name: "トロル", HP: 50, Attack: 12, Defense: 6, Color: 0x8B4513, Speed: 3},     // Brown
	'U': {Symbol: 'U', Name: "アンバーハルク", HP: 45, Attack: 11, Defense: 7, Color: 0xFFD700, Speed: 3}, // Gold
	'V': {Symbol: 'V', Name: "バンパイア", HP: 55, Attack: 13, Defense: 7, Color: 0x8B0000, Speed: 3},   // DarkRed
	'W': {Symbol: 'W', Name: "ワイト", HP: 32, Attack: 9, Defense: 5, Color: 0xF0E68C, Speed: 2},       // Khaki
	'X': {Symbol: 'X', Name: "ゼロックス", HP: 70, Attack: 16, Defense: 8, Color: 0x00CED1, Speed: 4},   // DarkTurquoise
	'Y': {Symbol: 'Y', Name: "イエティ", HP: 65, Attack: 14, Defense: 7, Color: 0xF0F8FF, Speed: 3},     // AliceBlue
	'Z': {Symbol: 'Z', Name: "ゾンビ", HP: 35, Attack: 10, Defense: 5, Color: 0x556B2F, Speed: 4},      // DarkOliveGreen
}

// AIState represents the current AI state of a monster
type AIState int

const (
	StateIdle AIState = iota
	StatePatrol
	StateChase
	StateAttack
	StateSearch
	StateFlee
)

// Monster represents a monster in the game
type Monster struct {
	*Actor
	Type            MonsterType
	TurnCount       int     // Turn management
	IsActive        bool    // Active state
	AIState         AIState // Current AI state
	LastPlayerPos   entity.Position // Last known player position
	PatrolPath      []entity.Position // Patrol path
	PatrolIndex     int     // Current patrol index
	AlertLevel      int     // How alert the monster is (0-10)
	SearchTurns     int     // Turns spent searching
	OriginalPos     entity.Position // Starting position for patrol
	ViewRange       int     // How far the monster can see
	DetectionRange  int     // How close player must be to detect
}

// NewMonster creates a new monster of the given type at the specified position
func NewMonster(x, y int, monsterType rune) *Monster {
	mType := MonsterTypes[monsterType]
	monster := &Monster{
		Actor:          NewActor(x, y, mType.Symbol, mType.Color, mType.HP, mType.Attack, mType.Defense),
		Type:           mType,
		TurnCount:      0,
		IsActive:       true,
		AIState:        StateIdle,
		LastPlayerPos:  entity.Position{X: -1, Y: -1},
		PatrolPath:     make([]entity.Position, 0),
		PatrolIndex:    0,
		AlertLevel:     0,
		SearchTurns:    0,
		OriginalPos:    entity.Position{X: x, Y: y},
		ViewRange:      calculateViewRange(monsterType),
		DetectionRange: calculateDetectionRange(monsterType),
	}

	// Initialize patrol path
	monster.generatePatrolPath()

	logger.Debug("Created new monster",
		"type", mType.Name,
		"x", x,
		"y", y,
		"hp", mType.HP,
		"attack", mType.Attack,
		"defense", mType.Defense,
		"view_range", monster.ViewRange,
		"detection_range", monster.DetectionRange,
	)

	return monster
}

// calculateViewRange calculates the view range for a monster type
func calculateViewRange(monsterType rune) int {
	switch monsterType {
	case 'E': // 目玉 - 遠視
		return 10
	case 'D': // ドラゴン - 遠視
		return 8
	case 'P': // ファントム - 遠視
		return 7
	case 'B': // コウモリ - 近視
		return 3
	case 'F': // ファンガス - 近視
		return 2
	default:
		return 5 // 標準的な視界
	}
}

// calculateDetectionRange calculates the detection range for a monster type
func calculateDetectionRange(monsterType rune) int {
	switch monsterType {
	case 'L': // レプラコーン - 高い感知能力
		return 6
	case 'N': // ニンフ - 高い感知能力
		return 6
	case 'V': // バンパイア - 高い感知能力
		return 7
	case 'F': // ファンガス - 低い感知能力
		return 1
	case 'Z': // ゾンビ - 低い感知能力
		return 2
	default:
		return 4 // 標準的な検知範囲
	}
}

// generatePatrolPath generates a simple patrol path for the monster
func (m *Monster) generatePatrolPath() {
	// Simple 4-point patrol pattern around the original position
	m.PatrolPath = []entity.Position{
		{X: m.OriginalPos.X, Y: m.OriginalPos.Y},
		{X: m.OriginalPos.X + 2, Y: m.OriginalPos.Y},
		{X: m.OriginalPos.X + 2, Y: m.OriginalPos.Y + 2},
		{X: m.OriginalPos.X, Y: m.OriginalPos.Y + 2},
	}
}

// LevelCollisionChecker is an interface for collision detection
type LevelCollisionChecker interface {
	IsInBounds(x, y int) bool
	IsWalkable(x, y int) bool
	GetMonsterAt(x, y int) *Monster
}

// Update handles monster AI logic with advanced behavior patterns
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

	// Check player visibility and distance
	playerDistance := m.DistanceToPlayer(player)
	canSeePlayer := m.CanSeePlayer(player, level)
	
	// Update AI state based on player detection
	m.UpdateAIState(player, level, canSeePlayer, playerDistance)

	// Execute behavior based on current AI state
	switch m.AIState {
	case StateIdle:
		m.behaviorIdle(player, level)
	case StatePatrol:
		m.behaviorPatrol(player, level)
	case StateChase:
		m.behaviorChase(player, level)
	case StateAttack:
		m.behaviorAttack(player, level)
	case StateSearch:
		m.behaviorSearch(player, level)
	case StateFlee:
		m.behaviorFlee(player, level)
	}

	// Update alert level decay
	m.updateAlertLevel()
}

// DistanceToPlayer calculates the distance to the player
func (m *Monster) DistanceToPlayer(player *Player) float64 {
	dx := float64(m.Position.X - player.Position.X)
	dy := float64(m.Position.Y - player.Position.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// AttackPlayer performs an attack on the player with enhanced combat mechanics
func (m *Monster) AttackPlayer(player *Player) {
	// Calculate hit chance based on monster type and player defense
	hitChance := m.calculateHitChance(player)
	
	// Roll for hit
	if rand.Float64() > hitChance {
		logger.Info("Monster attack missed",
			"monster", m.Type.Name,
			"hit_chance", hitChance,
		)
		return
	}

	// Calculate damage
	baseDamage := m.CalculateDamage(player.GetTotalDefense())
	
	// Apply monster-specific damage modifiers
	finalDamage := m.applyDamageModifiers(baseDamage, player)
	
	// Apply damage to player
	player.TakeDamage(finalDamage)

	// Apply special effects
	m.applySpecialEffects(player)

	logger.Info("Monster attacked player",
		"monster", m.Type.Name,
		"base_damage", baseDamage,
		"final_damage", finalDamage,
		"player_hp", player.HP,
		"hit_chance", hitChance,
	)
}

// calculateHitChance calculates the chance for the monster to hit the player
func (m *Monster) calculateHitChance(player *Player) float64 {
	// Base hit chance is 0.8 (80%)
	baseHitChance := 0.8
	
	// Modify based on monster type
	switch m.Type.Symbol {
	case 'A', 'B': // Fast monsters have higher hit chance
		baseHitChance += 0.1
	case 'E': // Eye has perfect accuracy
		baseHitChance = 1.0
	case 'F', 'Z': // Slow monsters have lower hit chance
		baseHitChance -= 0.1
	}
	
	// Modify based on player level (higher level = harder to hit)
	levelMod := float64(player.Level) * 0.02
	if levelMod > 0.2 {
		levelMod = 0.2
	}
	baseHitChance -= levelMod
	
	// Ensure hit chance is between 0.1 and 1.0
	if baseHitChance < 0.1 {
		baseHitChance = 0.1
	}
	if baseHitChance > 1.0 {
		baseHitChance = 1.0
	}
	
	return baseHitChance
}

// applyDamageModifiers applies monster-specific damage modifiers
func (m *Monster) applyDamageModifiers(baseDamage int, player *Player) int {
	finalDamage := baseDamage
	
	// Apply monster-specific damage modifiers
	switch m.Type.Symbol {
	case 'D': // Dragons do extra fire damage
		finalDamage += rand.Intn(5) + 1
	case 'V': // Vampires do life drain
		finalDamage += rand.Intn(3) + 1
		if m.HP < m.MaxHP {
			healAmount := finalDamage / 4
			m.Heal(healAmount)
		}
	case 'T': // Trolls do crushing damage
		finalDamage += rand.Intn(4) + 1
	case 'P': // Phantoms do psychic damage
		finalDamage += rand.Intn(3) + 1
	case 'R': // Rattlesnakes do poison damage
		finalDamage += rand.Intn(2) + 1
	}
	
	// Random damage variation (±25%)
	variation := float64(finalDamage) * 0.25
	modifier := (rand.Float64() - 0.5) * variation
	finalDamage += int(modifier)
	
	// Minimum damage is 1
	if finalDamage < 1 {
		finalDamage = 1
	}
	
	return finalDamage
}

// applySpecialEffects applies special combat effects
func (m *Monster) applySpecialEffects(player *Player) {
	switch m.Type.Symbol {
	case 'R': // Rattlesnake poison
		if rand.Float64() < 0.2 { // 20% chance
			logger.Info("Player poisoned by rattlesnake",
				"monster", m.Type.Name,
			)
			// TODO: Implement poison effect
		}
	case 'V': // Vampire level drain
		if rand.Float64() < 0.1 { // 10% chance
			logger.Info("Player drained by vampire",
				"monster", m.Type.Name,
			)
			// TODO: Implement level drain
		}
	case 'L': // Leprechaun steals gold
		if rand.Float64() < 0.15 && player.Gold > 0 { // 15% chance
			stolen := rand.Intn(player.Gold/4 + 1)
			if stolen > 0 {
				player.Gold -= stolen
				logger.Info("Leprechaun stole gold",
					"monster", m.Type.Name,
					"amount", stolen,
				)
			}
		}
	case 'N': // Nymph steals items
		if rand.Float64() < 0.1 { // 10% chance
			logger.Info("Nymph attempts to steal item",
				"monster", m.Type.Name,
			)
			// TODO: Implement item stealing
		}
	}
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

// CanSeePlayer checks if the monster can see the player using line of sight
func (m *Monster) CanSeePlayer(player *Player, level LevelCollisionChecker) bool {
	distance := m.DistanceToPlayer(player)
	
	// Check if player is within view range
	if distance > float64(m.ViewRange) {
		return false
	}

	// Check line of sight using bresenham algorithm
	return m.hasLineOfSight(player.Position.X, player.Position.Y, level)
}

// hasLineOfSight checks if there's a clear line of sight to the target position
func (m *Monster) hasLineOfSight(targetX, targetY int, level LevelCollisionChecker) bool {
	x0, y0 := m.Position.X, m.Position.Y
	x1, y1 := targetX, targetY

	// Bresenham's line algorithm
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	sy := 1
	
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	x, y := x0, y0

	for {
		// Don't check the monster's own position
		if !(x == x0 && y == y0) {
			if !level.IsInBounds(x, y) || !level.IsWalkable(x, y) {
				return false
			}
		}

		// Reached target
		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	return true
}

// UpdateAIState updates the monster's AI state based on player detection
func (m *Monster) UpdateAIState(player *Player, level LevelCollisionChecker, canSeePlayer bool, distance float64) {
	oldState := m.AIState

	if canSeePlayer {
		// Player is visible
		m.LastPlayerPos = entity.Position{X: player.Position.X, Y: player.Position.Y}
		m.AlertLevel = 10
		m.SearchTurns = 0

		if distance <= 1.5 {
			m.AIState = StateAttack
		} else if m.HP < m.MaxHP/3 && m.Type.Symbol != 'D' && m.Type.Symbol != 'T' {
			// Weak monsters flee when low on health (except dragons and trolls)
			m.AIState = StateFlee
		} else {
			m.AIState = StateChase
		}
	} else if distance <= float64(m.DetectionRange) {
		// Player is close but not visible
		if m.AlertLevel < 5 {
			m.AlertLevel += 2
		}
		if m.AlertLevel >= 5 && m.LastPlayerPos.X != -1 {
			m.AIState = StateSearch
		}
	} else {
		// Player is not detected
		if m.AlertLevel > 0 && m.LastPlayerPos.X != -1 {
			m.AIState = StateSearch
		} else if m.AlertLevel == 0 {
			// Return to patrol or idle
			if len(m.PatrolPath) > 0 {
				m.AIState = StatePatrol
			} else {
				m.AIState = StateIdle
			}
		}
	}

	if oldState != m.AIState {
		logger.Debug("Monster AI state changed",
			"monster", m.Type.Name,
			"old_state", oldState,
			"new_state", m.AIState,
			"alert_level", m.AlertLevel,
		)
	}
}

// behaviorIdle handles idle behavior
func (m *Monster) behaviorIdle(player *Player, level LevelCollisionChecker) {
	// 25% chance to move randomly
	if rand.Float32() < 0.25 {
		m.moveRandomly(level)
	}
}

// behaviorPatrol handles patrol behavior
func (m *Monster) behaviorPatrol(player *Player, level LevelCollisionChecker) {
	if len(m.PatrolPath) == 0 {
		m.behaviorIdle(player, level)
		return
	}

	target := m.PatrolPath[m.PatrolIndex]
	
	// Check if we've reached the patrol point
	if m.Position.X == target.X && m.Position.Y == target.Y {
		m.PatrolIndex = (m.PatrolIndex + 1) % len(m.PatrolPath)
		target = m.PatrolPath[m.PatrolIndex]
	}

	// Move towards the patrol point
	m.moveTowards(target.X, target.Y, level)
}

// behaviorChase handles chase behavior with pathfinding
func (m *Monster) behaviorChase(player *Player, level LevelCollisionChecker) {
	// Use A* pathfinding for intelligent monsters
	if m.isIntelligent() {
		path := m.FindPathToPlayer(player, level)
		if path != nil && len(path) > 1 {
			if m.MoveAlongPath(path, level) {
				return
			}
		}
	}
	
	// Fallback to simple movement
	m.MoveTowardsPlayer(player, level)
}

// isIntelligent checks if the monster should use advanced pathfinding
func (m *Monster) isIntelligent() bool {
	// Smart monsters use A* pathfinding
	switch m.Type.Symbol {
	case 'C', 'D', 'M', 'P', 'Q', 'V', 'X', 'Y': // Centaur, Dragon, Minotaur, Phantom, Quasar, Vampire, Xerocs, Yeti
		return true
	default:
		return false
	}
}

// behaviorAttack handles attack behavior
func (m *Monster) behaviorAttack(player *Player, level LevelCollisionChecker) {
	m.AttackPlayer(player)
}

// behaviorSearch handles search behavior
func (m *Monster) behaviorSearch(player *Player, level LevelCollisionChecker) {
	m.SearchTurns++
	
	// Search for a limited number of turns
	if m.SearchTurns > 10 {
		m.AlertLevel = 0
		m.SearchTurns = 0
		m.LastPlayerPos = entity.Position{X: -1, Y: -1}
		m.AIState = StateIdle
		return
	}

	// Move towards last known player position
	if m.LastPlayerPos.X != -1 {
		m.moveTowards(m.LastPlayerPos.X, m.LastPlayerPos.Y, level)
	} else {
		m.moveRandomly(level)
	}
}

// behaviorFlee handles flee behavior
func (m *Monster) behaviorFlee(player *Player, level LevelCollisionChecker) {
	// Move away from player
	dx := m.Position.X - player.Position.X
	dy := m.Position.Y - player.Position.Y
	
	// Normalize direction
	if dx > 0 {
		dx = 1
	} else if dx < 0 {
		dx = -1
	}
	if dy > 0 {
		dy = 1
	} else if dy < 0 {
		dy = -1
	}

	newX := m.Position.X + dx
	newY := m.Position.Y + dy

	if m.CanMoveTo(newX, newY, level) {
		m.Position.Move(dx, dy)
	} else {
		m.moveRandomly(level)
	}
}

// moveTowards moves the monster towards a target position
func (m *Monster) moveTowards(targetX, targetY int, level LevelCollisionChecker) {
	dx := 0
	dy := 0

	if m.Position.X < targetX {
		dx = 1
	} else if m.Position.X > targetX {
		dx = -1
	}

	if m.Position.Y < targetY {
		dy = 1
	} else if m.Position.Y > targetY {
		dy = -1
	}

	// Try to move diagonally first
	if dx != 0 && dy != 0 {
		if m.CanMoveTo(m.Position.X+dx, m.Position.Y+dy, level) {
			m.Position.Move(dx, dy)
			return
		}
	}

	// Try horizontal movement
	if dx != 0 && m.CanMoveTo(m.Position.X+dx, m.Position.Y, level) {
		m.Position.Move(dx, 0)
		return
	}

	// Try vertical movement
	if dy != 0 && m.CanMoveTo(m.Position.X, m.Position.Y+dy, level) {
		m.Position.Move(0, dy)
		return
	}

	// If all else fails, try random movement
	m.moveRandomly(level)
}

// moveRandomly moves the monster in a random direction
func (m *Monster) moveRandomly(level LevelCollisionChecker) {
	directions := []struct{ dx, dy int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, dir := range directions {
		if m.CanMoveTo(m.Position.X+dir.dx, m.Position.Y+dir.dy, level) {
			m.Position.Move(dir.dx, dir.dy)
			return
		}
	}
}

// updateAlertLevel updates the monster's alert level
func (m *Monster) updateAlertLevel() {
	if m.AlertLevel > 0 {
		m.AlertLevel--
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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
