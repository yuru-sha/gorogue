package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/magic"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const invalidItemLetterMsg = "Invalid item letter. Use a-z."

// registerGameCommands registers PyRogue-style game commands
func (c *CLIMode) registerGameCommands() {
	gameCommands := []*Command{
		{
			Name:        "move",
			Description: "Move player in direction",
			Usage:       "move <direction>",
			Handler:     c.moveCommand,
		},
		{
			Name:        "pickup",
			Description: "Pick up item at current position",
			Usage:       "pickup [all]",
			Handler:     c.pickupCommand,
		},
		{
			Name:        "drop",
			Description: "Drop item from inventory",
			Usage:       "drop <item_letter>",
			Handler:     c.dropCommand,
		},
		{
			Name:        "equip",
			Description: "Equip item from inventory",
			Usage:       "equip <item_letter>",
			Handler:     c.equipCommand,
		},
		{
			Name:        "unequip",
			Description: "Unequip item",
			Usage:       "unequip <slot>",
			Handler:     c.unequipCommand,
		},
		{
			Name:        "use",
			Description: "Use item (potion/scroll)",
			Usage:       "use <item_letter>",
			Handler:     c.useCommand,
		},
		{
			Name:        "attack",
			Description: "Attack monster at position",
			Usage:       "attack <x> <y>",
			Handler:     c.attackCommand,
		},
		{
			Name:        "look",
			Description: "Look at position or examine item",
			Usage:       "look [x] [y]",
			Handler:     c.lookCommand,
		},
		{
			Name:        "rest",
			Description: "Rest for specified turns",
			Usage:       "rest [turns]",
			Handler:     c.restCommand,
		},
		{
			Name:        "search",
			Description: "Search for hidden doors/traps",
			Usage:       "search",
			Handler:     c.searchCommand,
		},
		{
			Name:        "open",
			Description: "Open door at direction",
			Usage:       "open <direction>",
			Handler:     c.openCommand,
		},
		{
			Name:        "close",
			Description: "Close door at direction",
			Usage:       "close <direction>",
			Handler:     c.closeCommand,
		},
		{
			Name:        "stairs",
			Description: "Use stairs (up/down)",
			Usage:       "stairs <up|down>",
			Handler:     c.stairsCommand,
		},
		{
			Name:        "auto",
			Description: "Auto-explore or auto-pickup",
			Usage:       "auto <explore|pickup>",
			Handler:     c.autoCommand,
		},
		{
			Name:        "game",
			Description: "Game control commands",
			Usage:       "game <new|save|load|quit>",
			Handler:     c.gameCommand,
		},
	}

	for _, cmd := range gameCommands {
		c.Commands[cmd.Name] = cmd
	}
}

// moveCommand moves the player
func (c *CLIMode) moveCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: move <direction>\nDirections: n, s, e, w, ne, nw, se, sw, up, down"
	}

	var dx, dy int
	direction := strings.ToLower(args[0])

	switch direction {
	case "n", "north", "up", "k":
		dx, dy = 0, -1
	case "s", "south", "down", "j":
		dx, dy = 0, 1
	case "e", "east", "right", "l":
		dx, dy = 1, 0
	case "w", "west", "left", "h":
		dx, dy = -1, 0
	case "ne", "northeast", "u":
		dx, dy = 1, -1
	case "nw", "northwest", "y":
		dx, dy = -1, -1
	case "se", "southeast", "m":
		dx, dy = 1, 1
	case "sw", "southwest", "b":
		dx, dy = -1, 1
	default:
		return fmt.Sprintf("Unknown direction: %s", direction)
	}

	oldX, oldY := c.Player.Position.X, c.Player.Position.Y
	newX, newY := oldX+dx, oldY+dy

	// Check bounds
	if newX < 0 || newX >= c.Level.Width || newY < 0 || newY >= c.Level.Height {
		return "Cannot move out of bounds"
	}

	// Check for walls (simplified)
	tile := c.Level.GetTile(newX, newY)
	if tile != nil && !tile.Walkable() {
		return "Cannot move there - blocked"
	}

	// Check for monsters
	monster := c.Level.GetMonsterAt(newX, newY)
	if monster != nil && monster.IsAlive() {
		// Attack instead of move
		damage := c.Player.CalculateDamage(monster.Defense)
		monster.TakeDamage(damage)

		if monster.IsAlive() {
			return fmt.Sprintf("Attacked %s for %d damage! (%d HP remaining)",
				monster.Type.Name, damage, monster.HP)
		} else {
			exp := monster.MaxHP + monster.Attack
			c.Player.GainExp(exp)
			return fmt.Sprintf("Killed %s! Gained %d experience.", monster.Type.Name, exp)
		}
	}

	// Move player
	c.Player.Position.X = newX
	c.Player.Position.Y = newY

	logger.Debug("Player moved via CLI", "from", fmt.Sprintf("(%d,%d)", oldX, oldY),
		"to", fmt.Sprintf("(%d,%d)", newX, newY))

	return fmt.Sprintf("Moved %s to (%d, %d)", direction, newX, newY)
}

