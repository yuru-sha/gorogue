package item

import (
	"math/rand"
	"strings"
	"time"

	"github.com/yuru-sha/gorogue/internal/core/entity"
)

// ItemType represents different types of items
type ItemType int

const (
	ItemWeapon ItemType = iota
	ItemArmor
	ItemRing
	ItemScroll
	ItemPotion
	ItemWand
	ItemFood
	ItemGold
	ItemAmulet // イェンダーの魔除け
)

const (
	teleportationName          = "teleportation"
	weaponDamageOneDieOneSide  = "1x1"
	weaponDamageOneDieTwoSides = "1x2"
)

// Item represents an item in the game
type Item struct {
	*entity.Entity
	Type         ItemType
	Name         string
	RealName     string // canonical name, unchanged by identification
	Value        int    // gold value
	Quantity     int
	IsIdentified bool
	IsCursed     bool
	IsBlessed    bool
	IsProtected  bool
	Damage       int // maximum native melee damage for weapons
	Defense      int // positive defense bonus, derived from Rogue armor class
	Enchantment  int
	Charges      int
	MaxCharges   int
	ItemID       int
}

// NewItem creates a new item
func NewItem(x, y int, itemType ItemType, name string, value int) *Item {
	// Determine if item should start identified
	isIdentified := true
	switch itemType {
	case ItemScroll, ItemPotion, ItemRing, ItemWand:
		isIdentified = false // These need to be identified
	}

	return &Item{
		Entity:       entity.NewEntity(x, y),
		Type:         itemType,
		Name:         name,
		RealName:     name,
		Value:        value,
		Quantity:     1,
		IsIdentified: isIdentified,
		IsCursed:     false,
		IsBlessed:    false,
	}
}

// NewGold creates a source-sized gold pile for the first dungeon level.
func NewGold(x, y int, isSpecialRoom bool) *Item {
	return NewGoldWithRand(x, y, isSpecialRoom, nil)
}

// NewGoldWithRand preserves the existing API; Rogue uses the same floor-based
// gold formula in every room.
func NewGoldWithRand(x, y int, _ bool, rng *rand.Rand) *Item {
	return NewGoldForFloorWithRand(x, y, 1, rng)
}

// NewGoldForFloorWithRand creates gold using Rogue's GOLDCALC formula.
func NewGoldForFloorWithRand(x, y, floor int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	amount := rng.Intn(50+10*floor) + 2
	return NewItem(x, y, ItemGold, "Gold", amount)
}

// NewAmulet creates Yendor's amulet.
func NewAmulet(x, y int) *Item {
	return NewItem(x, y, ItemAmulet, "The Amulet of Yendor", 0)
}

// NewRandomScroll creates a random scroll.
func NewRandomScroll(x, y int) *Item {
	return NewRandomScrollWithRand(x, y, nil)
}

// NewRandomScrollWithRand creates a source-weighted Rogue scroll.
func NewRandomScrollWithRand(x, y int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	scroll := chooseItemDefinition(rng, ScrollTypes)
	return NewItem(x, y, ItemScroll, scroll.Name, scroll.Value)
}

// NewRandomPotion creates a random potion.
func NewRandomPotion(x, y int) *Item {
	return NewRandomPotionWithRand(x, y, nil)
}

// NewRandomPotionWithRand creates a source-weighted Rogue potion.
func NewRandomPotionWithRand(x, y int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	potion := chooseItemDefinition(rng, PotionTypes)
	return NewItem(x, y, ItemPotion, potion.Name, potion.Value)
}

// NewRandomRing creates a random ring.
func NewRandomRing(x, y int) *Item {
	return NewRandomRingWithRand(x, y, nil)
}

// NewRandomRingWithRand creates a source-weighted Rogue ring.
func NewRandomRingWithRand(x, y int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	ring := chooseItemDefinition(rng, RingTypes)
	result := NewItem(x, y, ItemRing, ring.Name, ring.Value)
	result.ItemID = ring.ID

	switch ring.Name {
	case "add strength", "protection", "dexterity", "increase damage":
		result.Enchantment = rng.Intn(3)
		if result.Enchantment == 0 {
			result.Enchantment = -1
			result.IsCursed = true
		}
	case "aggravate monster", teleportationName:
		result.IsCursed = true
	}
	return result
}

