package command

import "strings"

// ParseDirection converts a text direction into a movement vector.
func ParseDirection(input string) (Direction, bool) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "n", "north", "up", "k":
		return Direction{X: 0, Y: -1}, true
	case "s", "south", "down", "j":
		return Direction{X: 0, Y: 1}, true
	case "e", "east", "right", "l":
		return Direction{X: 1, Y: 0}, true
	case "w", "west", "left", "h":
		return Direction{X: -1, Y: 0}, true
	case "ne", "northeast", "u":
		return Direction{X: 1, Y: -1}, true
	case "nw", "northwest", "y":
		return Direction{X: -1, Y: -1}, true
	case "se", "southeast", "m", "numpad3":
		return Direction{X: 1, Y: 1}, true
	case "sw", "southwest", "b":
		return Direction{X: -1, Y: 1}, true
	default:
		return Direction{}, false
	}
}

// NewMoveCommand creates a movement command for a direction.
func NewMoveCommand(direction Direction) Command {
	var commandType Type
	switch direction {
	case Direction{X: -1, Y: 0}:
		commandType = CmdMoveWest
	case Direction{X: 1, Y: 0}:
		commandType = CmdMoveEast
	case Direction{X: 0, Y: -1}:
		commandType = CmdMoveNorth
	case Direction{X: 0, Y: 1}:
		commandType = CmdMoveSouth
	case Direction{X: -1, Y: -1}:
		commandType = CmdMoveNorthWest
	case Direction{X: 1, Y: -1}:
		commandType = CmdMoveNorthEast
	case Direction{X: -1, Y: 1}:
		commandType = CmdMoveSouthWest
	case Direction{X: 1, Y: 1}:
		commandType = CmdMoveSouthEast
	default:
		return Command{Type: CmdUnknown, Direction: direction}
	}
	return Command{Type: commandType, Direction: direction}
}
