package actor

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestPlayerEatFoodCapsHunger(t *testing.T) {
	player := NewPlayer(0, 0)
	player.Hunger = 90

	player.EatFood(20)

	if player.Hunger != 100 {
		t.Fatalf("Hunger = %d, want 100", player.Hunger)
	}
}

func TestPlayerAddExpFollowsRogueLevelTableAndHPGain(t *testing.T) {
	player := playerWithRolls(6)
	player.AddExp(9)
	if player.Level != 1 {
		t.Fatalf("Level at 9 XP = %d, want 1", player.Level)
	}

	player.TakeDamage(4)
	player.AddExp(1)

	if player.Level != 2 {
		t.Fatalf("Level at 10 XP = %d, want 2", player.Level)
	}
	if player.MaxHP != 19 {
		t.Fatalf("MaxHP after rolling 1d10 = %d, want 19", player.MaxHP)
	}
	if player.HP != 15 {
		t.Fatalf("HP after gaining 7 max HP = %d, want 15", player.HP)
	}
}

func TestPlayerAddExpUsesFullRogueThresholdTable(t *testing.T) {
	thresholds := []int{
		10, 20, 40, 80, 160, 320, 640, 1300, 2600, 5200,
		13000, 26000, 50000, 100000, 200000, 400000, 800000,
		2000000, 4000000, 8000000,
	}
	for index, threshold := range thresholds {
		t.Run(fmt.Sprintf("threshold_%d", threshold), func(t *testing.T) {
			player := playerWithRolls(0)
			player.AddExp(threshold - 1)
			if want := index + 1; player.Level != want {
				t.Fatalf("Level at %d XP = %d, want %d", threshold-1, player.Level, want)
			}

			player.AddExp(1)
			if want := index + 2; player.Level != want {
				t.Fatalf("Level at %d XP = %d, want %d", threshold, player.Level, want)
			}
		})
	}
}

func TestPlayerHungerChangesAtRogueFoodThreshold(t *testing.T) {
	player := playerWithRolls(1)
	player.FoodLeft = 300

	player.UpdateHunger()

	if player.FoodLeft != 299 {
		t.Fatalf("FoodLeft = %d, want 299", player.FoodLeft)
	}
	if player.HungerState != HungerHungry {
		t.Fatalf("HungerState = %d, want hungry (%d)", player.HungerState, HungerHungry)
	}
}

func TestPlayerHungerStageBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		foodLeft  int
		wantFood  int
		wantStage int
	}{
		{name: "exactly hungry threshold is unchanged", foodLeft: 301, wantFood: 300, wantStage: HungerSatisfied},
		{name: "below hungry threshold", foodLeft: 300, wantFood: 299, wantStage: HungerHungry},
		{name: "exactly weak threshold is unchanged", foodLeft: 151, wantFood: 150, wantStage: HungerSatisfied},
		{name: "below weak threshold", foodLeft: 150, wantFood: 149, wantStage: HungerWeak},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := playerWithRolls(1)
			player.FoodLeft = tt.foodLeft

			player.UpdateHunger()

			if player.FoodLeft != tt.wantFood || player.HungerState != tt.wantStage {
				t.Fatalf("food/stage = %d/%d, want %d/%d", player.FoodLeft, player.HungerState, tt.wantFood, tt.wantStage)
			}
		})
	}
}

func TestPlayerHungerFaintsAndStarvesAtRogueBoundaries(t *testing.T) {
	t.Run("faint uses source duration", func(t *testing.T) {
		player := playerWithRolls(0, 3)
		player.FoodLeft = 0

		player.UpdateHunger()

		if player.HungerState != HungerFainting || player.NoCommandTurns != 7 {
			t.Fatalf("state/forced turns = %d/%d, want fainting/7", player.HungerState, player.NoCommandTurns)
		}
		if player.HP != 12 {
			t.Fatalf("HP while fainting = %d, want unchanged 12", player.HP)
		}
	})

	t.Run("death follows 850 starvation ticks", func(t *testing.T) {
		player := playerWithRolls(1)
		player.FoodLeft = 0

		for range 851 {
			player.UpdateHunger()
		}
		if !player.IsAlive() || player.FoodLeft != -851 {
			t.Fatalf("after 851 ticks: alive=%v food=%d, want alive at -851", player.IsAlive(), player.FoodLeft)
		}

		player.UpdateHunger()
		if player.IsAlive() {
			t.Fatal("player survived the source starvation boundary at -851")
		}
	})
}

