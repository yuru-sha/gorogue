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
		{Name: "equip", Description: "Equip an item", Usage: "equip <item_letter>", Handler: c.equipCommand},
		{Name: "wield", Description: "Wield a weapon", Usage: "wield <item_letter>", Handler: c.wieldCommand},
		{Name: "wear", Description: "Wear armor", Usage: "wear <item_letter>", Handler: c.wearCommand},
		{Name: "unequip", Description: "Unequip equipment", Usage: "unequip <slot>", Handler: c.unequipCommand},
		{Name: "takeoff", Description: "Take off armor", Usage: "takeoff", Handler: c.takeOffCommand},
		{Name: "ring-on", Description: "Put on a ring", Usage: "ring-on <item_letter>", Handler: c.ringOnCommand},
		{Name: "ring-off", Description: "Remove a ring", Usage: "ring-off <left|right>", Handler: c.ringOffCommand},
		{Name: "use", Description: "Use an item", Usage: "use <item_letter>", Handler: c.useCommand},
		{Name: "quaff", Description: "Quaff a potion", Usage: "quaff <item_letter>", Handler: c.quaffCommand},
		{Name: "read", Description: "Read a scroll", Usage: "read <item_letter>", Handler: c.readCommand},
		{Name: "eat", Description: "Eat food", Usage: "eat <item_letter>", Handler: c.eatCommand},
		{Name: "throw", Description: "Throw an item", Usage: "throw <direction> <item_letter>", Handler: c.throwCommand},
		{Name: "zap", Description: "Zap a wand or staff", Usage: "zap <direction> <item_letter>", Handler: c.zapCommand},
		{Name: "attack", Description: "Attack in a direction or at coordinates", Usage: "attack <direction|x y>", Handler: c.attackCommand},
		{Name: "look", Description: "Look at position", Usage: "look [x] [y]", Handler: c.lookCommand},
		{Name: "call", Description: "Name an unidentified item", Usage: "call <item_letter> <name>", Handler: c.callCommand},
		{Name: "character", Description: "Show character information", Usage: "character", Handler: c.characterCommand},
		{Name: "rest", Description: "Rest for specified turns", Usage: "rest [turns]", Handler: c.restCommand},
		{Name: "search", Description: "Search nearby for hidden traps/passages", Usage: "search", Handler: c.searchCommand},
		{Name: "findtrap", Description: "Check for a trap in a direction", Usage: "findtrap <direction>", Handler: c.findTrapCommand},
		{Name: "open", Description: "Open door at direction", Usage: "open <direction>", Handler: c.openCommand},
		{Name: "close", Description: "Close door at direction", Usage: "close <direction>", Handler: c.closeCommand},
		{Name: "stairs", Description: "Use stairs (up/down)", Usage: "stairs <up|down>", Handler: c.stairsCommand},
		{Name: "repeat", Description: "Repeat the last turn-consuming command", Usage: "repeat", Handler: c.repeatCommand},
		{Name: "game", Description: "Save or load game", Usage: "game <save|load|quit>", Handler: c.gameCommand},
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

func (c *CLIMode) wieldCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdWield}, args...)
}

func (c *CLIMode) wearCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdWear}, args...)
}

func (c *CLIMode) takeOffCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdTakeOff}, "armor")
}

func (c *CLIMode) ringOnCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdRingOn}, args...)
}

func (c *CLIMode) ringOffCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdRingOff}, args...)
}

func (c *CLIMode) quaffCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdQuaff}, args...)
}

func (c *CLIMode) readCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdRead}, args...)
}

func (c *CLIMode) eatCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdEat}, args...)
}

func (c *CLIMode) characterCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdCharacter})
}

func (c *CLIMode) callCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdCall}, args...)
}

func (c *CLIMode) throwCommand(args []string) string {
	return c.executeDirectionalItem("throw", gamecommand.CmdThrow, args)
}

func (c *CLIMode) zapCommand(args []string) string {
	return c.executeDirectionalItem("zap", gamecommand.CmdZap, args)
}

func (c *CLIMode) executeDirectionalItem(name string, action gamecommand.Type, args []string) string {
	if len(args) != 2 {
		return fmt.Sprintf("Usage: %s <direction> <item_letter>", name)
	}
	direction, ok := gamecommand.ParseDirection(args[0])
	if !ok {
		return fmt.Sprintf("Unknown direction: %s", strings.ToLower(args[0]))
	}
	return c.executeGameplay(gamecommand.Command{Type: action, Direction: direction}, args[1])
}

func (c *CLIMode) findTrapCommand(args []string) string {
	if len(args) != 1 {
		return "Usage: findtrap <direction>"
	}
	direction, ok := gamecommand.ParseDirection(args[0])
	if !ok {
		return fmt.Sprintf("Unknown direction: %s", strings.ToLower(args[0]))
	}
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdFindTrap, Direction: direction})
}

func (c *CLIMode) repeatCommand(args []string) string {
	return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdRepeat}, args...)
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

func (c *CLIMode) gameCommand(args []string) string {
	if len(args) == 0 {
		return "Usage: game <save|load|quit>"
	}
	switch strings.ToLower(args[0]) {
	case commandSave:
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdSave})
	case commandLoad:
		return c.executeGameplay(gamecommand.Command{Type: gamecommand.CmdLoad})
	case "quit":
		return "Use 'quit' or 'exit' to quit CLI."
	default:
		return "Usage: game <save|load|quit>"
	}
}
