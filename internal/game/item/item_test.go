package item

import (
	"math/rand"
	"testing"
)

func TestSourceItemTables(t *testing.T) {
	weapons := []struct {
		name, melee, thrown                          string
		probability, value, minQuantity, maxQuantity int
	}{
		{"mace", "2x4", "1x3", 11, 8, 1, 1},
		{"long sword", "3x4", "1x2", 11, 15, 1, 1},
		{"short bow", "1x1", "1x1", 12, 15, 1, 1},
		{"arrow", "1x1", "2x3", 12, 1, 8, 15},
		{"dagger", "1x6", "1x4", 8, 3, 2, 5},
		{"two handed sword", "4x4", "1x2", 10, 75, 1, 1},
		{"dart", "1x1", "1x3", 12, 2, 8, 15},
		{"shuriken", "1x2", "2x4", 12, 5, 8, 15},
		{"spear", "2x3", "1x6", 12, 5, 1, 1},
	}
	if len(WeaponTypes) != len(weapons) {
		t.Fatalf("weapon table has %d entries, want %d", len(WeaponTypes), len(weapons))
	}
	for i, want := range weapons {
		got := WeaponTypes[i]
		if got.Name != want.name || got.MeleeDamage != want.melee || got.ThrownDamage != want.thrown || got.Probability != want.probability || got.Value != want.value || got.MinQuantity != want.minQuantity || got.MaxQuantity != want.maxQuantity {
			t.Errorf("weapon %d = %+v, want name=%q melee=%q thrown=%q probability=%d value=%d quantity=%d..%d", i, got, want.name, want.melee, want.thrown, want.probability, want.value, want.minQuantity, want.maxQuantity)
		}
	}

	armors := []struct {
		name                        string
		probability, value, defense int
	}{
		{"leather armor", 20, 20, 2},
		{"ring mail", 15, 25, 3},
		{"studded leather armor", 15, 20, 3},
		{"scale mail", 13, 30, 4},
		{"chain mail", 12, 75, 5},
		{"splint mail", 10, 80, 6},
		{"banded mail", 10, 90, 6},
		{"plate mail", 5, 150, 7},
	}
	if len(ArmorTypes) != len(armors) {
		t.Fatalf("armor table has %d entries, want %d", len(ArmorTypes), len(armors))
	}
	for i, want := range armors {
		got := ArmorTypes[i]
		if got.Name != want.name || got.Probability != want.probability || got.Value != want.value || got.Defense != want.defense {
			t.Errorf("armor %d = %+v, want %+v", i, got, want)
		}
	}

	potions := []struct {
		name               string
		probability, value int
	}{
		{"confusion", 7, 5}, {"hallucination", 8, 5}, {"poison", 8, 5}, {"gain strength", 13, 150},
		{"see invisible", 3, 100}, {"healing", 13, 130}, {"monster detection", 6, 130}, {"magic detection", 6, 105},
		{"raise level", 2, 250}, {"extra healing", 5, 200}, {"haste self", 5, 190}, {"restore strength", 13, 130},
		{"blindness", 5, 5}, {"levitation", 6, 75},
	}
	assertNamedValues(t, "potion", len(PotionTypes), potions, func(i int) string { return PotionTypes[i].Name }, func(i int) int { return PotionTypes[i].Probability }, func(i int) int { return PotionTypes[i].Value })

	scrolls := []struct {
		name               string
		probability, value int
	}{
		{"monster confusion", 7, 140}, {"magic mapping", 4, 150}, {"hold monster", 2, 180}, {"sleep", 3, 5},
		{"enchant armor", 7, 160}, {"identify potion", 10, 80}, {"identify scroll", 10, 80}, {"identify weapon", 6, 80},
		{"identify armor", 7, 100}, {"identify ring, wand or staff", 10, 115}, {"scare monster", 3, 200}, {"food detection", 2, 60},
		{"teleportation", 5, 165}, {"enchant weapon", 8, 150}, {"create monster", 4, 75}, {"remove curse", 7, 105},
		{"aggravate monsters", 3, 20}, {"protect armor", 2, 250},
	}
	assertNamedValues(t, "scroll", len(ScrollTypes), scrolls, func(i int) string { return ScrollTypes[i].Name }, func(i int) int { return ScrollTypes[i].Probability }, func(i int) int { return ScrollTypes[i].Value })

	rings := []struct {
		name               string
		probability, value int
	}{
		{"protection", 9, 400}, {"add strength", 9, 400}, {"sustain strength", 5, 280}, {"searching", 10, 420},
		{"see invisible", 10, 310}, {"adornment", 1, 10}, {"aggravate monster", 10, 10}, {"dexterity", 8, 440},
		{"increase damage", 8, 400}, {"regeneration", 4, 460}, {"slow digestion", 9, 240}, {"teleportation", 5, 30},
		{"stealth", 7, 470}, {"maintain armor", 5, 380},
	}
	assertNamedValues(t, "ring", len(RingTypes), rings, func(i int) string { return RingTypes[i].Name }, func(i int) int { return RingTypes[i].Probability }, func(i int) int { return RingTypes[i].Value })

	wands := []struct {
		name               string
		probability, value int
	}{
		{"light", 12, 250}, {"invisibility", 6, 5}, {"lightning", 3, 330}, {"fire", 3, 330}, {"cold", 3, 330},
		{"polymorph", 15, 310}, {"magic missile", 10, 170}, {"haste monster", 10, 5}, {"slow monster", 11, 350},
		{"drain life", 9, 300}, {"nothing", 1, 5}, {"teleport away", 6, 340}, {"teleport to", 6, 50}, {"cancellation", 5, 280},
	}
	assertNamedValues(t, "wand", len(WandTypes), wands, func(i int) string { return WandTypes[i].Name }, func(i int) int { return WandTypes[i].Probability }, func(i int) int { return WandTypes[i].Value })
	for _, wand := range WandTypes {
		min, max := 3, 7
		if wand.Name == "light" {
			min, max = 10, 19
		}
		if wand.MinCharges != min || wand.MaxCharges != max {
			t.Errorf("wand %q charges = %d..%d, want %d..%d", wand.Name, wand.MinCharges, wand.MaxCharges, min, max)
		}
	}
}

