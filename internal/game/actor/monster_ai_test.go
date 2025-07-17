package actor

import (
	"fmt"
	"testing"

	"github.com/yuru-sha/gorogue/internal/core/entity"
)

// MockLevelCollisionChecker is a mock implementation for testing
type MockLevelCollisionChecker struct {
	width, height int
	walkable      map[string]bool
	monsters      map[string]*Monster
}

func NewMockLevelCollisionChecker(width, height int) *MockLevelCollisionChecker {
	return &MockLevelCollisionChecker{
		width:    width,
		height:   height,
		walkable: make(map[string]bool),
		monsters: make(map[string]*Monster),
	}
}

func (m *MockLevelCollisionChecker) IsInBounds(x, y int) bool {
	return x >= 0 && x < m.width && y >= 0 && y < m.height
}

func (m *MockLevelCollisionChecker) IsWalkable(x, y int) bool {
	key := m.key(x, y)
	if walkable, exists := m.walkable[key]; exists {
		return walkable
	}
	return true // Default to walkable
}

func (m *MockLevelCollisionChecker) GetMonsterAt(x, y int) *Monster {
	key := m.key(x, y)
	return m.monsters[key]
}

func (m *MockLevelCollisionChecker) SetWalkable(x, y int, walkable bool) {
	m.walkable[m.key(x, y)] = walkable
}

func (m *MockLevelCollisionChecker) PlaceMonster(x, y int, monster *Monster) {
	m.monsters[m.key(x, y)] = monster
}