// pickupCommand picks up items
func (c *CLIMode) pickupCommand(args []string) string {
	currentPos := c.Player.Position
	itemsAtPos := make([]*item.Item, 0)

	// Find items at current position
	for _, itm := range c.Level.Items {
		if itm.Position.X == currentPos.X && itm.Position.Y == currentPos.Y {
			itemsAtPos = append(itemsAtPos, itm)
		}
	}

	if len(itemsAtPos) == 0 {
		return "No items here to pick up."
	}

	if len(args) > 0 && args[0] == "all" {
		// Pick up all items
		pickedUp := 0
		for _, itm := range itemsAtPos {
			if c.Player.Inventory.AddItem(itm) {
				c.Level.RemoveItem(itm)
				pickedUp++
			}
		}
		return fmt.Sprintf("Picked up %d items.", pickedUp)
	}

	// Pick up first item
	itm := itemsAtPos[0]
	if c.Player.Inventory.AddItem(itm) {
		c.Level.RemoveItem(itm)
		displayName := c.Player.IdentifyMgr.GetDisplayName(itm)
		return fmt.Sprintf("Picked up %s.", displayName)
	}

	return "Inventory is full!"
}

// dropCommand drops items
func (c *CLIMode) dropCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: drop <item_letter>\nExample: drop a"
	}

	letter := args[0]
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		return invalidItemLetterMsg
	}

	index := int(letter[0] - 'a')
	itm := c.Player.Inventory.GetItem(index)
	if itm == nil {
		return fmt.Sprintf("No item at slot %s.", letter)
	}

	// Set item position to player position
	itm.Position.X = c.Player.Position.X
	itm.Position.Y = c.Player.Position.Y

	// Add to level items
	c.Level.Items = append(c.Level.Items, itm)

	// Remove from inventory
	c.Player.Inventory.RemoveItem(index)

	displayName := c.Player.IdentifyMgr.GetDisplayName(itm)
	return fmt.Sprintf("Dropped %s.", displayName)
}

// equipCommand equips items
func (c *CLIMode) equipCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: equip <item_letter>\nExample: equip a"
	}

	letter := args[0]
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		return invalidItemLetterMsg
	}

	index := int(letter[0] - 'a')
	itm := c.Player.Inventory.GetItem(index)
	if itm == nil {
		return fmt.Sprintf("No item at slot %s.", letter)
	}

	// Check if item can be equipped
	canEquip := false
	switch itm.Type {
	case item.ItemWeapon, item.ItemArmor, item.ItemRing:
		canEquip = true
	}

	if !canEquip {
		return "That item cannot be equipped."
	}

	// Try to equip
	if c.Player.Equipment.EquipItem(itm) {
		c.Player.Inventory.RemoveItem(index)
		displayName := c.Player.IdentifyMgr.GetDisplayName(itm)
		return fmt.Sprintf("Equipped %s.", displayName)
	}

	return "Cannot equip that item (slot occupied?)."
}

