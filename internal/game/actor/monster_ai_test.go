package actor

import (
	"fmt"
	"math/rand"
	"slices"
	"sort"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestValidMonsterTypesAreStableAndFloorBound(t *testing.T) {
	for floor := 1; floor <= 26; floor++ {
		types := GetValidMonsterTypesForFloor(floor)
		if !sort.SliceIsSorted(types, func(i, j int) bool { return types[i] < types[j] }) {
			t.Fatalf("floor %d monster types are not sorted: %q", floor, string(types))
		}
		for _, symbol := range types {
			monsterType := MonsterTypes[symbol]
			if floor < monsterType.MinFloor || floor > monsterType.MaxFloor {
				t.Errorf("floor %d contains invalid monster %c", floor, symbol)
			}
		}
	}
}
func TestMonsterDefinitionsMatchRogue544(t *testing.T) {
	want := map[rune]struct {
		name                     string
		level, experience, armor int
		damage                   string
		firstFloor, lastFloor    int
	}{
		'A': {"aquator", 5, 20, 2, "0x0/0x0", 9, 18},
		'B': {"bat", 1, 1, 3, "1x2", 1, 8},
		'C': {"centaur", 4, 17, 4, "1x2/1x5/1x5", 7, 16},
		'D': {"dragon", 10, 5000, -1, "1x8/1x8/3x10", 22, 26},
		'E': {"emu", 1, 2, 7, "1x2", 1, 7},
		'F': {"venus flytrap", 8, 80, 3, "%%%x0", 12, 21},
		'G': {"griffin", 13, 2000, 2, "4x3/3x5", 20, 26},
		'H': {"hobgoblin", 1, 3, 5, "1x8", 1, 10},
		'I': {"ice monster", 1, 5, 9, "0x0", 2, 11},
		'J': {"jabberwock", 15, 3000, 6, "2x12/2x4", 21, 26},
		'K': {"kestrel", 1, 1, 7, "1x4", 1, 6},
		'L': {"leprechaun", 3, 10, 8, "1x1", 6, 15},
		'M': {"medusa", 8, 200, 2, "3x4/3x4/2x5", 18, 26},
		'N': {"nymph", 3, 37, 9, "0x0", 10, 19},
		'O': {"orc", 1, 5, 6, "1x8", 4, 13},
		'P': {"phantom", 8, 120, 3, "4x4", 15, 24},
		'Q': {"quagga", 3, 15, 3, "1x5/1x5", 8, 17},
		'R': {"rattlesnake", 2, 9, 3, "1x6", 3, 12},
		'S': {"snake", 1, 2, 5, "1x3", 1, 9},
		'T': {"troll", 6, 120, 4, "1x8/1x8/2x6", 13, 22},
		'U': {"black unicorn", 7, 190, -2, "1x9/1x9/2x9", 17, 26},
		'V': {"vampire", 8, 350, 1, "1x10", 19, 26},
		'W': {"wraith", 5, 55, 4, "1x6", 14, 23},
		'X': {"xeroc", 7, 100, 7, "4x4", 16, 25},
		'Y': {"yeti", 4, 50, 6, "1x6/1x6", 11, 20},
		'Z': {"zombie", 2, 6, 8, "1x8", 5, 14},
	}
	if len(MonsterTypes) != len(want) {
		t.Fatalf("got %d monster definitions, want all %d Rogue symbols", len(MonsterTypes), len(want))
	}
	for symbol, expected := range want {
		got, ok := MonsterTypes[symbol]
		if !ok {
			t.Errorf("missing Rogue monster %c", symbol)
			continue
		}
		if got.Name != expected.name || got.Level != expected.level || got.Experience != expected.experience ||
			got.Defense != expected.armor || got.Damage != expected.damage ||
			got.MinFloor != expected.firstFloor || got.MaxFloor != expected.lastFloor {
			t.Errorf("%c definition = %+v, want name=%q level=%d xp=%d armor=%d damage=%q spawn=%d..%d",
				symbol, got, expected.name, expected.level, expected.experience, expected.armor,
				expected.damage, expected.firstFloor, expected.lastFloor)
		}
	}
	carryChance := map[rune]int{
		'C': 15, 'D': 100, 'G': 20, 'J': 70, 'M': 40, 'N': 100,
		'O': 15, 'T': 50, 'U': 30, 'V': 20, 'X': 30, 'Y': 30,
	}
	for symbol, got := range MonsterTypes {
		if got.HP != 1 || got.Attack != got.Level || got.CarryChance != carryChance[symbol] {
			t.Errorf("%c source template = HP %d attack-level %d carry %d, want HP die marker 1, attack-level %d, carry %d",
				symbol, got.HP, got.Attack, got.CarryChance, got.Level, carryChance[symbol])
		}
	}
}

func TestMonsterSpawnUsesRogueLevelHPArmorAndExperience(t *testing.T) {
	newSpawn := func() *Monster {
		return NewMonsterWithRandAndFloor(2, 3, 'D', 27, rand.New(rand.NewSource(27)))
	}
	first, second := newSpawn(), newSpawn()
	if first.HP != second.HP || first.Type.Experience != second.Type.Experience ||
		first.Type.Level != second.Type.Level || first.Defense != second.Defense {
		t.Fatalf("same seed produced different monster stats: first=%+v second=%+v", first, second)
	}
	if first.Type.Level != 11 {
		t.Errorf("dragon level = %d, want 11 on floor 27", first.Type.Level)
	}
	if first.HP < 11 || first.HP > 88 {
		t.Errorf("dragon HP = %d, want 11d8 roll (11..88)", first.HP)
	}
	if first.Defense != -2 {
		t.Errorf("dragon armor class = %d, want source armor -1 minus floor bonus 1", first.Defense)
	}
	wantExp := 5000 + 10 + first.HP/6*20
	if first.Type.Experience != wantExp {
		t.Errorf("dragon XP = %d, want source XP formula result %d", first.Type.Experience, wantExp)
	}
}

func TestMonsterSpawnDistributionMatchesRogueLevelWindow(t *testing.T) {
	for _, floor := range []int{1, 2, 5, 14, 22, 26} {
		rng := rand.New(rand.NewSource(int64(floor)))
		first := GetRandomMonsterTypeForFloorWithRand(floor, rng)
		second := GetRandomMonsterTypeForFloorWithRand(floor, rand.New(rand.NewSource(int64(floor))))
		if first != second {
			t.Errorf("floor %d selection is not seed deterministic: %c != %c", floor, first, second)
		}
		if got := GetValidMonsterTypesForFloor(floor); !slices.Contains(got, first) {
			t.Errorf("floor %d selected %c outside source-eligible monsters %q", floor, first, string(got))
		}
	}
}

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

func TestMonsterCreationUsesRogueCombatStats(t *testing.T) {
	for symbol, source := range MonsterTypes {
		t.Run(string(symbol), func(t *testing.T) {
			monster := NewMonsterWithRandAndFloor(5, 5, symbol, 1, rand.New(rand.NewSource(int64(symbol))))
			if monster == nil || monster.Type.Code != symbol {
				t.Fatalf("monster for %c was not created", symbol)
			}
			if monster.Type.Level != source.Level || monster.Attack != source.Level ||
				monster.Defense != source.Defense {
				t.Errorf("%c combat stats = level %d armor %d, want level %d armor %d",
					symbol, monster.Attack, monster.Defense, source.Level, source.Defense)
			}
			if monster.HP < source.Level || monster.HP > source.Level*8 {
				t.Errorf("%c HP = %d, want source %dd8 bounds [%d,%d]",
					symbol, monster.HP, source.Level, source.Level, source.Level*8)
			}
		})
	}
}

func TestMonsterMovementFollowsSourceRunState(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	player := NewPlayerWithSeed(4, 4, 1)
	monster := NewMonsterWithRandAndFloor(1, 1, 'H', 1, rand.New(rand.NewSource(1)))
	player.SetRandomSource(rand.New(rand.NewSource(1)))

	monster.Update(player, level)
	if monster.Position.X != 1 || monster.Position.Y != 1 {
		t.Fatalf("unawakened hobgoblin moved to (%d,%d)", monster.Position.X, monster.Position.Y)
	}
	monster.IsRunning = true
	monster.Update(player, level)
	if monster.Position.X != 2 || monster.Position.Y != 2 {
		t.Errorf("running hobgoblin moved to (%d,%d), want direct chase step (2,2)",
			monster.Position.X, monster.Position.Y)
	}
}

func TestMeanMonsterWakesAndAttacksWhenAdjacent(t *testing.T) {
	level := NewMockLevelCollisionChecker(10, 10)
	player := NewPlayerWithSeed(2, 1, 1)
	player.SetRandomSource(rand.New(rand.NewSource(1)))
	player.Defense = 1000
	monster := NewMonsterWithRandAndFloor(1, 1, 'H', 1, rand.New(rand.NewSource(1)))
	before := player.HP

	monster.Update(player, level)
	damage := before - player.HP
	if !monster.IsRunning {
		t.Fatal("adjacent mean monster did not wake")
	}
	if damage < 1 || damage > 8 {
		t.Errorf("hobgoblin hit damage = %d, want its source 1d8", damage)
	}
}

func TestMonsterAttackUsesSourceDice(t *testing.T) {
	newAttack := func() (int, int) {
		player := NewPlayerWithSeed(0, 0, 42)
		player.Defense = 1000
		monster := NewMonsterWithRandAndFloor(1, 1, 'Q', 1, rand.New(rand.NewSource(1)))
		before := player.HP
		monster.AttackPlayer(player)
		return before - player.HP, player.HP
	}
	damage1, hp1 := newAttack()
	damage2, hp2 := newAttack()
	if damage1 != damage2 || hp1 != hp2 {
		t.Fatalf("same seed produced attacks (%d,%d) and (%d,%d)", damage1, hp1, damage2, hp2)
	}
	if damage1 < 2 || damage1 > 10 {
		t.Errorf("quagga attack damage = %d, want two source 1d5 attacks", damage1)
	}
}

func TestFlytrapAttackHoldsAndEscalatesDamage(t *testing.T) {
	player := NewPlayerWithSeed(0, 0, 12)
	player.Defense = 1000
	flytrap := NewMonsterWithRandAndFloor(1, 0, 'F', 10, rand.New(rand.NewSource(12)))
	initialHP := player.HP

	flytrap.AttackPlayer(player)
	firstDamage := initialHP - player.HP
	if firstDamage != 1 {
		t.Fatalf("first flytrap hit dealt %d damage, want 1", firstDamage)
	}
	initialHP = player.HP
	flytrap.AttackPlayer(player)
	if got := initialHP - player.HP; got != 3 {
		t.Errorf("second flytrap hit dealt %d damage, want 1d1 plus escalating 2", got)
	}
	if !player.Held {
		t.Error("flytrap hit did not hold the player")
	}
}
func TestRogueMonsterSpecialAttacks(t *testing.T) {
	t.Run("ice freezes", func(t *testing.T) {
		player := NewPlayerWithSeed(0, 0, 9)
		player.SetRandomSource(rand.New(rand.NewSource(9)))
		player.Defense = 1000
		ice := NewMonsterWithRandAndFloor(1, 0, 'I', 1, rand.New(rand.NewSource(9)))

		ice.AttackPlayer(player)
		if player.NoCommandTurns < 2 || player.NoCommandTurns > 3 {
			t.Errorf("ice hit freeze duration = %d, want source 2..3 turns", player.NoCommandTurns)
		}
	})
	t.Run("aquator rusts armor", func(t *testing.T) {
		player := NewPlayerWithSeed(0, 0, 4)
		player.SetRandomSource(rand.New(rand.NewSource(4)))
		player.Defense = 1000
		player.Equipment.Armor = item.NewItem(0, 0, item.ItemArmor, "leather armor", 20)
		player.Equipment.Armor.Defense = 8
		before := player.GetTotalDefense()
		aquator := NewMonsterWithRandAndFloor(1, 0, 'A', 1, rand.New(rand.NewSource(4)))

		aquator.AttackPlayer(player)
		if got := player.GetTotalDefense(); got <= before {
			t.Errorf("aquator hit left armor class %d unchanged or improved from %d", got, before)
		}
	})
	t.Run("nymph steals a carried magic item and leaves", func(t *testing.T) {
		player := NewPlayerWithSeed(0, 0, 11)
		player.SetRandomSource(rand.New(rand.NewSource(11)))
		player.Defense = 1000
		scroll := item.NewItem(0, 0, item.ItemScroll, "magic mapping", 150)
		player.Inventory.AddItem(scroll)
		nymph := NewMonsterWithRandAndFloor(1, 0, 'N', 10, rand.New(rand.NewSource(11)))

		nymph.AttackPlayer(player)
		if len(player.Inventory.Items) != 0 || nymph.IsActive {
			t.Errorf("nymph attack left inventory size %d and active=%t; want stolen scroll and fleeing nymph",
				len(player.Inventory.Items), nymph.IsActive)
		}
	})
	t.Run("rattlesnake drains strength when save fails", func(t *testing.T) {
		drained := false
		for seed := int64(1); seed <= 64 && !drained; seed++ {
			player := NewPlayerWithSeed(0, 0, seed)
			player.SetRandomSource(rand.New(rand.NewSource(seed)))
			player.Defense = 1000
			player.Strength = 10
			before := player.Strength
			snake := NewMonsterWithRandAndFloor(1, 0, 'R', 1, rand.New(rand.NewSource(seed)))
			snake.AttackPlayer(player)
			drained = player.Strength < before
		}
		if !drained {
			t.Fatal("rattlesnake never reduced strength across deterministic failed-save seeds")
		}
	})
	t.Run("leprechaun steals gold and flees", func(t *testing.T) {
		stolen := false
		for seed := int64(1); seed <= 64 && !stolen; seed++ {
			player := NewPlayerWithSeed(0, 0, seed)
			player.SetRandomSource(rand.New(rand.NewSource(seed)))
			player.Defense = 1000
			player.Gold = 1000
			before := player.Gold
			leprechaun := NewMonsterWithRandAndFloor(1, 0, 'L', 1, rand.New(rand.NewSource(seed)))
			leprechaun.AttackPlayer(player)
			stolen = player.Gold < before && !leprechaun.IsActive
		}
		if !stolen {
			t.Fatal("leprechaun did not steal gold and disappear across deterministic save rolls")
		}
	})
}

func TestMonsterDeathLeavesSourceExperienceValue(t *testing.T) {
	monster := NewMonsterWithRandAndFloor(0, 0, 'B', 1, rand.New(rand.NewSource(4)))
	want := 1 + monster.HP/8
	if monster.Type.Experience != want {
		t.Errorf("bat kill XP = %d, want source base XP plus hp/8 (%d)", monster.Type.Experience, want)
	}
	monster.TakeDamage(monster.HP)
	if monster.IsAlive() {
		t.Fatal("monster remained alive after lethal damage")
	}
	if monster.Type.Experience != want {
		t.Errorf("monster XP changed on death: got %d, want %d", monster.Type.Experience, want)
	}
}