func (m *MockLevelCollisionChecker) key(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func TestMonsterCreation(t *testing.T) {
	// Test all A-Z monsters
	monsterTypes := []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

	for _, monsterType := range monsterTypes {
		t.Run(string(monsterType), func(t *testing.T) {
			monster := NewMonster(5, 5, monsterType)

			if monster == nil {
				t.Fatalf("Failed to create monster type %c", monsterType)
			}

			if monster.Type.Symbol != monsterType {
				t.Errorf("Expected symbol %c, got %c", monsterType, monster.Type.Symbol)
			}

			if monster.Position.X != 5 || monster.Position.Y != 5 {
				t.Errorf("Expected position (5,5), got (%d,%d)", monster.Position.X, monster.Position.Y)
			}

			if monster.AIState != StateIdle {
				t.Errorf("Expected initial AI state to be StateIdle, got %v", monster.AIState)
			}

			if monster.ViewRange <= 0 {
				t.Errorf("Expected positive view range, got %d", monster.ViewRange)
			}

			if monster.DetectionRange <= 0 {
				t.Errorf("Expected positive detection range, got %d", monster.DetectionRange)
			}
		})
	}
}

func TestMonsterLineOfSight(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	monster := NewMonster(1, 1, 'G') // Goblin

	// Test clear line of sight
	if !monster.hasLineOfSight(3, 3, level) {
		t.Error("Expected clear line of sight to (3,3)")
	}

	// Test blocked line of sight
	level.SetWalkable(2, 2, false)
	if monster.hasLineOfSight(3, 3, level) {
		t.Error("Expected blocked line of sight to (3,3)")
	}

	// Test line of sight to same position
	if !monster.hasLineOfSight(1, 1, level) {
		t.Error("Expected line of sight to same position")
	}
}

func TestMonsterAIStateMachine(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	monster := NewMonster(5, 5, 'G') // Goblin
	player := NewPlayer(1, 1)

	// Test initial state
	if monster.AIState != StateIdle {
		t.Errorf("Expected initial state to be StateIdle, got %v", monster.AIState)
	}

	// Test state change when player is far
	monster.UpdateAIState(player, level, false, 10.0)
	if monster.AIState != StateIdle && monster.AIState != StatePatrol {
		t.Errorf("Expected state to be StateIdle or StatePatrol when player is far, got %v", monster.AIState)
	}

	// Test state change when player is visible
	monster.UpdateAIState(player, level, true, 3.0)
	if monster.AIState != StateChase {
		t.Errorf("Expected state to be StateChase when player is visible, got %v", monster.AIState)
	}

	// Test state change when player is adjacent
	monster.UpdateAIState(player, level, true, 1.0)
	if monster.AIState != StateAttack {
		t.Errorf("Expected state to be StateAttack when player is adjacent, got %v", monster.AIState)
	}
}

func TestMonsterMovement(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	monster := NewMonster(5, 5, 'G') // Goblin

	// Test valid movement
	if !monster.CanMoveTo(6, 6, level) {
		t.Error("Expected to be able to move to (6,6)")
	}

	// Test blocked movement
	level.SetWalkable(6, 6, false)
	if monster.CanMoveTo(6, 6, level) {
		t.Error("Expected to be unable to move to blocked tile")
	}

	// Test out of bounds movement
	if monster.CanMoveTo(-1, 5, level) {
		t.Error("Expected to be unable to move out of bounds")
	}

	// Test movement blocked by another monster
	level.SetWalkable(6, 6, true)
	otherMonster := NewMonster(6, 6, 'O') // Orc
	level.PlaceMonster(6, 6, otherMonster)
	if monster.CanMoveTo(6, 6, level) {
		t.Error("Expected to be unable to move to tile occupied by another monster")
	}
}

func TestMonsterPatrolBehavior(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	monster := NewMonster(5, 5, 'G') // Goblin
	player := NewPlayer(1, 1)

	// Set monster to patrol state
	monster.AIState = StatePatrol

	// Test patrol behavior
	monster.behaviorPatrol(player, level)

	// Monster should attempt to move towards patrol point
	if len(monster.PatrolPath) == 0 {
		t.Error("Expected patrol path to be generated")
	}
}

func TestMonsterCombat(t *testing.T) {
	monster := NewMonster(5, 5, 'G') // Goblin
	player := NewPlayer(1, 1)

	// Test hit chance calculation
	hitChance := monster.calculateHitChance(player)
	if hitChance < 0.1 || hitChance > 1.0 {
		t.Errorf("Expected hit chance between 0.1 and 1.0, got %f", hitChance)
	}

	// Test damage calculation
	baseDamage := monster.CalculateDamage(player.GetTotalDefense())
	if baseDamage < 1 {
		t.Errorf("Expected base damage >= 1, got %d", baseDamage)
	}

	// Test special abilities for specific monsters
	dragon := NewMonster(5, 5, 'D')
	dragonBaseDamage := dragon.CalculateDamage(player.GetTotalDefense())
	dragonFinalDamage := dragon.applyDamageModifiers(dragonBaseDamage, player)
	if dragonFinalDamage <= dragonBaseDamage {
		t.Error("Expected dragon to have damage bonus")
	}
}

func TestMonsterPathfinding(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	monster := NewMonster(1, 1, 'D') // Dragon (intelligent)
	player := NewPlayer(9, 9)

	// Test A* pathfinding
	path := monster.FindPathToPlayer(player, level)
	if path == nil {
		t.Error("Expected path to be found")
		return
	}

	if len(path) == 0 {
		t.Error("Expected non-empty path")
		return
	}

	// First node should be monster's position
	if path[0].X != 1 || path[0].Y != 1 {
		t.Errorf("Expected first node to be (1,1), got (%d,%d)", path[0].X, path[0].Y)
	}

	// Last node should be player's position
	lastNode := path[len(path)-1]
	if lastNode.X != 9 || lastNode.Y != 9 {
		t.Errorf("Expected last node to be (9,9), got (%d,%d)", lastNode.X, lastNode.Y)
	}
}

func TestMonsterSpecialAbilities(t *testing.T) {
	player := NewPlayer(5, 5)

	// Test Eye perfect accuracy
	eye := NewMonster(5, 5, 'E')
	eyeHitChance := eye.calculateHitChance(player)
	if eyeHitChance < 0.95 { // Allow for small floating point differences
		t.Errorf("Expected Eye to have high accuracy (>0.95), got %f", eyeHitChance)
	}

	// Test Dragon fire damage
	dragon := NewMonster(5, 5, 'D')
	baseDamage := dragon.CalculateDamage(player.GetTotalDefense())
	dragonDamage := dragon.applyDamageModifiers(baseDamage, player)
	if dragonDamage <= baseDamage {
		t.Error("Expected Dragon to have fire damage bonus")
	}

	// Test Vampire life drain
	vampire := NewMonster(5, 5, 'V')
	vampire.HP = vampire.MaxHP / 2 // Damage the vampire first
	vampireDamage := vampire.applyDamageModifiers(baseDamage, player)
	if vampireDamage <= baseDamage {
		t.Error("Expected Vampire to have damage bonus")
	}

	// Test Leprechaun gold stealing
	leprechaun := NewMonster(5, 5, 'L')
	player.Gold = 100
	initialGold := player.Gold

	// Apply special effects multiple times to test probability
	for i := 0; i < 100; i++ {
		leprechaun.applySpecialEffects(player)
	}

	// Should have stolen some gold in 100 attempts
	if player.Gold == initialGold {
		t.Error("Expected Leprechaun to steal some gold in 100 attempts")
	}
}

func TestMonsterIntelligence(t *testing.T) {
	// Test intelligent monsters
	intelligentMonsters := []rune{'C', 'D', 'M', 'P', 'Q', 'V', 'X', 'Y'}

	for _, monsterType := range intelligentMonsters {
		monster := NewMonster(5, 5, monsterType)
		if !monster.isIntelligent() {
			t.Errorf("Expected monster %c to be intelligent", monsterType)
		}
	}

	// Test non-intelligent monsters
	simpleMonster := NewMonster(5, 5, 'B') // Bat
	if simpleMonster.isIntelligent() {
		t.Error("Expected Bat to not be intelligent")
	}
}

func TestMonsterViewRanges(t *testing.T) {
	// Test monsters with special view ranges
	eye := NewMonster(5, 5, 'E')
	if eye.ViewRange != 10 {
		t.Errorf("Expected Eye to have view range 10, got %d", eye.ViewRange)
	}

	dragon := NewMonster(5, 5, 'D')
	if dragon.ViewRange != 8 {
		t.Errorf("Expected Dragon to have view range 8, got %d", dragon.ViewRange)
	}

	bat := NewMonster(5, 5, 'B')
	if bat.ViewRange != 3 {
		t.Errorf("Expected Bat to have view range 3, got %d", bat.ViewRange)
	}

	fungus := NewMonster(5, 5, 'F')
	if fungus.ViewRange != 2 {
		t.Errorf("Expected Fungus to have view range 2, got %d", fungus.ViewRange)
	}
}

func TestMonsterDetectionRanges(t *testing.T) {
	// Test monsters with special detection ranges
	leprechaun := NewMonster(5, 5, 'L')
	if leprechaun.DetectionRange != 6 {
		t.Errorf("Expected Leprechaun to have detection range 6, got %d", leprechaun.DetectionRange)
	}

	vampire := NewMonster(5, 5, 'V')
	if vampire.DetectionRange != 7 {
		t.Errorf("Expected Vampire to have detection range 7, got %d", vampire.DetectionRange)
	}

	fungus := NewMonster(5, 5, 'F')
	if fungus.DetectionRange != 1 {
		t.Errorf("Expected Fungus to have detection range 1, got %d", fungus.DetectionRange)
	}

	zombie := NewMonster(5, 5, 'Z')
	if zombie.DetectionRange != 2 {
		t.Errorf("Expected Zombie to have detection range 2, got %d", zombie.DetectionRange)
	}
}

func TestMonsterAIStateTransitions(t *testing.T) {
	level := NewMockLevelCollisionChecker(20, 20)
	monster := NewMonster(10, 10, 'G') // Goblin
	player := NewPlayer(5, 5)

	// Test idle -> patrol
	monster.AIState = StateIdle
	monster.AlertLevel = 0
	monster.UpdateAIState(player, level, false, 10.0)
	if monster.AIState != StatePatrol && monster.AIState != StateIdle {
		t.Errorf("Expected state to be StatePatrol or StateIdle, got %v", monster.AIState)
	}

	// Test patrol -> chase (when player is visible)
	monster.AIState = StatePatrol
	monster.UpdateAIState(player, level, true, 5.0)
	if monster.AIState != StateChase {
		t.Errorf("Expected state to be StateChase when player is visible, got %v", monster.AIState)
	}

	// Test chase -> attack (when player is adjacent)
	monster.AIState = StateChase
	monster.UpdateAIState(player, level, true, 1.0)
	if monster.AIState != StateAttack {
		t.Errorf("Expected state to be StateAttack when player is adjacent, got %v", monster.AIState)
	}

	// Test chase -> search (when player disappears)
	monster.AIState = StateChase
	monster.LastPlayerPos = entity.Position{X: 5, Y: 5}
	monster.UpdateAIState(player, level, false, 10.0)
	if monster.AIState != StateSearch {
		t.Errorf("Expected state to be StateSearch when player disappears, got %v", monster.AIState)
	}
}

func TestMonsterFleeingBehavior(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	monster := NewMonster(5, 5, 'G') // Goblin
	player := NewPlayer(2, 2)        // Further away to avoid attack state

	// Damage the monster to low health
	monster.HP = monster.MaxHP / 4

	// Test fleeing behavior (distance > 1.5)
	monster.UpdateAIState(player, level, true, 3.0)
	if monster.AIState != StateFlee {
		t.Errorf("Expected state to be StateFlee when low on health, got %d", monster.AIState)
	}

	// Test that dragons and trolls don't flee
	dragon := NewMonster(5, 5, 'D')
	dragon.HP = dragon.MaxHP / 4
	dragon.UpdateAIState(player, level, true, 3.0)
	if dragon.AIState == StateFlee {
		t.Error("Expected Dragon to not flee even when low on health")
	}
}