// NewFood creates a source-weighted food ration or slime-mold.
func NewFood(x, y int) *Item {
	return NewFoodWithRand(x, y, nil)
}

// NewFoodWithRand creates food using the supplied random source.
func NewFoodWithRand(x, y int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	name := FoodTypes[0].Name
	if rng.Intn(10) == 0 {
		name = FoodTypes[1].Name
	}
	return NewItem(x, y, ItemFood, name, 0)

}

// NewRandomThingWithRand creates one source-weighted generated item.
// Gold and the Amulet are placed separately in Rogue.
func NewRandomThingWithRand(x, y, floor int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	switch rollProbability(rng, ThingProbabilities[:]) {
	case 0:
		return NewRandomPotionWithRand(x, y, rng)
	case 1:
		return NewRandomScrollWithRand(x, y, rng)
	case 2:
		return NewFoodWithRand(x, y, rng)
	case 3:
		return NewRandomWeaponWithRand(x, y, floor, rng)
	case 4:
		return NewRandomArmorWithRand(x, y, floor, rng)
	case 5:
		return NewRandomRingWithRand(x, y, rng)
	default:
		return NewRandomWandWithRand(x, y, floor, rng)
	}
}

// NewRandomThingWithStateWithRand applies Rogue's cross-level food guarantee.
// noFood is the number of consecutive generated floors without food; the
// returned value resets to zero when the generated item is food.
func NewRandomThingWithStateWithRand(x, y, floor, noFood int, rng *rand.Rand) (generated *Item, consecutiveNoFood int) {
	rng = ensureRand(rng)
	if noFood > 3 {
		return NewFoodWithRand(x, y, rng), 0
	}
	generated = NewRandomThingWithRand(x, y, floor, rng)
	consecutiveNoFood = noFood
	if generated.Type == ItemFood {
		consecutiveNoFood = 0
	}
	return generated, consecutiveNoFood
}

func chooseItemDefinition(rng *rand.Rand, definitions []ItemDefinition) ItemDefinition {
	roll := rng.Intn(100)
	cumulative := 0
	for _, definition := range definitions {
		cumulative += definition.Probability
		if roll < cumulative {
			return definition
		}
	}
	return definitions[0]
}

func rollProbability(rng *rand.Rand, probabilities []int) int {
	roll := rng.Intn(100)
	cumulative := 0
	for i, probability := range probabilities {
		cumulative += probability
		if roll < cumulative {
			return i
		}
	}
	return 0
}

func ensureRand(rng *rand.Rand) *rand.Rand {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return rng
}

// ItemDefinition is a source-native item name, probability, and value.
type ItemDefinition struct {
	ID          int
	Name        string
	Probability int
	Value       int
}

// WeaponDefinition stores Rogue's weapon damage dice and appearance probability.
type WeaponDefinition struct {
	ID           int
	Name         string
	Damage       int
	Value        int
	MinFloor     int
	MaxFloor     int
	Probability  int
	MeleeDamage  string
	ThrownDamage string
	MinQuantity  int
	MaxQuantity  int
	LauncherID   int
}

// WeaponTypes follows weap_info and init_dam in Rogue 5.4.4.
var WeaponTypes = []WeaponDefinition{
	{102, "mace", 8, 8, 1, 26, 11, "2x4", "1x3", 1, 1, 0},
	{103, "long sword", 12, 15, 1, 26, 11, "3x4", weaponDamageOneDieTwoSides, 1, 1, 0},
	{104, "short bow", 1, 15, 1, 26, 12, weaponDamageOneDieOneSide, weaponDamageOneDieOneSide, 1, 1, 0},
	{107, "arrow", 1, 1, 1, 26, 12, weaponDamageOneDieOneSide, "2x3", 8, 15, 104},
	{101, "dagger", 6, 3, 1, 26, 8, "1x6", "1x4", 2, 5, 0},
	{106, "two handed sword", 16, 75, 1, 26, 10, "4x4", "1x2", 1, 1, 0},
	{108, "dart", 1, 2, 1, 26, 12, weaponDamageOneDieOneSide, "1x3", 8, 15, 0},
	{109, "shuriken", 2, 5, 1, 26, 12, "1x2", "2x4", 8, 15, 0},
	{110, "spear", 6, 5, 1, 26, 12, "2x3", "1x6", 1, 1, 0},
}

