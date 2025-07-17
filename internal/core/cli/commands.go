package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	equipmentWeapon = "weapon"
	equipmentArmor  = "armor"
	commandAll      = "all"
)

// CLIMode provides command-line interface for debugging and AI control
type CLIMode struct {
	IsActive bool
	Level    *dungeon.Level
	Player   *actor.Player
	Commands map[string]*Command
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Usage       string
	Handler     func(args []string) string
}

// NewCLIMode creates a new CLI mode instance
func NewCLIMode(level *dungeon.Level, player *actor.Player) *CLIMode {
	cli := &CLIMode{
		IsActive: false,
		Level:    level,
		Player:   player,
		Commands: make(map[string]*Command),
	}

	cli.registerCommands()
	cli.registerGameCommands()
	return cli
}

// registerCommands registers all available commands
func (c *CLIMode) registerCommands() {
	commands := []*Command{
		{
			Name:        "help",
			Description: "Show all available commands",
			Usage:       "help [command]",
			Handler:     c.helpCommand,
		},
		{
			Name:        "status",
			Description: "Show player status",
			Usage:       "status",
			Handler:     c.statusCommand,
		},
		{
			Name:        "heal",
			Description: "Heal player",
			Usage:       "heal [amount|full]",
			Handler:     c.healCommand,
		},
		{
			Name:        "gold",
			Description: "Add gold to player",
			Usage:       "gold <amount>",
			Handler:     c.goldCommand,
		},
		{
			Name:        "exp",
			Description: "Add experience to player",
			Usage:       "exp <amount>",
			Handler:     c.expCommand,
		},
		{
			Name:        "teleport",
			Description: "Teleport player to coordinates",
			Usage:       "teleport <x> <y>",
			Handler:     c.teleportCommand,
		},
		{
			Name:        "create",
			Description: "Create item at position",
			Usage:       "create <item_type> [x] [y]",
			Handler:     c.createCommand,
		},
		{
			Name:        "kill",
			Description: "Kill monsters",
			Usage:       "kill [all|<x> <y>]",
			Handler:     c.killCommand,
		},
		{
			Name:        "level",
			Description: "Change dungeon level",
			Usage:       "level <floor>",
			Handler:     c.levelCommand,
		},
		{
			Name:        "identify",
			Description: "Identify all items",
			Usage:       "identify [all|<item_type>]",
			Handler:     c.identifyCommand,
		},
		{
			Name:        "map",
			Description: "Reveal map or get map info",
			Usage:       "map [reveal|info]",
			Handler:     c.mapCommand,
		},
		{
			Name:        "inventory",
			Description: "Show or modify inventory",
			Usage:       "inventory [clear|add <item>]",
			Handler:     c.inventoryCommand,
		},
		{
			Name:        "save",
			Description: "Save game state",
			Usage:       "save [filename]",
			Handler:     c.saveCommand,
		},
		{
			Name:        "load",
			Description: "Load game state",
			Usage:       "load [filename]",
			Handler:     c.loadCommand,
		},
		{
			Name:        "set",
			Description: "Set player attributes",
			Usage:       "set <attribute> <value>",
			Handler:     c.setCommand,
		},
		{
			Name:        "spawn",
			Description: "Spawn monsters",
			Usage:       "spawn <monster_type> [x] [y] [count]",
			Handler:     c.spawnCommand,
		},
		{
			Name:        "debug",
			Description: "Debug information",
			Usage:       "debug [memory|performance|entities]",
			Handler:     c.debugCommand,
		},
	}

	for _, cmd := range commands {
		c.Commands[cmd.Name] = cmd
	}
}

// ExecuteCommand executes a CLI command
func (c *CLIMode) ExecuteCommand(input string) string {
	if !c.IsActive {
		return "CLI mode is not active"
	}

	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return "No command entered"
	}

	commandName := strings.ToLower(parts[0])
	args := parts[1:]

	if cmd, exists := c.Commands[commandName]; exists {
		logger.Info("CLI command executed", "command", commandName, "args", args)
		return cmd.Handler(args)
	}

	return fmt.Sprintf("Unknown command: %s. Type 'help' for available commands.", commandName)
}

// Toggle toggles CLI mode on/off
func (c *CLIMode) Toggle() {
	c.IsActive = !c.IsActive
	status := "OFF"
	if c.IsActive {
		status = "ON"
	}
	logger.Info("CLI mode toggled", "status", status)
}

// SetLevel updates the level reference
func (c *CLIMode) SetLevel(level *dungeon.Level) {
	c.Level = level
}

// helpCommand shows help information
func (c *CLIMode) helpCommand(args []string) string {
	if len(args) > 0 {
		// Show help for specific command
		if cmd, exists := c.Commands[args[0]]; exists {
			return fmt.Sprintf("%s: %s\nUsage: %s", cmd.Name, cmd.Description, cmd.Usage)
		}
		return fmt.Sprintf("Unknown command: %s", args[0])
	}

	// Show all commands
	result := "Available commands:\n"
	for name, cmd := range c.Commands {
		result += fmt.Sprintf("  %-12s - %s\n", name, cmd.Description)
	}
	result += "\nType 'help <command>' for detailed usage."
	return result
}

