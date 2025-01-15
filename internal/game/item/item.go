package item

import (
	"math/rand"

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
	Type     ItemType
	Name     string
	Value    int // ゴールドとしての価値
	Quantity int
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

// GetItemColor returns the color for a given item type
func GetItemColor(t ItemType) [3]uint8 {
	switch t {
	case ItemAmulet:
		return [3]uint8{255, 215, 0} // Gold color
	default:
		return [3]uint8{255, 255, 255} // White
	}
}

// NewItem creates a new item
func NewItem(x, y int, itemType ItemType, name string, value int) *Item {
	return &Item{
		Entity:   entity.NewEntity(x, y, GetItemSymbol(itemType), GetItemColor(itemType)),
		Type:     itemType,
		Name:     name,
		Value:    value,
		Quantity: 1,
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