func TestPlayerEatFoodUsesRogueRefillRangeAndCap(t *testing.T) {
	player := playerWithRolls(0)
	player.FoodLeft = -200
	player.HungerState = HungerFainting

	player.EatFood(1)

	if player.FoodLeft != 1100 || player.HungerState != HungerSatisfied {
		t.Fatalf("food/stage after source minimum refill = %d/%d, want 1100/satisfied", player.FoodLeft, player.HungerState)
	}

	player = playerWithRolls(399)
	player.FoodLeft = 1500
	player.EatFood(0)
	if player.FoodLeft != 2000 {
		t.Fatalf("FoodLeft after source maximum refill = %d, want 2000", player.FoodLeft)
	}
}

func TestPlayerFoodDrainUsesEquippedSourceRingsAndAmulet(t *testing.T) {
	player := playerWithRolls(1)
	player.FoodLeft = 1000
	player.Equipment.RingLeft = item.NewItem(0, 0, item.ItemRing, "regeneration", 0)
	player.UpdateHunger()
	if player.FoodLeft != 997 {
		t.Fatalf("FoodLeft with regeneration ring = %d, want 997", player.FoodLeft)
	}

	player = playerWithRolls(1)
	player.FoodLeft = 1000
	if !player.Inventory.AddItem(item.NewAmulet(0, 0)) {
		t.Fatal("failed to add source amulet")
	}
	player.UpdateHunger()
	if player.FoodLeft != 1000 {
		t.Fatalf("FoodLeft with amulet = %d, want 1000", player.FoodLeft)
	}

	player = playerWithRolls(1)
	player.FoodLeft = 1000
	player.Equipment.RingLeft = item.NewItem(0, 0, item.ItemRing, "slow digestion", 0)
	player.UpdateHunger()
	if player.FoodLeft != 1000 {
		t.Fatalf("FoodLeft with successful slow digestion = %d, want 1000", player.FoodLeft)
	}
}

func TestPlayerAttackRollUsesRogueHitAndWeaponDamage(t *testing.T) {
	tests := []struct {
		name          string
		rolls         []int64
		targetRunning bool
		weapon        bool
		wantHit       bool
		wantDamage    int
	}{
		{name: "hit at inclusive threshold", rolls: []int64{9, 0}, targetRunning: true, wantHit: true, wantDamage: 2},
		{name: "miss below threshold", rolls: []int64{8}, targetRunning: true, wantHit: false},
		{name: "source bonus against non-running target", rolls: []int64{5, 0}, targetRunning: false, wantHit: true, wantDamage: 2},
		{name: "mace rolls two four-sided dice", rolls: []int64{9, 0, 3}, targetRunning: true, weapon: true, wantHit: true, wantDamage: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := playerWithRolls(tt.rolls...)
			if tt.weapon {
				player.Equipment.Weapon = item.NewItem(0, 0, item.ItemWeapon, "mace", 8)
			}

			got := player.AttackRoll(10, tt.targetRunning)

			if got.Hit != tt.wantHit {
				t.Fatalf("Hit = %v, want %v", got.Hit, tt.wantHit)
			}
			if got.Damage != tt.wantDamage {
				t.Fatalf("Damage = %d, want %d", got.Damage, tt.wantDamage)
			}
		})
	}
}