// statusCommand shows player status
func (c *CLIMode) statusCommand(args []string) string {
	return fmt.Sprintf("Player Status:\n"+
		"Level: %d\n"+
		"HP: %d/%d\n"+
		"Attack: %d\n"+
		"Defense: %d\n"+
		"Experience: %d\n"+
		"Gold: %d\n"+
		"Hunger: %d\n"+
		"Position: (%d, %d)\n"+
		"Inventory: %d/%d items",
		c.Player.Level, c.Player.HP, c.Player.MaxHP,
		c.Player.Attack, c.Player.Defense, c.Player.Exp,
		c.Player.Gold, c.Player.Hunger,
		c.Player.Position.X, c.Player.Position.Y,
		c.Player.Inventory.Size(), c.Player.Inventory.Capacity)
}

// healCommand heals the player
func (c *CLIMode) healCommand(args []string) string {
	if len(args) == 0 || args[0] == "full" {
		c.Player.HP = c.Player.MaxHP
		return "Player fully healed"
	}

	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return "Invalid heal amount"
	}

	oldHP := c.Player.HP
	c.Player.Heal(amount)
	healed := c.Player.HP - oldHP
	return fmt.Sprintf("Healed %d HP (was %d, now %d)", healed, oldHP, c.Player.HP)
}

// goldCommand adds gold to player
func (c *CLIMode) goldCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: gold <amount>"
	}

	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return "Invalid gold amount"
	}

	c.Player.AddGold(amount)
	return fmt.Sprintf("Added %d gold (total: %d)", amount, c.Player.Gold)
}

// expCommand adds experience to player
func (c *CLIMode) expCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: exp <amount>"
	}

	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return "Invalid experience amount"
	}

	oldLevel := c.Player.Level
	c.Player.GainExp(amount)

	if c.Player.Level > oldLevel {
		return fmt.Sprintf("Added %d exp, leveled up! (Level %d -> %d)", amount, oldLevel, c.Player.Level)
	}
	return fmt.Sprintf("Added %d exp (total: %d)", amount, c.Player.Exp)
}

// teleportCommand teleports player
func (c *CLIMode) teleportCommand(args []string) string {
	if len(args) < 2 {
		return "Usage: teleport <x> <y>"
	}

	x, err1 := strconv.Atoi(args[0])
	y, err2 := strconv.Atoi(args[1])

	if err1 != nil || err2 != nil {
		return "Invalid coordinates"
	}

	if x < 0 || x >= c.Level.Width || y < 0 || y >= c.Level.Height {
		return fmt.Sprintf("Coordinates out of bounds (0--%d, 0--%d)", c.Level.Width-1, c.Level.Height-1)
	}

	oldX, oldY := c.Player.Position.X, c.Player.Position.Y
	c.Player.Position.X = x
	c.Player.Position.Y = y

	return fmt.Sprintf("Teleported from (%d, %d) to (%d, %d)", oldX, oldY, x, y)
}

// createCommand creates items
func (c *CLIMode) createCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: create <item_type> [x] [y]\nTypes: weapon, armor, ring, scroll, potion, food, gold, amulet"
	}

	x := c.Player.Position.X
	y := c.Player.Position.Y

	if len(args) >= 3 {
		newX, err1 := strconv.Atoi(args[1])
		newY, err2 := strconv.Atoi(args[2])
		if err1 == nil && err2 == nil {
			x, y = newX, newY
		}
	}

	var newItem *item.Item
	itemType := strings.ToLower(args[0])

	switch itemType {
	case equipmentWeapon:
		newItem = item.NewItem(x, y, item.ItemWeapon, "debug sword", 100)
	case equipmentArmor:
		newItem = item.NewItem(x, y, item.ItemArmor, "debug armor", 100)
	case "ring":
		newItem = item.NewRandomRing(x, y)
	case "scroll":
		newItem = item.NewRandomScroll(x, y)
	case "potion":
		newItem = item.NewRandomPotion(x, y)
	case "food":
		newItem = item.NewFood(x, y)
	case "gold":
		newItem = item.NewGold(x, y, false)
	case "amulet":
		newItem = item.NewAmulet(x, y)
	default:
		return fmt.Sprintf("Unknown item type: %s", itemType)
	}

	c.Level.Items = append(c.Level.Items, newItem)
	return fmt.Sprintf("Created %s at (%d, %d)", newItem.Name, x, y)
}