func assertNamedValues(t *testing.T, category string, count int, want []struct {
	name               string
	probability, value int
}, nameAt func(int) string, probabilityAt, valueAt func(int) int) {
	t.Helper()
	if count != len(want) {
		t.Fatalf("%s table has %d entries, want %d", category, count, len(want))
	}
	for i, expected := range want {
		if name, probability, value := nameAt(i), probabilityAt(i), valueAt(i); name != expected.name || probability != expected.probability || value != expected.value {
			t.Errorf("%s %d = %q, probability %d, value %d; want %q, probability %d, value %d", category, i, name, probability, value, expected.name, expected.probability, expected.value)
		}
	}
}

func TestSourceFoodAndCategoryProbabilities(t *testing.T) {
	if len(FoodTypes) != 2 || FoodTypes[0].Name != "food ration" || FoodTypes[0].Probability != 90 || FoodTypes[1].Name != "slime-mold" || FoodTypes[1].Probability != 10 {
		t.Fatalf("food table = %+v, want food ration 90%% and slime-mold 10%%", FoodTypes)
	}
	want := []int{26, 36, 16, 7, 7, 4, 4}
	if len(ThingProbabilities) != len(want) {
		t.Fatalf("category probabilities = %v, want %v", ThingProbabilities, want)
	}
	for i, probability := range want {
		if ThingProbabilities[i] != probability {
			t.Errorf("category probability %d = %d, want %d", i, ThingProbabilities[i], probability)
		}
	}
}