func WeaponForItem(itm *Item) (WeaponDefinition, bool) {
	if itm == nil {
		return WeaponDefinition{}, false
	}
	if itm.ItemID != 0 {
		for _, definition := range WeaponTypes {
			if definition.ID == itm.ItemID {
				return definition, true
			}
		}
	}
	name := strings.TrimSpace(itm.RealName)
	if name == "" {
		name = itm.Name
	}
	for _, definition := range WeaponTypes {
		if strings.EqualFold(name, definition.Name) {
			return definition, true
		}
	}
	return WeaponDefinition{}, false
}

// ArmorDefinition stores protection bonus (10 - Rogue armor class).
type ArmorDefinition struct {
	ID          int
	Name        string
	Defense     int
	Value       int
	MinFloor    int
	MaxFloor    int
	Probability int
	ArmorClass  int
}

// ArmorTypes follows arm_info and a_class in Rogue 5.4.4.
var ArmorTypes = []ArmorDefinition{
	{201, "leather armor", 2, 20, 1, 26, 20, 8},
	{203, "ring mail", 3, 25, 1, 26, 15, 7},
	{202, "studded leather armor", 3, 20, 1, 26, 15, 7},
	{204, "scale mail", 4, 30, 1, 26, 13, 6},
	{205, "chain mail", 5, 75, 1, 26, 12, 5},
	{206, "splint mail", 6, 80, 1, 26, 10, 4},
	{207, "banded mail", 6, 90, 1, 26, 10, 4},
	{208, "plate mail", 7, 150, 1, 26, 5, 3},
}

// PotionTypes follows pot_info in Rogue 5.4.4.
var PotionTypes = []ItemDefinition{
	{1, "confusion", 7, 5},
	{2, "hallucination", 8, 5},
	{3, "poison", 8, 5},
	{4, "gain strength", 13, 150},
	{5, "see invisible", 3, 100},
	{6, "healing", 13, 130},
	{7, "monster detection", 6, 130},
	{8, "magic detection", 6, 105},
	{9, "raise level", 2, 250},
	{10, "extra healing", 5, 200},
	{11, "haste self", 5, 190},
	{12, "restore strength", 13, 130},
	{13, "blindness", 5, 5},
	{14, "levitation", 6, 75},
}

// ScrollTypes follows scr_info in Rogue 5.4.4.
var ScrollTypes = []ItemDefinition{
	{1, "monster confusion", 7, 140},
	{2, "magic mapping", 4, 150},
	{3, "hold monster", 2, 180},
	{4, "sleep", 3, 5},
	{5, "enchant armor", 7, 160},
	{6, "identify potion", 10, 80},
	{7, "identify scroll", 10, 80},
	{8, "identify weapon", 6, 80},
	{9, "identify armor", 7, 100},
	{10, "identify ring, wand or staff", 10, 115},
	{11, "scare monster", 3, 200},
	{12, "food detection", 2, 60},
	{13, teleportationName, 5, 165},
	{14, "enchant weapon", 8, 150},
	{15, "create monster", 4, 75},
	{16, "remove curse", 7, 105},
	{17, "aggravate monsters", 3, 20},
	{18, "protect armor", 2, 250},
}

// RingTypes follows ring_info in Rogue 5.4.4.
var RingTypes = []ItemDefinition{
	{1, "protection", 9, 400},
	{2, "add strength", 9, 400},
	{3, "sustain strength", 5, 280},
	{4, "searching", 10, 420},
	{5, "see invisible", 10, 310},
	{6, "adornment", 1, 10},
	{7, "aggravate monster", 10, 10},
	{8, "dexterity", 8, 440},
	{9, "increase damage", 8, 400},
	{10, "regeneration", 4, 460},
	{11, "slow digestion", 9, 240},
	{12, teleportationName, 5, 30},
	{13, "stealth", 7, 470},
	{14, "maintain armor", 5, 380},
}

// FoodTypes follows Rogue's 90% ration / 10% slime-mold selection.
var FoodTypes = []ItemDefinition{
	{1, "food ration", 90, 0},
	{2, "slime-mold", 10, 0},
}

// ThingProbabilities follows things[] in Rogue 5.4.4, in its native order.
var ThingProbabilities = [...]int{26, 36, 16, 7, 7, 4, 4}

