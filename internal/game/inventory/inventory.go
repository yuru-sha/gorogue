package inventory

import (
	"fmt"

	"github.com/yuru-sha/gorogue/internal/game/identification"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	MaxInventorySize = 26 // PyRogue互換：A-Zの26個
)

// Inventory represents a player's inventory
type Inventory struct {
	Items    []*item.Item
	Capacity int
}

// Equipment represents equipped items
type Equipment struct {
	Weapon    *item.Item
	Armor     *item.Item
	RingLeft  *item.Item
	RingRight *item.Item
}

// NewInventory creates a new inventory
func NewInventory() *Inventory {
	return &Inventory{
		Items:    make([]*item.Item, 0, MaxInventorySize),
		Capacity: MaxInventorySize,
	}
}

// NewEquipment creates a new equipment set
func NewEquipment() *Equipment {
	return &Equipment{}
}

// AddItem adds an item to the inventory
func (inv *Inventory) AddItem(newItem *item.Item) bool {
	if len(inv.Items) >= inv.Capacity {
		logger.Debug("Inventory full", "capacity", inv.Capacity)
		return false
	}

	// Check if item can stack (for gold, food, etc.)
	if newItem.Type == item.ItemGold {
		// Try to stack with existing gold
		for _, existingItem := range inv.Items {
			if existingItem.Type == item.ItemGold {
				existingItem.Value += newItem.Value
				existingItem.Quantity += newItem.Quantity
				logger.Debug("Stacked gold",
					"total_value", existingItem.Value,
					"total_quantity", existingItem.Quantity,
				)
				return true
			}
		}
	}

	// Add as new item
	inv.Items = append(inv.Items, newItem)
	logger.Debug("Added item to inventory",
		"item", newItem.Name,
		"type", newItem.Type,
		"inventory_size", len(inv.Items),
	)
	return true
}

// RemoveItem removes an item from the inventory
func (inv *Inventory) RemoveItem(index int) *item.Item {
	if index < 0 || index >= len(inv.Items) {
		return nil
	}

	removedItem := inv.Items[index]
	inv.Items = append(inv.Items[:index], inv.Items[index+1:]...)

	logger.Debug("Removed item from inventory",
		"item", removedItem.Name,
		"inventory_size", len(inv.Items),
	)
	return removedItem
}

// GetItem returns an item by index
func (inv *Inventory) GetItem(index int) *item.Item {
	if index < 0 || index >= len(inv.Items) {
		return nil
	}
	return inv.Items[index]
}

// GetItemByLetter returns an item by letter (a-z)
func (inv *Inventory) GetItemByLetter(letter rune) (int, *item.Item) {
	if letter < 'a' || letter > 'z' {
		return -1, nil
	}

	index := int(letter - 'a')
	if index >= len(inv.Items) {
		return -1, nil
	}

	return index, inv.Items[index]
}

// IsFull checks if the inventory is full
func (inv *Inventory) IsFull() bool {
	return len(inv.Items) >= inv.Capacity
}

// IsEmpty checks if the inventory is empty
func (inv *Inventory) IsEmpty() bool {
	return len(inv.Items) == 0
}

// Size returns the current inventory size
func (inv *Inventory) Size() int {
	return len(inv.Items)
}

// HasItemType checks if the inventory contains an item of the specified type
func (inv *Inventory) HasItemType(itemType item.ItemType) bool {
	for _, itm := range inv.Items {
		if itm.Type == itemType {
			return true
		}
	}
	return false
}