func TestRandomConstructorsUseOnlySourceItemsAndSourceQuantities(t *testing.T) {
	weaponNames := map[string]bool{}
	armorNames := map[string]bool{}
	wandNames := map[string]bool{}
	potionNames := map[string]bool{}
	scrollNames := map[string]bool{}
	ringNames := map[string]bool{}
	foodNames := map[string]bool{}
	for seed := int64(0); seed < 2000; seed++ {
		weapon := NewRandomWeaponWithRand(0, 0, 26, rand.New(rand.NewSource(seed)))
		weaponNames[weapon.Name] = true
		minQuantity, maxQuantity := 1, 1
		switch weapon.Name {
		case "dagger":
			minQuantity, maxQuantity = 2, 5
		case "arrow", "dart", "shuriken":
			minQuantity, maxQuantity = 8, 15
		}
		if weapon.Quantity < minQuantity || weapon.Quantity > maxQuantity || weapon.Damage < 1 {
			t.Errorf("%s quantity/damage = %d/%d, want quantity %d..%d and positive damage", weapon.Name, weapon.Quantity, weapon.Damage, minQuantity, maxQuantity)
		}
		if weapon.IsCursed != (weapon.Enchantment < 0) || weapon.Enchantment < -3 || weapon.Enchantment > 3 {
			t.Errorf("%s curse/enchantment = %t/%d, outside source generation", weapon.Name, weapon.IsCursed, weapon.Enchantment)
		}
		armor := NewRandomArmorWithRand(0, 0, 26, rand.New(rand.NewSource(seed)))
		armorNames[armor.Name] = true
		baseDefense := 0
		for _, definition := range ArmorTypes {
			if definition.Name == armor.Name {
				baseDefense = definition.Defense
				break
			}
		}
		if armor.Defense != baseDefense || armor.IsCursed != (armor.Enchantment < 0) {
			t.Errorf("%s base defense/enchantment/curse = %d/%d/%t, want base %d with separate source enchantment", armor.Name, armor.Defense, armor.Enchantment, armor.IsCursed, baseDefense)
		}
		wand := NewRandomWandWithRand(0, 0, 26, rand.New(rand.NewSource(seed)))
		wandNames[wand.Name] = true
		minCharges, maxCharges := 3, 7
		if wand.Name == "light" {
			minCharges, maxCharges = 10, 19
		}
		if wand.MaxCharges != 0 || wand.Charges < minCharges || wand.Charges > maxCharges {
			t.Errorf("%s charges = %d/%d, want %d..%d and no non-native maximum", wand.Name, wand.Charges, wand.MaxCharges, minCharges, maxCharges)
		}
		potionNames[NewRandomPotionWithRand(0, 0, rand.New(rand.NewSource(seed))).Name] = true
		scrollNames[NewRandomScrollWithRand(0, 0, rand.New(rand.NewSource(seed))).Name] = true
		ring := NewRandomRingWithRand(0, 0, rand.New(rand.NewSource(seed)))
		ringNames[ring.Name] = true
		curseExpected := ring.Name == "aggravate monster" || ring.Name == "teleportation" ||
			(ring.Enchantment == -1 && (ring.Name == "protection" || ring.Name == "add strength" || ring.Name == "dexterity" || ring.Name == "increase damage"))
		if ring.IsCursed != curseExpected {
			t.Errorf("ring %q curse/enchantment = %t/%d, want cursed=%t", ring.Name, ring.IsCursed, ring.Enchantment, curseExpected)
		}
		food := NewFoodWithRand(0, 0, rand.New(rand.NewSource(seed)))
		foodNames[food.Name] = true
		if food.Quantity != 1 || food.Value != 0 {
			t.Errorf("food %q quantity/value = %d/%d, want 1/0", food.Name, food.Quantity, food.Value)
		}
		thing := NewRandomThingWithRand(0, 0, 26, rand.New(rand.NewSource(seed)))
		if thing.Type == ItemGold || thing.Type == ItemAmulet || thing.Type < ItemWeapon || thing.Type > ItemFood {
			t.Errorf("source-generated item has non-native category %d", thing.Type)
		}
	}
	for _, name := range []string{"mace", "long sword", "short bow", "arrow", "dagger", "two handed sword", "dart", "shuriken", "spear"} {
		if !weaponNames[name] {
			t.Errorf("source weapon %q was never generated", name)
		}
	}
	for _, name := range []string{"leather armor", "ring mail", "studded leather armor", "scale mail", "chain mail", "splint mail", "banded mail", "plate mail"} {
		if !armorNames[name] {
			t.Errorf("source armor %q was never generated", name)
		}
	}
	for _, name := range []string{"light", "invisibility", "lightning", "fire", "cold", "polymorph", "magic missile", "haste monster", "slow monster", "drain life", "nothing", "teleport away", "teleport to", "cancellation"} {
		if !wandNames[name] {
			t.Errorf("source wand %q was never generated", name)
		}
	}
	for _, name := range []string{"confusion", "hallucination", "poison", "gain strength", "see invisible", "healing", "monster detection", "magic detection", "raise level", "extra healing", "haste self", "restore strength", "blindness", "levitation"} {
		if !potionNames[name] {
			t.Errorf("source potion %q was never generated", name)
		}
	}
	for _, name := range []string{"monster confusion", "magic mapping", "hold monster", "sleep", "enchant armor", "identify potion", "identify scroll", "identify weapon", "identify armor", "identify ring, wand or staff", "scare monster", "food detection", "teleportation", "enchant weapon", "create monster", "remove curse", "aggravate monsters", "protect armor"} {
		if !scrollNames[name] {
			t.Errorf("source scroll %q was never generated", name)
		}
	}
	for _, name := range []string{"protection", "add strength", "sustain strength", "searching", "see invisible", "adornment", "aggravate monster", "dexterity", "increase damage", "regeneration", "slow digestion", "teleportation", "stealth", "maintain armor"} {
		if !ringNames[name] {
			t.Errorf("source ring %q was never generated", name)
		}
	}
	if len(foodNames) != 2 || !foodNames["food ration"] || !foodNames["slime-mold"] {
		t.Errorf("generated food names = %v, want only food ration and slime-mold", foodNames)
	}
}

