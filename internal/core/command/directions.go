package command

import "strings"

type directionMetadata struct {
	commandType Type
	name        string
}

var directions = map[Direction]directionMetadata{
	{X: -1, Y: 0}:  {commandType: CmdMoveWest, name: "west"},
	{X: 1, Y: 0}:   {commandType: CmdMoveEast, name: "east"},
	{X: 0, Y: -1}:  {commandType: CmdMoveNorth, name: "north"},
	{X: 0, Y: 1}:   {commandType: CmdMoveSouth, name: "south"},
	{X: -1, Y: -1}: {commandType: CmdMoveNorthWest, name: "northwest"},
	{X: 1, Y: -1}:  {commandType: CmdMoveNorthEast, name: "northeast"},
	{X: -1, Y: 1}:  {commandType: CmdMoveSouthWest, name: "southwest"},
	{X: 1, Y: 1}:   {commandType: CmdMoveSouthEast, name: "southeast"},
}

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
	metadata, ok := directions[direction]
	if !ok {
		return Command{Type: CmdUnknown, Direction: direction}
	}
	return Command{Type: metadata.commandType, Direction: direction}
}

func directionName(direction Direction) string {
	if metadata, ok := directions[direction]; ok {
		return metadata.name
	}
	return "unknown direction"
}
