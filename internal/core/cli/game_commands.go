package cli

import (
	"fmt"
	"strings"

	gamecommand "github.com/yuru-sha/gorogue/internal/core/command"
)

// registerGameCommands registers PyRogue-style game commands.
func (c *CLIMode) registerGameCommands() {
	gameCommands := []*Command{
		{Name: "move", Description: "Move player in direction", Usage: "move <direction>", Handler: c.moveCommand},
		{Name: "pickup", Description: "Pick up item at current position", Usage: "pickup [all]", Handler: c.pickupCommand},
		{Name: "drop", Description: "Drop item from inventory", Usage: "drop <item_letter>", Handler: c.dropCommand},
		{Name: "equip", Description: "Equip item from inventory", Usage: "equip <item_letter>", Handler: c.equipCommand},
		{Name: "unequip", Description: "Unequip item", Usage: "unequip <slot>", Handler: c.unequipCommand},
		{Name: "use", Description: "Use item (potion/scroll/food)", Usage: "use <item_letter>", Handler: c.useCommand},
		{Name: "attack", Description: "Attack monster at position", Usage: "attack <x> <y>", Handler: c.attackCommand},
		{Name: "look", Description: "Look at position or examine item", Usage: "look [x] [y]", Handler: c.lookCommand},
		{Name: "rest", Description: "Rest for specified turns", Usage: "rest [turns]", Handler: c.restCommand},
		{Name: "search", Description: "Search for hidden doors/traps", Usage: "search", Handler: c.searchCommand},
		{Name: "open", Description: "Open door at direction", Usage: "open <direction>", Handler: c.openCommand},
		{Name: "close", Description: "Close door at direction", Usage: "close <direction>", Handler: c.closeCommand},
		{Name: "stairs", Description: "Use stairs (up/down)", Usage: "stairs <up|down>", Handler: c.stairsCommand},
		{Name: "auto", Description: "Auto-explore or auto-pickup", Usage: "auto <explore|pickup>", Handler: c.autoCommand},
		{Name: "game", Description: "Game control commands", Usage: "game <new|save|load|quit>", Handler: c.gameCommand},
	}

	for _, cmd := range gameCommands {
		c.Commands[cmd.Name] = cmd
	}
}

func (c *CLIMode) moveCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: move <direction>\nDirections: n, s, e, w, ne, nw, se, sw"
	}
	direction, ok := gamecommand.ParseDirection(args[0])
	if !ok {
		return fmt.Sprintf("Unknown direction: %s", strings.ToLower(args[0]))
	}
	return c.executeGameplay(gamecommand.NewMoveCommand(direction))
}

func (c *CLIMode) pickupCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdPickUp}, args...)
}

func (c *CLIMode) dropCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdDrop}, args...)
}

func (c *CLIMode) equipCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdEquip}, args...)
}

func (c *CLIMode) unequipCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdUnequip}, args...)
}

func (c *CLIMode) useCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdUse}, args...)
}

func (c *CLIMode) attackCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdFight}, args...)
}

func (c *CLIMode) lookCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdLook}, args...)
}

func (c *CLIMode) restCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdWait}, args...)
}

func (c *CLIMode) searchCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdSearch})
}

func (c *CLIMode) openCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdOpen}, args...)
}

func (c *CLIMode) closeCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdClose}, args...)
}

func (c *CLIMode) stairsCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: stairs <up|down>"
	}
	switch strings.ToLower(args[0]) {
	case "up", "u":
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdGoUpstairs})
	case "down", "d":
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdGoDownstairs})
	default:
		return "Usage: stairs <up|down>"
	}
}

func (c *CLIMode) autoCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: auto <explore|pickup>"
	}
	switch strings.ToLower(args[0]) {
	case "explore":
		return "Auto-explore started. (TODO: implement auto-explore)"
	case "pickup":
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdPickUp}, "all")
	default:
		return "Usage: auto <explore|pickup>"
	}
}

func (c *CLIMode) gameCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: game <new|save|load|quit>"
	}
	switch strings.ToLower(args[0]) {
	case "new":
		return "Starting new game. (TODO: implement new game)"
	case "save":
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdSave})
	case "load":
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdLoad})
	case "quit":
		return "Use 'quit' or 'exit' to quit CLI."
	default:
		return "Usage: game <new|save|load|quit>"
	}
}