func TestGoldAndAmuletUseRogueValues(t *testing.T) {
	for _, floor := range []int{1, 26} {
		for seed := int64(0); seed < 100; seed++ {
			amount := NewGoldForFloorWithRand(0, 0, floor, rand.New(rand.NewSource(seed))).Value
			maximum := 51 + 10*floor
			if amount < 2 || amount > maximum {
				t.Fatalf("floor %d gold amount = %d, want 2..%d", floor, amount, maximum)
			}
		}
	}
	amulet := NewAmulet(0, 0)
	if amulet.Name != "The Amulet of Yendor" || amulet.Value != 0 {
		t.Fatalf("amulet = %q worth %d, want Rogue amulet with no item gold value", amulet.Name, amulet.Value)
	}
}

func TestForcedRationAfterFourFoodlessFloors(t *testing.T) {
	forced, nextNoFood := NewRandomThingWithStateWithRand(1, 1, 5, 4, rand.New(rand.NewSource(1)))
	if forced.Type != ItemFood || nextNoFood != 0 {
		t.Fatalf("forced generation = (%v, no_food=%d), want food and reset counter", forced.Type, nextNoFood)
	}

	seed := int64(0)
	for {
		candidate := NewRandomThingWithRand(1, 1, 4, rand.New(rand.NewSource(seed)))
		if candidate.Type != ItemFood {
			break
		}
		seed++
	}
	generated, nextNoFood := NewRandomThingWithStateWithRand(1, 1, 4, 3, rand.New(rand.NewSource(seed)))
	if generated.Type == ItemFood || nextNoFood != 3 {
		t.Fatalf("ordinary generation = (%v, no_food=%d), want non-food and unchanged counter", generated.Type, nextNoFood)
	}
}