// GetInventoryListing returns a formatted inventory listing
func (inv *Inventory) GetInventoryListing(identifyMgr *identification.IdentificationManager) []string {
	if inv.IsEmpty() {
		return []string{"Your pack is empty."}
	}

	listing := make([]string, 0, len(inv.Items)+1)
	listing = append(listing, "Current inventory:")

	for i, itm := range inv.Items {
		letter := rune('a' + i)
		var line string

		// 識別マネージャーを使用して表示名を取得
		if identifyMgr != nil {
			displayName := identifyMgr.GetDisplayName(itm)
			line = fmt.Sprintf("%c) %s", letter, displayName)
		} else {
			// フォールバック：識別マネージャーがない場合
			if itm.Type == item.ItemGold {
				line = fmt.Sprintf("%c) %d gold pieces", letter, itm.Value)
			} else {
				line = fmt.Sprintf("%c) %s", letter, itm.Name)
			}
		}

		listing = append(listing, line)
	}

	return listing
}

// EquipItem equips an item from inventory
func (eq *Equipment) EquipItem(itm *item.Item) bool {
	switch itm.Type {
	case item.ItemWeapon:
		eq.Weapon = itm
		logger.Debug("Equipped weapon", "weapon", itm.Name)
		return true
	case item.ItemArmor:
		eq.Armor = itm
		logger.Debug("Equipped armor", "armor", itm.Name)
		return true
	case item.ItemRing:
		if eq.RingLeft == nil {
			eq.RingLeft = itm
			logger.Debug("Equipped ring on left hand", "ring", itm.Name)
			return true
		} else if eq.RingRight == nil {
			eq.RingRight = itm
			logger.Debug("Equipped ring on right hand", "ring", itm.Name)
			return true
		} else {
			logger.Debug("Both ring slots occupied")
			return false
		}
	default:
		logger.Debug("Item cannot be equipped", "item", itm.Name, "type", itm.Type)
		return false
	}
}

// UnequipItem unequips an item by slot
func (eq *Equipment) UnequipItem(slot string) *item.Item {
	switch slot {
	case "weapon":
		if eq.Weapon != nil {
			item := eq.Weapon
			eq.Weapon = nil
			logger.Debug("Unequipped weapon", "weapon", item.Name)
			return item
		}
	case "armor":
		if eq.Armor != nil {
			item := eq.Armor
			eq.Armor = nil
			logger.Debug("Unequipped armor", "armor", item.Name)
			return item
		}
	case "ring_left":
		if eq.RingLeft != nil {
			item := eq.RingLeft
			eq.RingLeft = nil
			logger.Debug("Unequipped left ring", "ring", item.Name)
			return item
		}
	case "ring_right":
		if eq.RingRight != nil {
			item := eq.RingRight
			eq.RingRight = nil
			logger.Debug("Unequipped right ring", "ring", item.Name)
			return item
		}
	}
	return nil
}

const noneEquipped = "None"

// GetEquippedNames returns equipped item names for display
func (eq *Equipment) GetEquippedNames() (string, string, string, string) {
	weapon := noneEquipped
	armor := noneEquipped
	ringLeft := noneEquipped
	ringRight := noneEquipped

	if eq.Weapon != nil {
		weapon = eq.Weapon.Name
	}
	if eq.Armor != nil {
		armor = eq.Armor.Name
	}
	if eq.RingLeft != nil {
		ringLeft = eq.RingLeft.Name
	}
	if eq.RingRight != nil {
		ringRight = eq.RingRight.Name
	}

	return weapon, armor, ringLeft, ringRight
}

// GetAttackBonus returns attack bonus from equipped weapon
func (eq *Equipment) GetAttackBonus() int {
	if eq.Weapon != nil {
		return eq.Weapon.Value / 10 // Simple calculation
	}
	return 0
}

// GetDefenseBonus returns defense bonus from equipped armor
func (eq *Equipment) GetDefenseBonus() int {
	bonus := 0
	if eq.Armor != nil {
		bonus += eq.Armor.Value / 10
	}
	if eq.RingLeft != nil && eq.RingLeft.Type == item.ItemRing {
		bonus += eq.RingLeft.Value / 20
	}
	if eq.RingRight != nil && eq.RingRight.Type == item.ItemRing {
		bonus += eq.RingRight.Value / 20
	}
	return bonus
}
