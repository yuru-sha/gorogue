package item

import (
	"math/rand"

	"github.com/anaseto/gruid"
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

// Item represents an item in the game
type Item struct {
	*entity.Entity
	Type         ItemType
	Name         string
	RealName     string // 真の名前（識別前後で同じ）
	Value        int    // ゴールドとしての価値
	Quantity     int
	IsIdentified bool // このアイテムが識別済みかどうか
	IsCursed     bool // 呪われているかどうか
	IsBlessed    bool // 祝福されているかどうか
	// 武器・防具用
	Damage      int // 武器のダメージ
	Defense     int // 防具の防御力
	Enchantment int // 強化値 (+1, +2, etc.)
	// 杖用
	Charges    int // 杖の残りチャージ数
	MaxCharges int // 杖の最大チャージ数
	// 識別用
	ItemID int // PyRouge準拠のID
}

// GetItemSymbol returns the symbol for a given item type
func GetItemSymbol(t ItemType) rune {
	switch t {
	case ItemWeapon:
		return ')'
	case ItemArmor:
		return '['
	case ItemRing:
		return '='
	case ItemScroll:
		return '?'
	case ItemPotion:
		return '!'
	case ItemWand:
		return '/'
	case ItemFood:
		return '%'
	case ItemGold:
		return '$'
	case ItemAmulet:
		return '*'
	default:
		return '*'
	}
}

// GetItemColor returns the color for a given item type - PyRogue風
func GetItemColor(t ItemType) gruid.Color {
	switch t {
	case ItemWeapon:
		return 0xC0C0C0 // Silver - PyRogue風
	case ItemArmor:
		return 0x8B4513 // Brown - PyRogue風
	case ItemRing:
		return 0xFFD700 // Gold - PyRogue風
	case ItemScroll:
		return 0xFFFFFF // White - PyRogue風
	case ItemPotion:
		return 0xFF1493 // DeepPink - PyRogue風
	case ItemWand:
		return 0x8A2BE2 // BlueViolet - PyRogue風
	case ItemFood:
		return 0xFFA500 // Orange - PyRogue風
	case ItemGold:
		return 0xFFD700 // Gold - PyRogue風
	case ItemAmulet:
		return 0x9400D3 // Purple - PyRogue風（特別なアイテム）
	default:
		return 0xDA70D6 // Orchid - PyRogue風（デフォルト紫系）
	}
}

// NewItem creates a new item
func NewItem(x, y int, itemType ItemType, name string, value int) *Item {
	// Determine if item should start identified
	isIdentified := true
	switch itemType {
	case ItemScroll, ItemPotion, ItemRing:
		isIdentified = false // These need to be identified
	}

	return &Item{
		Entity:       entity.NewEntity(x, y, GetItemSymbol(itemType), GetItemColor(itemType)),
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

// NewGold creates a new gold pile with random amount
func NewGold(x, y int, isSpecialRoom bool) *Item {
	var amount int
	if isSpecialRoom {
		amount = 100 + rand.Intn(151) // 100-250
	} else {
		amount = 1 + rand.Intn(250) // 1-250
	}
	return NewItem(x, y, ItemGold, "Gold", amount)
}

// NewAmulet creates Yendor's amulet
func NewAmulet(x, y int) *Item {
	return NewItem(x, y, ItemAmulet, "イェンダーの魔除け", 1000)
}

// NewRandomScroll creates a random scroll
func NewRandomScroll(x, y int) *Item {
	scrollTypes := []string{
		"identify", "teleportation", "sleep", "enchant armor", "enchant weapon",
		"create monster", "remove curse", "aggravate monster", "magic mapping",
		"hold monster", "confuse monster", "scare monster", "blank paper",
		"light", "food detection", "gold detection", "potion detection",
		"magic detection", "monster detection", "trap detection",
	}

	scrollType := scrollTypes[rand.Intn(len(scrollTypes))]
	return NewItem(x, y, ItemScroll, scrollType, 50+rand.Intn(100))
}

// NewRandomPotion creates a random potion
func NewRandomPotion(x, y int) *Item {
	potionTypes := []string{
		"healing", "extra healing", "haste self", "restore strength", "blindness",
		"paralysis", "confusion", "hallucination", "poison", "gain strength",
		"see invisible", "gain experience", "thirst quenching", "magic detection",
		"monster detection", "object detection", "raise level", "gain dexterity",
		"gain constitution", "gain intelligence", "levitation", "invisibility",
	}

	potionType := potionTypes[rand.Intn(len(potionTypes))]
	return NewItem(x, y, ItemPotion, potionType, 25+rand.Intn(75))
}

// NewRandomRing creates a random ring
func NewRandomRing(x, y int) *Item {
	ringTypes := []string{
		"protection", "add strength", "sustain strength", "searching", "see invisible",
		"adornment", "teleportation", "stealth", "regeneration", "slow digestion",
		"dexterity", "increase damage", "protection from magic", "hunger",
		"aggravate monster", "maintain armor", "teleport control",
	}

	ringType := ringTypes[rand.Intn(len(ringTypes))]
	return NewItem(x, y, ItemRing, ringType, 100+rand.Intn(200))
}

// NewFood creates food item
func NewFood(x, y int) *Item {
	foodTypes := []string{"food ration", "slime-mold", "fruit"}
	foodType := foodTypes[rand.Intn(len(foodTypes))]
	return NewItem(x, y, ItemFood, foodType, 10+rand.Intn(20))
}

// WeaponDefinition represents a weapon type
type WeaponDefinition struct {
	ID       int
	Name     string
	Damage   int
	Value    int
	MinFloor int
	MaxFloor int
}

// PyRogue準拠の武器定義
var WeaponTypes = []WeaponDefinition{
	{101, "Dagger", 3, 10, 1, 26},
	{102, "Mace", 5, 15, 1, 26},
	{103, "Long Sword", 8, 25, 3, 26},
	{104, "Bow", 6, 20, 2, 26},
	{105, "Crossbow", 10, 30, 5, 26},
	{106, "Two-Handed Sword", 14, 75, 7, 26},
}

// ArmorDefinition represents an armor type
type ArmorDefinition struct {
	ID       int
	Name     string
	Defense  int
	Value    int
	MinFloor int
	MaxFloor int
}

// PyRogue準拠の防具定義
var ArmorTypes = []ArmorDefinition{
	{201, "Leather Armor", 2, 20, 1, 26},
	{202, "Studded Leather", 3, 30, 2, 26},
	{203, "Ring Mail", 4, 40, 3, 26},
	{204, "Scale Mail", 5, 50, 4, 26},
	{205, "Chain Mail", 6, 60, 5, 26},
	{206, "Splint Mail", 7, 70, 6, 26},
	{207, "Banded Mail", 8, 80, 7, 26},
	{208, "Plate Mail", 9, 90, 8, 26},
}

// WandDefinition represents a wand type
type WandDefinition struct {
	ID         int
	Name       string
	MinCharges int
	MaxCharges int
	Value      int
	MinFloor   int
	MaxFloor   int
}

// PyRogue準拠の杖定義
var WandTypes = []WandDefinition{
	{601, "Wand of Light", 10, 20, 120, 1, 26},
	{602, "Wand of Lightning", 4, 8, 200, 3, 26},
	{603, "Wand of Fire", 4, 8, 200, 3, 26},
	{604, "Wand of Cold", 4, 8, 200, 3, 26},
	{605, "Wand of Polymorph", 5, 10, 210, 5, 26},
	{606, "Wand of Magic Missile", 3, 6, 170, 2, 26},
	{607, "Wand of Haste Monster", 5, 10, 180, 4, 26},
	{608, "Wand of Slow Monster", 5, 10, 180, 4, 26},
	{609, "Wand of Invisibility", 5, 10, 190, 6, 26},
	{610, "Wand of Teleportation", 3, 6, 210, 5, 26},
	{611, "Wand of Sleep", 5, 10, 170, 3, 26},
	{612, "Wand of Drain Life", 2, 4, 280, 8, 26},
}

// NewRandomWeapon creates a random weapon appropriate for the floor
func NewRandomWeapon(x, y, floor int) *Item {
	var validWeapons []WeaponDefinition
	for _, weapon := range WeaponTypes {
		if floor >= weapon.MinFloor && floor <= weapon.MaxFloor {
			validWeapons = append(validWeapons, weapon)
		}
	}

	if len(validWeapons) == 0 {
		validWeapons = WeaponTypes // Fallback
	}

	weapon := validWeapons[rand.Intn(len(validWeapons))]
	item := NewItem(x, y, ItemWeapon, weapon.Name, weapon.Value)
	item.Damage = weapon.Damage
	item.ItemID = weapon.ID
	item.Enchantment = rand.Intn(3) - 1 // -1, 0, or +1
	return item
}

// NewRandomArmor creates a random armor appropriate for the floor
func NewRandomArmor(x, y, floor int) *Item {
	var validArmors []ArmorDefinition
	for _, armor := range ArmorTypes {
		if floor >= armor.MinFloor && floor <= armor.MaxFloor {
			validArmors = append(validArmors, armor)
		}
	}

	if len(validArmors) == 0 {
		validArmors = ArmorTypes // Fallback
	}

	armor := validArmors[rand.Intn(len(validArmors))]
	item := NewItem(x, y, ItemArmor, armor.Name, armor.Value)
	item.Defense = armor.Defense
	item.ItemID = armor.ID
	item.Enchantment = rand.Intn(3) - 1 // -1, 0, or +1
	return item
}

// NewRandomWand creates a random wand appropriate for the floor
func NewRandomWand(x, y, floor int) *Item {
	var validWands []WandDefinition
	for _, wand := range WandTypes {
		if floor >= wand.MinFloor && floor <= wand.MaxFloor {
			validWands = append(validWands, wand)
		}
	}

	if len(validWands) == 0 {
		validWands = WandTypes // Fallback
	}

	wand := validWands[rand.Intn(len(validWands))]
	item := NewItem(x, y, ItemWand, wand.Name, wand.Value)
	item.Charges = wand.MinCharges + rand.Intn(wand.MaxCharges-wand.MinCharges+1)
	item.MaxCharges = item.Charges
	item.ItemID = wand.ID
	item.IsIdentified = false // 杖は要識別
	return item
}