// unequipCommand unequips items
func (c *CLIMode) unequipCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: unequip <slot>\nSlots: weapon, armor, ring-left, ring-right"
	}

	slot := strings.ToLower(args[0])
	var slotName string

	switch slot {
	case "weapon", "w":
		slotName = "weapon"
	case "armor", "a":
		slotName = "armor"
	case "ring-left", "left", "l":
		slotName = "ring_left"
	case "ring-right", "right", "r":
		slotName = "ring_right"
	default:
		return "Unknown slot. Use: weapon, armor, ring-left, ring-right"
	}

	itm := c.Player.Equipment.UnequipItem(slotName)
	if itm == nil {
		return fmt.Sprintf("No item equipped in %s slot.", slot)
	}

	if c.Player.Inventory.AddItem(itm) {
		displayName := c.Player.IdentifyMgr.GetDisplayName(itm)
		return fmt.Sprintf("Unequipped %s.", displayName)
	}

	// Inventory full, re-equip item
	c.Player.Equipment.EquipItem(itm)
	return "Inventory is full! Cannot unequip."
}

// useCommand uses items
func (c *CLIMode) useCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: use <item_letter>\nExample: use a"
	}

	letter := args[0]
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		return invalidItemLetterMsg
	}

	index := int(letter[0] - 'a')
	itm := c.Player.Inventory.GetItem(index)
	if itm == nil {
		return fmt.Sprintf("No item at slot %s.", letter)
	}

	var result *magic.EffectResult

	switch itm.Type {
	case item.ItemPotion:
		result = magic.UsePotion(itm.Name, c.Player)
	case item.ItemScroll:
		result = magic.UseScroll(itm.Name, c.Player, c.Level)
	case item.ItemFood:
		c.Player.Hunger = 100
		result = &magic.EffectResult{
			Message:    "You eat the food and feel satisfied.",
			Success:    true,
			Identified: true,
		}
	default:
		return "That item cannot be used."
	}

	if result.Identified {
		c.Player.IdentifyMgr.IdentifyByUse(itm)
	}

	// Remove item from inventory
	c.Player.Inventory.RemoveItem(index)

	return result.Message
}

// attackCommand attacks monsters
func (c *CLIMode) attackCommand(args []string) string {
	if len(args) < 2 {
		return "Usage: attack <x> <y>"
	}

	x, err1 := strconv.Atoi(args[0])
	y, err2 := strconv.Atoi(args[1])

	if err1 != nil || err2 != nil {
		return "Invalid coordinates."
	}

	monster := c.Level.GetMonsterAt(x, y)
	if monster == nil || !monster.IsAlive() {
		return fmt.Sprintf("No monster at (%d, %d).", x, y)
	}

	damage := c.Player.CalculateDamage(monster.Defense)
	monster.TakeDamage(damage)

	if monster.IsAlive() {
		return fmt.Sprintf("Attacked %s for %d damage! (%d HP remaining)",
			monster.Type.Name, damage, monster.HP)
	} else {
		exp := monster.MaxHP + monster.Attack
		c.Player.GainExp(exp)
		return fmt.Sprintf("Killed %s! Gained %d experience.", monster.Type.Name, exp)
	}
}

// lookCommand examines positions or items
func (c *CLIMode) lookCommand(args []string) string {
	if len(args) == 0 {
		// Look at current position
		x, y := c.Player.Position.X, c.Player.Position.Y
		return c.describeTile(x, y)
	}

	if len(args) >= 2 {
		x, err1 := strconv.Atoi(args[0])
		y, err2 := strconv.Atoi(args[1])

		if err1 != nil || err2 != nil {
			return "Invalid coordinates."
		}

		return c.describeTile(x, y)
	}

	return "Usage: look [x] [y]"
}