func TestPlayerStrengthAndEquipmentModifyRogueCombat(t *testing.T) {
	tests := []struct {
		name       string
		rolls      []int64
		strength   int
		weapon     bool
		weaponPlus int
		ring       string
		ringPlus   int
		wantDamage int
	}{
		{name: "minimum strength hit and damage penalties", rolls: []int64{13, 0}, strength: 3, wantDamage: 0},
		{name: "maximum strength hit and damage bonuses", rolls: []int64{6, 0}, strength: 31, wantDamage: 7},
		{name: "weapon and dexterity ring hit bonuses", rolls: []int64{7, 0, 0}, strength: 16, weapon: true, weaponPlus: 1, ring: "dexterity", ringPlus: 1, wantDamage: 4},
		{name: "increase damage ring", rolls: []int64{9, 0}, strength: 16, ring: "increase damage", ringPlus: 2, wantDamage: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := playerWithRolls(tt.rolls...)
			player.Strength = tt.strength
			if tt.weapon {
				weapon := item.NewItem(0, 0, item.ItemWeapon, "mace", 8)
				weapon.Enchantment = tt.weaponPlus
				player.Equipment.Weapon = weapon
			}
			if tt.ring != "" {
				ring := item.NewItem(0, 0, item.ItemRing, tt.ring, 0)
				ring.Enchantment = tt.ringPlus
				player.Equipment.RingLeft = ring
			}

			got := player.AttackRoll(10, true)
			if !got.Hit || got.Damage != tt.wantDamage {
				t.Fatalf("attack = hit %v, damage %d; want hit, damage %d", got.Hit, got.Damage, tt.wantDamage)
			}
		})
	}
}

func TestPlayerAttackUsesRogueWeaponDice(t *testing.T) {
	tests := []struct {
		weapon     string
		wantDamage int
	}{
		{weapon: "long sword", wantDamage: 4},
		{weapon: "short bow", wantDamage: 2},
		{weapon: "bow", wantDamage: 2},
		{weapon: "arrow", wantDamage: 2},
		{weapon: "dagger", wantDamage: 2},
		{weapon: "two handed sword", wantDamage: 5},
		{weapon: "two-handed sword", wantDamage: 5},
		{weapon: "dart", wantDamage: 2},
		{weapon: "shuriken", wantDamage: 2},
		{weapon: "spear", wantDamage: 3},
	}
	for _, tt := range tests {
		t.Run(tt.weapon, func(t *testing.T) {
			player := playerWithRolls(9, 0, 0, 0, 0)
			player.Equipment.Weapon = item.NewItem(0, 0, item.ItemWeapon, tt.weapon, 8)

			got := player.AttackRoll(10, true)
			if !got.Hit || got.Damage != tt.wantDamage {
				t.Fatalf("%q attack = hit %v, damage %d; want hit, damage %d", tt.weapon, got.Hit, got.Damage, tt.wantDamage)
			}
		})
	}
}

func TestPlayerAttackConsumesConfusionOnlyOnHit(t *testing.T) {
	player := playerWithRolls(8, 9, 0)
	player.CanConfuse = true
	player.CanConfuseTurns = 3

	miss := player.AttackRoll(10, true)
	if miss.Hit || miss.Confuses || !player.CanConfuse {
		t.Fatalf("miss = %+v with CanConfuse=%v, want miss without consuming confusion", miss, player.CanConfuse)
	}

	hit := player.AttackRoll(10, true)
	if !hit.Hit || !hit.Confuses || player.CanConfuse || player.CanConfuseTurns != 0 {
		t.Fatalf("hit = %+v, remaining confusion=%v/%d; want confused hit and consumed state", hit, player.CanConfuse, player.CanConfuseTurns)
	}
}

func TestPlayerTakeDamageResetsRecoveryQuietTurns(t *testing.T) {
	player := NewPlayer(0, 0)
	player.QuietTurns = 12

	player.TakeDamage(1)

	if player.QuietTurns != 0 {
		t.Fatalf("QuietTurns after damage = %d, want 0", player.QuietTurns)
	}
}

func TestPlayerRecoveryUsesRogueQuietThreshold(t *testing.T) {
	player := NewPlayer(0, 0)
	player.HP = 8

	if player.Recover(17) {
		t.Fatal("Recover healed before Rogue's level-one quiet threshold")
	}
	if !player.Recover(18) {
		t.Fatal("Recover did not heal at Rogue's level-one quiet threshold")
	}
	if player.HP != 9 {
		t.Fatalf("HP after the quiet threshold = %d, want 9", player.HP)
	}
	if player.QuietTurns != 0 {
		t.Fatalf("QuietTurns after recovery = %d, want 0", player.QuietTurns)
	}
}

