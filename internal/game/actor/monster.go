package actor

import (
	"math"
	"math/rand"

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
	Color   [3]uint8
	Speed   int // ターン頻度（低いほど速い）
}

// Predefined monster types
var MonsterTypes = map[rune]MonsterType{
	'B': {Symbol: 'B', Name: "コウモリ", HP: 10, Attack: 3, Defense: 1, Color: [3]uint8{255, 0, 0}, Speed: 1},
	'D': {Symbol: 'D', Name: "ドラゴン", HP: 100, Attack: 20, Defense: 10, Color: [3]uint8{255, 0, 0}, Speed: 3},
	'E': {Symbol: 'E', Name: "目玉", HP: 15, Attack: 5, Defense: 2, Color: [3]uint8{255, 0, 0}, Speed: 2},
	'F': {Symbol: 'F', Name: "ファンガス", HP: 8, Attack: 2, Defense: 1, Color: [3]uint8{255, 0, 0}, Speed: 4},
	'G': {Symbol: 'G', Name: "ゴブリン", HP: 20, Attack: 6, Defense: 3, Color: [3]uint8{255, 0, 0}, Speed: 2},
	'O': {Symbol: 'O', Name: "オーク", HP: 25, Attack: 8, Defense: 4, Color: [3]uint8{255, 0, 0}, Speed: 2},
	'S': {Symbol: 'S', Name: "スケルトン", HP: 18, Attack: 7, Defense: 3, Color: [3]uint8{255, 255, 255}, Speed: 2},
	'T': {Symbol: 'T', Name: "トロル", HP: 50, Attack: 12, Defense: 6, Color: [3]uint8{0, 255, 0}, Speed: 3},
}

// Monster represents a monster in the game
type Monster struct {
	*entity.Entity
	Type      MonsterType
	HP        int
	MaxHP     int
	Attack    int
	Defense   int
	TurnCount int  // ターン管理用
	IsActive  bool // アクティブ状態
}

// NewMonster creates a new monster of the given type at the specified position
func NewMonster(x, y int, monsterType rune) *Monster {
	mType := MonsterTypes[monsterType]
	monster := &Monster{
		Entity:    entity.NewEntity(x, y, mType.Symbol, mType.Color),
		Type:      mType,
		HP:        mType.HP,
		MaxHP:     mType.HP,
		Attack:    mType.Attack,
		Defense:   mType.Defense,
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

// TakeDamage applies damage to the monster
func (m *Monster) TakeDamage(damage int) {
	m.HP -= damage
	if m.HP < 0 {
		m.HP = 0
	}
}

// IsAlive checks if the monster is still alive
func (m *Monster) IsAlive() bool {
	return m.HP > 0
}

// Attack calculates the damage dealt to a target
func (m *Monster) CalculateDamage(targetDefense int) int {
	// 基本的な攻撃力計算（オリジナルRogueに準拠する必要あり）
	damage := m.Attack - targetDefense
	if damage < 1 {
		damage = 1
	}
	return damage
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
	
	// ターン管理
	m.TurnCount++
	if m.TurnCount < m.Type.Speed {
		return
	}
	m.TurnCount = 0
	
	// プレイヤーとの距離計算
	distance := m.DistanceToPlayer(player)
	
	// 隣接している場合は攻撃
	if distance <= 1.5 {
		m.AttackPlayer(player)
		return
	}
	
	// プレイヤーに向かって移動
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
	
	// ランダムな要素を追加（25%の確率で異なる方向に移動）
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
		// 新しい位置を計算
		newX := m.Position.X + dx
		newY := m.Position.Y + dy
		
		// 移動可能かチェック
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
	// 境界チェック
	if !level.IsInBounds(x, y) {
		return false
	}
	
	// タイルが歩行可能かチェック
	if !level.IsWalkable(x, y) {
		return false
	}
	
	// 他のモンスターがいるかチェック
	if level.GetMonsterAt(x, y) != nil {
		return false
	}
	
	return true
}
