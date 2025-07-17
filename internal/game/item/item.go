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
	case ItemFood:
		return '%'
	case ItemGold:
		return '$'

	case ItemAmulet:
		return '&'
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