// killCommand kills monsters
func (c *CLIMode) killCommand(args []string) string {
	if len(args) == 0 || args[0] == commandAll {
		count := 0
		for _, monster := range c.Level.Monsters {
			if monster.IsAlive() {
				monster.HP = 0
				count++
			}
		}
		return fmt.Sprintf("Killed %d monsters", count)
	}

	if len(args) >= 2 {
		x, err1 := strconv.Atoi(args[0])
		y, err2 := strconv.Atoi(args[1])
		if err1 != nil || err2 != nil {
			return "Invalid coordinates"
		}

		monster := c.Level.GetMonsterAt(x, y)
		if monster != nil && monster.IsAlive() {
			monster.HP = 0
			return fmt.Sprintf("Killed monster at (%d, %d)", x, y)
		}
		return fmt.Sprintf("No alive monster at (%d, %d)", x, y)
	}

	return "Usage: kill [all|<x> <y>]"
}

// levelCommand changes dungeon level
func (c *CLIMode) levelCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: level <floor>"
	}

	floor, err := strconv.Atoi(args[0])
	if err != nil {
		return "Invalid floor number"
	}

	// TODO: Implement level changing
	return fmt.Sprintf("Level change to floor %d (TODO: implement)", floor)
}

// identifyCommand identifies items
func (c *CLIMode) identifyCommand(args []string) string {
	if len(args) == 0 || args[0] == commandAll {
		// Identify all items in inventory
		count := 0
		for _, item := range c.Player.Inventory.Items {
			if !c.Player.IdentifyMgr.IsIdentified(item) {
				c.Player.IdentifyMgr.IdentifyItem(item)
				count++
			}
		}
		return fmt.Sprintf("Identified %d items", count)
	}

	return "Usage: identify [all|<item_type>]"
}

// mapCommand provides map information
func (c *CLIMode) mapCommand(args []string) string {
	if len(args) == 0 || args[0] == "info" {
		return fmt.Sprintf("Map Info:\nSize: %dx%d\nRooms: %d\nMonsters: %d\nItems: %d",
			c.Level.Width, c.Level.Height,
			len(c.Level.Rooms), len(c.Level.Monsters), len(c.Level.Items))
	}

	if args[0] == "reveal" {
		// TODO: Implement map revelation
		return "Map revealed (TODO: implement)"
	}

	return "Usage: map [info|reveal]"
}

// inventoryCommand manages inventory
func (c *CLIMode) inventoryCommand(args []string) string {
	if len(args) == 0 {
		listing := c.Player.Inventory.GetInventoryListing(c.Player.IdentifyMgr)
		return strings.Join(listing, "\n")
	}

	if args[0] == "clear" {
		count := c.Player.Inventory.Size()
		c.Player.Inventory.Items = make([]*item.Item, 0)
		return fmt.Sprintf("Cleared %d items from inventory", count)
	}

	return "Usage: inventory [clear|add <item>]"
}

// saveCommand saves game state
func (c *CLIMode) saveCommand(args []string) string {
	filename := "debug_save.json"
	if len(args) > 0 {
		filename = args[0]
	}

	// TODO: Implement save functionality
	return fmt.Sprintf("Game saved to %s (TODO: implement)", filename)
}

// loadCommand loads game state
func (c *CLIMode) loadCommand(args []string) string {
	filename := "debug_save.json"
	if len(args) > 0 {
		filename = args[0]
	}

	// TODO: Implement load functionality
	return fmt.Sprintf("Game loaded from %s (TODO: implement)", filename)
}

// setCommand sets player attributes
func (c *CLIMode) setCommand(args []string) string {
	if len(args) < 2 {
		return "Usage: set <attribute> <value>\nAttributes: level, hp, maxhp, attack, defense, hunger"
	}

	attribute := strings.ToLower(args[0])
	value, err := strconv.Atoi(args[1])
	if err != nil {
		return "Invalid value"
	}

	switch attribute {
	case "level":
		c.Player.Level = value
	case "hp":
		c.Player.HP = value
	case "maxhp":
		c.Player.MaxHP = value
	case "attack":
		c.Player.Attack = value
	case "defense":
		c.Player.Defense = value
	case "hunger":
		c.Player.Hunger = value
	default:
		return fmt.Sprintf("Unknown attribute: %s", attribute)
	}

	return fmt.Sprintf("Set %s to %d", attribute, value)
}

// spawnCommand spawns monsters
func (c *CLIMode) spawnCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: spawn <monster_type> [x] [y] [count]"
	}

	// TODO: Implement monster spawning
	return fmt.Sprintf("Spawned %s (TODO: implement)", args[0])
}

// debugCommand provides debug information
func (c *CLIMode) debugCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: debug [memory|performance|entities]"
	}

	switch args[0] {
	case "memory":
		return "Memory usage: TODO"
	case "performance":
		return "Performance stats: TODO"
	case "entities":
		return fmt.Sprintf("Entities:\nPlayer: (%d, %d)\nMonsters: %d\nItems: %d",
			c.Player.Position.X, c.Player.Position.Y,
			len(c.Level.Monsters), len(c.Level.Items))
	default:
		return "Unknown debug option"
	}
}