// WandDefinition follows ws_info and fix_stick in Rogue 5.4.4.
type WandDefinition struct {
	ID          int
	Name        string
	MinCharges  int
	MaxCharges  int
	Value       int
	MinFloor    int
	MaxFloor    int
	Probability int
}

var WandTypes = []WandDefinition{
	{601, "light", 10, 19, 250, 1, 26, 12},
	{609, "invisibility", 3, 7, 5, 1, 26, 6},
	{602, "lightning", 3, 7, 330, 1, 26, 3},
	{603, "fire", 3, 7, 330, 1, 26, 3},
	{604, "cold", 3, 7, 330, 1, 26, 3},
	{605, "polymorph", 3, 7, 310, 1, 26, 15},
	{606, "magic missile", 3, 7, 170, 1, 26, 10},
	{607, "haste monster", 3, 7, 5, 1, 26, 10},
	{608, "slow monster", 3, 7, 350, 1, 26, 11},
	{612, "drain life", 3, 7, 300, 1, 26, 9},
	{613, "nothing", 3, 7, 5, 1, 26, 1},
	{610, "teleport away", 3, 7, 340, 1, 26, 6},
	{614, "teleport to", 3, 7, 50, 1, 26, 6},
	{615, "cancellation", 3, 7, 280, 1, 26, 5},
}

// NewRandomWeapon creates a random weapon using source probabilities.
func NewRandomWeapon(x, y, floor int) *Item {
	return NewRandomWeaponWithRand(x, y, floor, nil)
}

// NewRandomWeaponWithRand creates a source-weighted weapon using the supplied RNG.
func NewRandomWeaponWithRand(x, y, _ int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	roll := rng.Intn(100)
	weapon := WeaponTypes[0]
	for _, candidate := range WeaponTypes {
		if roll < candidate.Probability {
			weapon = candidate
			break
		}
		roll -= candidate.Probability
	}

	result := NewItem(x, y, ItemWeapon, weapon.Name, weapon.Value)
	result.Damage = weapon.Damage
	result.ItemID = weapon.ID
	result.Quantity = weapon.MinQuantity
	if weapon.MaxQuantity > weapon.MinQuantity {
		result.Quantity += rng.Intn(weapon.MaxQuantity - weapon.MinQuantity + 1)
	}
	enchantmentRoll := rng.Intn(100)
	if enchantmentRoll < 10 {
		result.IsCursed = true
		result.Enchantment = -(rng.Intn(3) + 1)
	} else if enchantmentRoll < 15 {
		result.Enchantment = rng.Intn(3) + 1
	}
	return result
}

// NewRandomArmor creates random armor using source probabilities.
func NewRandomArmor(x, y, floor int) *Item {
	return NewRandomArmorWithRand(x, y, floor, nil)
}

// NewRandomArmorWithRand creates source-weighted armor using the supplied RNG.
func NewRandomArmorWithRand(x, y, floor int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	roll := rng.Intn(100)
	armor := ArmorTypes[0]
	for _, candidate := range ArmorTypes {
		if roll < candidate.Probability {
			armor = candidate
			break
		}
		roll -= candidate.Probability
	}

	result := NewItem(x, y, ItemArmor, armor.Name, armor.Value)
	result.Defense = armor.Defense
	result.ItemID = armor.ID
	enchantmentRoll := rng.Intn(100)
	if enchantmentRoll < 20 {
		result.IsCursed = true
		result.Enchantment = -(rng.Intn(3) + 1)
	} else if enchantmentRoll < 28 {
		result.Enchantment = rng.Intn(3) + 1
	}
	return result
}

// NewRandomWand creates a random wand using source probabilities.
func NewRandomWand(x, y, floor int) *Item {
	return NewRandomWandWithRand(x, y, floor, nil)
}

// NewRandomWandWithRand creates a source-weighted wand using the supplied RNG.
func NewRandomWandWithRand(x, y, floor int, rng *rand.Rand) *Item {
	rng = ensureRand(rng)
	roll := rng.Intn(100)
	wand := WandTypes[0]
	for _, candidate := range WandTypes {
		if roll < candidate.Probability {
			wand = candidate
			break
		}
		roll -= candidate.Probability
	}

	result := NewItem(x, y, ItemWand, wand.Name, wand.Value)
	result.Charges = wand.MinCharges + rng.Intn(wand.MaxCharges-wand.MinCharges+1)
	result.ItemID = wand.ID
	result.IsIdentified = false
	return result
}