func TestPlayerRecoveryUsesHighLevelRandomGainAndCannotRevive(t *testing.T) {
	player := playerWithRolls(0)
	player.Level = 8
	player.HP = 10
	player.MaxHP = 20

	if player.Recover(1) {
		t.Fatal("level-eight recovery happened before three quiet turns")
	}
	if !player.Recover(2) || player.HP != 11 {
		t.Fatalf("level-eight recovery HP = %d, want 11 after the threshold", player.HP)
	}

	player.HP = 0
	if player.Recover(100) || player.HP != 0 {
		t.Fatalf("dead player recovered: HP=%d", player.HP)
	}
}

func TestPlayerAdvanceStatusesExpiresSourceTimers(t *testing.T) {
	player := NewPlayer(0, 0)
	player.NoCommandTurns = 2
	player.NoMoveTurns = 1
	player.BlindTurns = 1
	player.HallucinationTurns = 2
	player.CanConfuse = true
	player.CanConfuseTurns = 1

	player.AdvanceStatuses()

	if player.NoCommandTurns != 1 || player.NoMoveTurns != 0 {
		t.Fatalf("forced-turn counters = %d/%d, want 1/0", player.NoCommandTurns, player.NoMoveTurns)
	}
	if player.BlindTurns != 0 || player.HallucinationTurns != 1 {
		t.Fatalf("status timers = blind %d, hallucination %d, want 0/1", player.BlindTurns, player.HallucinationTurns)
	}
	if player.CanConfuse || player.CanConfuseTurns != 0 {
		t.Fatalf("confuse-attack status remained active: %v/%d", player.CanConfuse, player.CanConfuseTurns)
	}
}

func TestPlayerArmorClassAndRustUseRogueEquipmentRules(t *testing.T) {
	player := NewPlayer(0, 0)
	armor := item.NewItem(0, 0, item.ItemArmor, "ring mail", 0)
	armor.Defense = 4
	armor.Enchantment = 1
	player.Equipment.Armor = armor
	protection := item.NewItem(0, 0, item.ItemRing, "protection", 0)
	protection.Enchantment = 1
	player.Equipment.RingLeft = protection

	if got := player.GetTotalDefense(); got != 4 {
		t.Fatalf("armor class with ring mail and protection = %d, want 4", got)
	}
	player.RustArmor()
	if armor.Defense != 3 || player.GetTotalDefense() != 5 {
		t.Fatalf("rusted armor/class = %d/%d, want 3/5", armor.Defense, player.GetTotalDefense())
	}

	armor.IsProtected = true
	player.RustArmor()
	if armor.Defense != 3 {
		t.Fatalf("protected armor defense = %d, want unchanged 3", armor.Defense)
	}
}

func TestPlayerSaveAgainstUsesRogueRollAndProtectionRing(t *testing.T) {
	tests := []struct {
		name       string
		roll       int64
		protection int
		wantSave   bool
	}{
		{name: "source save threshold", roll: 13, wantSave: true},
		{name: "below source save threshold", roll: 12, wantSave: false},
		{name: "protection ring lowers threshold", roll: 12, protection: 1, wantSave: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := playerWithRolls(tt.roll)
			if tt.protection != 0 {
				ring := item.NewItem(0, 0, item.ItemRing, "protection", 0)
				ring.Enchantment = tt.protection
				player.Equipment.RingLeft = ring
			}

			if got := player.SaveAgainst(SavingThrowPoison); got != tt.wantSave {
				t.Fatalf("SaveAgainst(VS_POISON) = %v, want %v", got, tt.wantSave)
			}
		})
	}
}

func TestReduceStrengthRespectsSustainStrengthRing(t *testing.T) {
	player := NewPlayerWithSeed(0, 0, 42)
	before := player.Strength
	player.Equipment.RingLeft = item.NewItem(0, 0, item.ItemRing, "sustain strength", 0)

	player.ReduceStrength(1)

	if player.Strength != before {
		t.Fatalf("strength with sustain strength ring = %d, want %d", player.Strength, before)
	}
}

type sequenceSource struct {
	values []int64
	index  int
}

func (s *sequenceSource) Int63() int64 {
	value := s.values[s.index%len(s.values)]
	s.index++
	return value << 32
}

func (s *sequenceSource) Seed(int64) {
	s.index = 0
}

func playerWithRolls(values ...int64) *Player {
	player := NewPlayer(0, 0)
	player.SetRandomSource(rand.New(&sequenceSource{values: values}))
	return player
}
