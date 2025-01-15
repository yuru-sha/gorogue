package actor

import (
	"github.com/yuru-sha/gorogue/internal/core/entity"
)

// MonsterType represents different types of monsters
type MonsterType struct {
	Symbol  rune
	Name    string
	HP      int
	Attack  int
	Defense int
	Color   [3]uint8
}

// Predefined monster types
var MonsterTypes = map[rune]MonsterType{
	'B': {Symbol: 'B', Name: "コウモリ", HP: 10, Attack: 3, Defense: 1, Color: [3]uint8{255, 0, 0}},
	'D': {Symbol: 'D', Name: "ドラゴン", HP: 100, Attack: 20, Defense: 10, Color: [3]uint8{255, 0, 0}},
	'E': {Symbol: 'E', Name: "目玉", HP: 15, Attack: 5, Defense: 2, Color: [3]uint8{255, 0, 0}},
	'F': {Symbol: 'F', Name: "ファンガス", HP: 8, Attack: 2, Defense: 1, Color: [3]uint8{255, 0, 0}},
	'G': {Symbol: 'G', Name: "ゴブリン", HP: 20, Attack: 6, Defense: 3, Color: [3]uint8{255, 0, 0}},
	// ... 他のモンスタータイプも同様に定義
}

// Monster represents a monster in the game
type Monster struct {
	*entity.Entity
	Type    MonsterType
	HP      int
	MaxHP   int
	Attack  int
	Defense int
}

// NewMonster creates a new monster of the given type at the specified position
func NewMonster(x, y int, monsterType rune) *Monster {
	mType := MonsterTypes[monsterType]
	return &Monster{
		Entity:  entity.NewEntity(x, y, mType.Symbol, mType.Color),
		Type:    mType,
		HP:      mType.HP,
		MaxHP:   mType.HP,
		Attack:  mType.Attack,
		Defense: mType.Defense,
	}
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