// describeTile describes what's at a tile
func (c *CLIMode) describeTile(x, y int) string {
	if x < 0 || x >= c.Level.Width || y < 0 || y >= c.Level.Height {
		return "Out of bounds."
	}

	description := fmt.Sprintf("Position (%d, %d):\n", x, y)

	// Tile info
	tile := c.Level.GetTile(x, y)
	if tile != nil {
		description += fmt.Sprintf("Terrain: %s\n", tile.Type.String())
	}

	// Monster info
	monster := c.Level.GetMonsterAt(x, y)
	if monster != nil && monster.IsAlive() {
		description += fmt.Sprintf("Monster: %s (HP: %d/%d)\n",
			monster.Type.Name, monster.HP, monster.MaxHP)
	}

	// Item info
	itemsHere := make([]*item.Item, 0)
	for _, itm := range c.Level.Items {
		if itm.Position.X == x && itm.Position.Y == y {
			itemsHere = append(itemsHere, itm)
		}
	}

	if len(itemsHere) > 0 {
		description += "Items: "
		for i, itm := range itemsHere {
			if i > 0 {
				description += ", "
			}
			displayName := c.Player.IdentifyMgr.GetDisplayName(itm)
			description += displayName
		}
		description += "\n"
	}

	return strings.TrimSpace(description)
}

// restCommand rests for turns
func (c *CLIMode) restCommand(args []string) string {
	turns := 1
	if len(args) > 0 {
		if t, err := strconv.Atoi(args[0]); err == nil && t > 0 {
			turns = t
		}
	}

	if turns > 100 {
		turns = 100 // Safety limit
	}

	healAmount := turns / 5 // Heal slowly while resting
	if healAmount > 0 {
		c.Player.Heal(healAmount)
	}

	return fmt.Sprintf("Rested for %d turns. (Healed %d HP)", turns, healAmount)
}

// searchCommand searches for hidden things
func (c *CLIMode) searchCommand(args []string) string {
	// TODO: Implement search functionality
	return "You search carefully but find nothing hidden."
}

// openCommand opens doors
func (c *CLIMode) openCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: open <direction>"
	}

	// TODO: Implement door opening
	return fmt.Sprintf("Attempted to open door to the %s. (TODO: implement doors)", args[0])
}

// closeCommand closes doors
func (c *CLIMode) closeCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: close <direction>"
	}

	// TODO: Implement door closing
	return fmt.Sprintf("Attempted to close door to the %s. (TODO: implement doors)", args[0])
}

// stairsCommand uses stairs
func (c *CLIMode) stairsCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: stairs <up|down>"
	}

	direction := strings.ToLower(args[0])

	switch direction {
	case "up", "u":
		return "Climbed up the stairs. (TODO: implement level changing)"
	case "down", "d":
		return "Descended down the stairs. (TODO: implement level changing)"
	default:
		return "Usage: stairs <up|down>"
	}
}

// autoCommand provides automation
func (c *CLIMode) autoCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: auto <explore|pickup>"
	}

	action := strings.ToLower(args[0])

	switch action {
	case "explore":
		return "Auto-explore started. (TODO: implement auto-explore)"
	case "pickup":
		return c.pickupCommand([]string{"all"})
	default:
		return "Usage: auto <explore|pickup>"
	}
}

// gameCommand handles game control
func (c *CLIMode) gameCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: game <new|save|load|quit>"
	}

	action := strings.ToLower(args[0])

	switch action {
	case "new":
		return "Starting new game. (TODO: implement new game)"
	case "save":
		filename := "savegame.json"
		if len(args) > 1 {
			filename = args[1]
		}
		return fmt.Sprintf("Game saved to %s. (TODO: implement save)", filename)
	case "load":
		filename := "savegame.json"
		if len(args) > 1 {
			filename = args[1]
		}
		return fmt.Sprintf("Game loaded from %s. (TODO: implement load)", filename)
	case "quit":
		return "Use 'quit' or 'exit' to quit CLI."
	default:
		return "Usage: game <new|save|load|quit>"
	}
}
