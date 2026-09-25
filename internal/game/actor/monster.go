package actor

import (
	"math"
	"math/rand"
	"slices"

	"github.com/yuru-sha/gorogue/internal/core/entity"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

// MonsterType contains the source-defined Rogue monster record.
type MonsterType struct {
	Code        rune
	Name        string
	Level       int
	HP          int
	Attack      int
	Defense     int // Rogue armor class; lower values are harder to hit.
	Experience  int
	Damage      string
	CarryChance int
	Speed       int
	MinFloor    int
	MaxFloor    int
}

const monsterOneDieEightSides = "1x8"

// MonsterTypes is Rogue 5.4.4's A-Z monster table. Spawn bounds describe the
// floors where the source's randmonster level window can select the symbol.
var MonsterTypes = map[rune]MonsterType{
	'A': {Code: 'A', Name: "aquator", Level: 5, HP: 1, Attack: 5, Defense: 2, Experience: 20, Damage: "0x0/0x0", Speed: 1, MinFloor: 9, MaxFloor: 18},
	'B': {Code: 'B', Name: "bat", Level: 1, HP: 1, Attack: 1, Defense: 3, Experience: 1, Damage: "1x2", Speed: 1, MinFloor: 1, MaxFloor: 8},
	'C': {Code: 'C', Name: "centaur", Level: 4, HP: 1, Attack: 4, Defense: 4, Experience: 17, Damage: "1x2/1x5/1x5", CarryChance: 15, Speed: 1, MinFloor: 7, MaxFloor: 16},
	'D': {Code: 'D', Name: "dragon", Level: 10, HP: 1, Attack: 10, Defense: -1, Experience: 5000, Damage: monsterOneDieEightSides + "/" + monsterOneDieEightSides + "/3x10", CarryChance: 100, Speed: 1, MinFloor: 22, MaxFloor: 26},
	'E': {Code: 'E', Name: "emu", Level: 1, HP: 1, Attack: 1, Defense: 7, Experience: 2, Damage: "1x2", Speed: 1, MinFloor: 1, MaxFloor: 7},
	'F': {Code: 'F', Name: "venus flytrap", Level: 8, HP: 1, Attack: 8, Defense: 3, Experience: 80, Damage: "%%%x0", Speed: 1, MinFloor: 12, MaxFloor: 21},
	'G': {Code: 'G', Name: "griffin", Level: 13, HP: 1, Attack: 13, Defense: 2, Experience: 2000, Damage: "4x3/3x5", CarryChance: 20, Speed: 1, MinFloor: 20, MaxFloor: 26},
	'H': {Code: 'H', Name: "hobgoblin", Level: 1, HP: 1, Attack: 1, Defense: 5, Experience: 3, Damage: monsterOneDieEightSides, Speed: 1, MinFloor: 1, MaxFloor: 10},
	'I': {Code: 'I', Name: "ice monster", Level: 1, HP: 1, Attack: 1, Defense: 9, Experience: 5, Damage: "0x0", Speed: 1, MinFloor: 2, MaxFloor: 11},
	'J': {Code: 'J', Name: "jabberwock", Level: 15, HP: 1, Attack: 15, Defense: 6, Experience: 3000, Damage: "2x12/2x4", CarryChance: 70, Speed: 1, MinFloor: 21, MaxFloor: 26},
	'K': {Code: 'K', Name: "kestrel", Level: 1, HP: 1, Attack: 1, Defense: 7, Experience: 1, Damage: "1x4", Speed: 1, MinFloor: 1, MaxFloor: 6},
	'L': {Code: 'L', Name: "leprechaun", Level: 3, HP: 1, Attack: 3, Defense: 8, Experience: 10, Damage: "1x1", Speed: 1, MinFloor: 6, MaxFloor: 15},
	'M': {Code: 'M', Name: "medusa", Level: 8, HP: 1, Attack: 8, Defense: 2, Experience: 200, Damage: "3x4/3x4/2x5", CarryChance: 40, Speed: 1, MinFloor: 18, MaxFloor: 26},
	'N': {Code: 'N', Name: "nymph", Level: 3, HP: 1, Attack: 3, Defense: 9, Experience: 37, Damage: "0x0", CarryChance: 100, Speed: 1, MinFloor: 10, MaxFloor: 19},
	'O': {Code: 'O', Name: "orc", Level: 1, HP: 1, Attack: 1, Defense: 6, Experience: 5, Damage: monsterOneDieEightSides, CarryChance: 15, Speed: 1, MinFloor: 4, MaxFloor: 13},
	'P': {Code: 'P', Name: "phantom", Level: 8, HP: 1, Attack: 8, Defense: 3, Experience: 120, Damage: "4x4", Speed: 1, MinFloor: 15, MaxFloor: 24},
	'Q': {Code: 'Q', Name: "quagga", Level: 3, HP: 1, Attack: 3, Defense: 3, Experience: 15, Damage: "1x5/1x5", Speed: 1, MinFloor: 8, MaxFloor: 17},
	'R': {Code: 'R', Name: "rattlesnake", Level: 2, HP: 1, Attack: 2, Defense: 3, Experience: 9, Damage: "1x6", Speed: 1, MinFloor: 3, MaxFloor: 12},
	'S': {Code: 'S', Name: "snake", Level: 1, HP: 1, Attack: 1, Defense: 5, Experience: 2, Damage: "1x3", Speed: 1, MinFloor: 1, MaxFloor: 9},
	'T': {Code: 'T', Name: "troll", Level: 6, HP: 1, Attack: 6, Defense: 4, Experience: 120, Damage: monsterOneDieEightSides + "/" + monsterOneDieEightSides + "/2x6", CarryChance: 50, Speed: 1, MinFloor: 13, MaxFloor: 22},
	'U': {Code: 'U', Name: "black unicorn", Level: 7, HP: 1, Attack: 7, Defense: -2, Experience: 190, Damage: "1x9/1x9/2x9", CarryChance: 30, Speed: 1, MinFloor: 17, MaxFloor: 26},
	'V': {Code: 'V', Name: "vampire", Level: 8, HP: 1, Attack: 8, Defense: 1, Experience: 350, Damage: "1x10", CarryChance: 20, Speed: 1, MinFloor: 19, MaxFloor: 26},
	'W': {Code: 'W', Name: "wraith", Level: 5, HP: 1, Attack: 5, Defense: 4, Experience: 55, Damage: "1x6", Speed: 1, MinFloor: 14, MaxFloor: 23},
	'X': {Code: 'X', Name: "xeroc", Level: 7, HP: 1, Attack: 7, Defense: 7, Experience: 100, Damage: "4x4", CarryChance: 30, Speed: 1, MinFloor: 16, MaxFloor: 25},
	'Y': {Code: 'Y', Name: "yeti", Level: 4, HP: 1, Attack: 4, Defense: 6, Experience: 50, Damage: "1x6/1x6", CarryChance: 30, Speed: 1, MinFloor: 11, MaxFloor: 20},
	'Z': {Code: 'Z', Name: "zombie", Level: 2, HP: 1, Attack: 2, Defense: 8, Experience: 6, Damage: monsterOneDieEightSides, Speed: 1, MinFloor: 5, MaxFloor: 14},
}

// Monster represents a monster in the game
type Monster struct {
	*Actor
	Type           MonsterType
	TurnCount      int
	IsActive       bool
	IsRunning      bool
	IsHeld         bool
	WasAdjacent    bool
	IsFound        bool
	IsConfused     bool
	IsCancelled    bool
	IsInvisible    bool
	IsHasted       bool
	IsSlowed       bool
	FlytrapHits    int
	GoldValue      int
	GreedTarget    entity.Position
	HasGreedTarget bool
	Carry          bool
	Mean           bool
	Flying         bool
	Greedy         bool
	Regenerates    bool
	Floor          int
}

// NewMonster creates a floor-one monster with an unshared convenience RNG.
// Game spawners should use NewMonsterWithRandAndFloor with the level's RNG.
func NewMonster(x, y int, monsterType rune) *Monster {
	return NewMonsterWithRandAndFloor(x, y, monsterType, 1, newRandomSource())
}

// NewMonsterWithRandAndFloor constructs a monster using Rogue's level-based
// hit-point and experience formulas and the game's random source.
func NewMonsterWithRandAndFloor(x, y int, monsterType rune, floor int, rng *rand.Rand) *Monster {
	mType, ok := MonsterTypes[monsterType]
	if !ok {
		return nil
	}
	if rng == nil {
		rng = newRandomSource()
	}
	levAdd := max(0, floor-26)
	level := mType.Level + levAdd
	mType.Level = level
	hp := 0
	for range level {
		hp += rng.Intn(8) + 1
	}
	armor := mType.Defense - levAdd
	mType.Experience += levAdd*10 + monsterExperienceBonus(level, hp)
	flags := monsterFlags(monsterType)
	monster := &Monster{
		Actor:       NewActor(x, y, hp, level, armor),
		Type:        mType,
		TurnCount:   1,
		IsActive:    true,
		Mean:        flags.mean,
		Flying:      flags.flying,
		Greedy:      flags.greedy,
		Regenerates: flags.regenerates,
		Floor:       floor,
		IsInvisible: flags.invisible,
	}
	return monster
}

func monsterExperienceBonus(level, hp int) int {
	divisor := 6
	if level == 1 {
		divisor = 8
	}
	bonus := hp / divisor
	if level > 9 {
		bonus *= 20
	} else if level > 6 {
		bonus *= 4
	}
	return bonus
}

type monsterFlagsValue struct {
	mean, flying, greedy, regenerates, invisible bool
}

func monsterFlags(symbol rune) monsterFlagsValue {
	switch symbol {
	case 'A', 'D', 'E', 'F', 'H', 'M', 'Q', 'R', 'S', 'U', 'Z':
		return monsterFlagsValue{mean: true}
	case 'G':
		return monsterFlagsValue{mean: true, flying: true, regenerates: true}
	case 'T', 'V':
		return monsterFlagsValue{mean: true, regenerates: true}
	case 'B', 'K':
		return monsterFlagsValue{flying: true}
	case 'O':
		return monsterFlagsValue{greedy: true}
	case 'P':
		return monsterFlagsValue{invisible: true}
	default:
		return monsterFlagsValue{}
	}
}

// GetValidMonsterTypesForFloor returns symbols that Rogue's level window can
// select on a floor, in stable symbol order.
func GetValidMonsterTypesForFloor(floor int) []rune {
	if floor < 1 || floor > 26 {
		return nil
	}
	var validTypes []rune
	for symbol, mType := range MonsterTypes {
		if floor >= mType.MinFloor && floor <= mType.MaxFloor {
			validTypes = append(validTypes, symbol)
		}
	}
	slices.Sort(validTypes)
	return validTypes
}

// GetRandomMonsterTypeForFloor returns a source-distributed monster symbol.
func GetRandomMonsterTypeForFloor(floor int) rune {
	return GetRandomMonsterTypeForFloorWithRand(floor, nil)
}

// GetRandomMonsterTypeForFloorWithRand uses Rogue's randmonster level window:
// level + rnd(10)-6, with the source's shallow/deep boundary wraps.
func GetRandomMonsterTypeForFloorWithRand(floor int, rng *rand.Rand) rune {
	if rng == nil {
		rng = newRandomSource()
	}
	d := floor + rng.Intn(10) - 6
	if d < 0 {
		d = rng.Intn(5)
	} else if d > 25 {
		d = rng.Intn(5) + 21
	}
	return monsterLevelOrder[d]
}

var monsterLevelOrder = [...]rune{
	'K', 'E', 'B', 'S', 'H', 'I', 'R', 'O', 'Z', 'L', 'C', 'Q', 'A',
	'N', 'Y', 'F', 'T', 'W', 'P', 'X', 'U', 'M', 'V', 'G', 'J', 'D',
}

// LevelCollisionChecker is an interface for collision detection
type LevelCollisionChecker interface {
	IsInBounds(x, y int) bool
	IsWalkable(x, y int) bool
	GetMonsterAt(x, y int) *Monster
}

func (m *Monster) DistanceToPlayer(player *Player) float64 {
	dx := float64(m.Position.X - player.Position.X)
	dy := float64(m.Position.Y - player.Position.Y)
	return math.Hypot(dx, dy)
}

func (m *Monster) Update(player *Player, level LevelCollisionChecker) {
	if !m.IsActive || !m.IsAlive() || !player.IsAlive() || m.IsHeld {
		return
	}
	rng := player.RandomSource()
	adjacent := m.DistanceToPlayer(player) <= math.Sqrt2
	m.updateRunningState(player, adjacent, rng)
	m.WasAdjacent = adjacent
	if !m.IsRunning {
		return
	}
	m.checkFound(player, level)
	m.moveMonsterTurn(player, level, rng)
	if m.Flying && m.DistanceToPlayer(player) >= 3 {
		m.moveMonsterTurn(player, level, rng)
	}
}

func (m *Monster) updateRunningState(player *Player, adjacent bool, rng *rand.Rand) {
	if m.IsRunning || !m.Mean || !adjacent || m.WasAdjacent ||
		player.LevitationTurns != 0 || wearingStealth(player) || rng.Intn(3) == 0 {
		return
	}
	m.IsRunning = true
}

func (m *Monster) checkFound(player *Player, level LevelCollisionChecker) {
	if m.Type.Code != 'M' || m.IsCancelled || m.IsFound ||
		player.BlindTurns != 0 || m.DistanceToPlayer(player) >= 3 ||
		!m.hasLineOfSight(player.Position.X, player.Position.Y, level) {
		return
	}
	m.IsFound = true
	if !player.SaveAgainst(3) {
		player.ConfusedTurns += 20
	}
}

func (m *Monster) moveMonsterTurn(player *Player, level LevelCollisionChecker, rng *rand.Rand) {
	if !m.IsSlowed || m.TurnCount != 0 {
		if !m.tryDragonBreath(player, level, rng) {
			m.chaseAction(player, level, rng)
		}
	}
	if m.IsHasted {
		if !m.tryDragonBreath(player, level, rng) {
			m.chaseAction(player, level, rng)
		}
	}
	m.TurnCount ^= 1
}

func (m *Monster) chaseAction(player *Player, level LevelCollisionChecker, rng *rand.Rand) {
	confused := m.IsConfused && rng.Intn(5) != 0
	if m.Type.Code == 'P' && rng.Intn(5) == 0 {
		confused = true
	}
	if m.Type.Code == 'B' && rng.Intn(2) == 0 {
		confused = true
	}
	if confused {
		m.moveRandomlyWithRand(level, rng)
		if m.IsConfused && rng.Intn(20) == 0 {
			m.IsConfused = false
		}
		return
	}
	if m.DistanceToPlayer(player) <= math.Sqrt2 {
		m.AttackPlayer(player)
	} else {
		m.MoveTowardsPlayer(player, level)
	}
}

func wearingStealth(player *Player) bool {
	return (player.Equipment.RingLeft != nil && player.Equipment.RingLeft.RealName == ringStealthName) ||
		(player.Equipment.RingRight != nil && player.Equipment.RingRight.RealName == ringStealthName)
}

func (m *Monster) tryDragonBreath(player *Player, level LevelCollisionChecker, rng *rand.Rand) bool {
	dx := abs(m.Position.X - player.Position.X)
	dy := abs(m.Position.Y - player.Position.Y)
	if m.Type.Code != 'D' || m.IsCancelled || dx*dx+dy*dy > 36 ||
		(dx != 0 && dy != 0 && dx != dy) || rng.Intn(5) != 0 {
		return false
	}
	if player.SaveAgainst(3) {
		return true
	}
	damage := 0
	for range 6 {
		damage += rng.Intn(6) + 1
	}
	player.TakeDamage(damage)
	return true
}

// AttackPlayer uses Rogue's level/armor swing and the monster's source dice.
func (m *Monster) AttackPlayer(player *Player) {
	rng := player.RandomSource()
	hit := false
	damage := 0
	damageSpec := m.Type.Damage
	for start := 0; start < len(damageSpec); {
		end := start
		for end < len(damageSpec) && damageSpec[end] != '/' {
			end++
		}
		dice, sides := parseMonsterDice(damageSpec[start:end])
		if m.Type.Code == 'F' && m.FlytrapHits > 0 {
			dice, sides = m.FlytrapHits, 1
		}
		if player.RollHit(m.Type.Level, player.GetTotalDefense(), 0, 10, player.Running) {
			hit = true
			for range dice {
				damage += rng.Intn(sides) + 1
			}
		}
		start = end + 1
	}
	if !hit {
		return
	}
	player.TakeDamage(damage)
	if !player.IsAlive() {
		return
	}
	m.applySourceSpecial(player, rng)
}

func parseMonsterDice(spec string) (dice, sides int) {
	value := 0
	readingSides := false
	for i := 0; i < len(spec); i++ {
		if spec[i] == 'x' {
			dice, value, readingSides = value, 0, true
			continue
		}
		if spec[i] >= '0' && spec[i] <= '9' {
			value = value*10 + int(spec[i]-'0')
		} else {
			value = 0
		}
		if readingSides {
			sides = value
		}
	}
	return dice, sides
}

func (m *Monster) applySourceSpecial(player *Player, rng *rand.Rand) {
	if m.IsCancelled {
		return
	}
	switch m.Type.Code {
	case 'A':
		player.RustArmor()
	case 'I':
		player.Freeze(rng.Intn(2) + 2)
	case 'R':
		if !player.SaveAgainst(0) {
			player.ReduceStrength(1)
		}
	case 'W':
		if rng.Intn(100) < 15 {
			if player.Exp == 0 {
				player.TakeDamage(player.HP)
				return
			}
			player.DrainLevel()
			player.DrainMaxHP(rng.Intn(10) + 1)
		}
	case 'V':
		if rng.Intn(100) < 30 {
			player.DrainMaxHP(rng.Intn(3) + 1)
		}
	case 'F':
		m.FlytrapHits++
		player.Hold()
		player.TakeDamage(m.FlytrapHits)
	case 'L':
		m.stealGold(player, rng)
		m.IsActive = false
	case 'N':
		m.stealMagicItem(player, rng)
	}
}

func (m *Monster) stealGold(player *Player, rng *rand.Rand) {
	if player.Gold <= 0 {
		return
	}
	amount := rng.Intn(50+10*m.Floor) + 2
	if player.SaveAgainst(3) {
		return
	}
	amount *= 5
	if amount > player.Gold {
		amount = player.Gold
	}
	player.Gold -= amount
}

func (m *Monster) stealMagicItem(player *Player, rng *rand.Rand) {
	selected, count := -1, 0
	for index, carried := range player.Inventory.Items {
		if carried == player.Equipment.Armor || carried == player.Equipment.Weapon ||
			carried == player.Equipment.RingLeft || carried == player.Equipment.RingRight {
			continue
		}
		magical := false
		switch carried.Type {
		case item.ItemPotion, item.ItemScroll, item.ItemWand, item.ItemRing, item.ItemAmulet:
			magical = true
		case item.ItemArmor:
			magical = carried.IsProtected || carried.Enchantment != 0
		case item.ItemWeapon:
			magical = carried.Enchantment != 0
		}
		if magical {
			count++
			if rng.Intn(count) == 0 {
				selected = index
			}
		}
	}
	if selected >= 0 {
		player.Inventory.RemoveItem(selected)
		m.IsActive = false
	}
}

// MoveTowardsPlayer advances one step along Rogue's direct chase route.
func (m *Monster) MoveTowardsPlayer(player *Player, level LevelCollisionChecker) {
	m.moveTowards(player.Position.X, player.Position.Y, level)
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
		if x != x0 || y != y0 {
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

func (m *Monster) moveTowards(targetX, targetY int, level LevelCollisionChecker) {
	dx := sign(targetX - m.Position.X)
	dy := sign(targetY - m.Position.Y)
	if dx != 0 && dy != 0 && m.canMoveStep(m.Position.X+dx, m.Position.Y+dy, level) {
		m.Position.Move(dx, dy)
		return
	}
	if dx != 0 && m.canMoveStep(m.Position.X+dx, m.Position.Y, level) {
		m.Position.Move(dx, 0)
		return
	}
	if dy != 0 && m.canMoveStep(m.Position.X, m.Position.Y+dy, level) {
		m.Position.Move(0, dy)
		return
	}
	bestDistance := int(^uint(0) >> 1)
	bestDX, bestDY := 0, 0
	for _, direction := range monsterSteps {
		nx, ny := m.Position.X+direction.X, m.Position.Y+direction.Y
		if !m.canMoveStep(nx, ny, level) {
			continue
		}
		dx, dy := nx-targetX, ny-targetY
		distance := dx*dx + dy*dy
		if distance < bestDistance {
			bestDistance, bestDX, bestDY = distance, direction.X, direction.Y
		}
	}
	m.Position.Move(bestDX, bestDY)
}

var monsterSteps = [...]entity.Position{
	{X: -1, Y: -1}, {X: -1}, {X: -1, Y: 1}, {Y: -1},
	{Y: 1}, {X: 1, Y: -1}, {X: 1}, {X: 1, Y: 1},
}

func (m *Monster) moveRandomlyWithRand(level LevelCollisionChecker, rng *rand.Rand) {
	selected := entity.Position{}
	count := 0
	for _, direction := range monsterSteps {
		if !m.canMoveStep(m.Position.X+direction.X, m.Position.Y+direction.Y, level) {
			continue
		}
		count++
		if rng.Intn(count) == 0 {
			selected = direction
		}
	}
	m.Position.Move(selected.X, selected.Y)
}

func (m *Monster) canMoveStep(x, y int, level LevelCollisionChecker) bool {
	dx, dy := x-m.Position.X, y-m.Position.Y
	if dx != 0 && dy != 0 &&
		(!level.IsWalkable(m.Position.X+dx, m.Position.Y) ||
			!level.IsWalkable(m.Position.X, m.Position.Y+dy)) {
		return false
	}
	return m.CanMoveTo(x, y, level)
}

// CanMoveTo checks that a source monster can enter an open, unoccupied tile.
func (m *Monster) CanMoveTo(x, y int, level LevelCollisionChecker) bool {
	return level.IsInBounds(x, y) && level.IsWalkable(x, y) && level.GetMonsterAt(x, y) == nil
}

func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
