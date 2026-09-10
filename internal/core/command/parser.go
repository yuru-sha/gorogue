package command

import (
	"github.com/anaseto/gruid"
)

// Parser converts key inputs to structured commands
type Parser struct {
	keyMap map[gruid.Key]Command
}

// NewParser creates a new command parser
func NewParser() *Parser {
	p := &Parser{
		keyMap: make(map[gruid.Key]Command),
	}
	p.initializeKeyMap()
	return p
}

// initializeKeyMap sets up the key to command mappings.
func (p *Parser) initializeKeyMap() {
	// Movement commands - vi keys.
	p.keyMap["h"] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap["j"] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap["k"] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap["l"] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}
	p.keyMap["y"] = Command{Type: CmdMoveNorthWest, Direction: Direction{X: -1, Y: -1}}
	p.keyMap["u"] = Command{Type: CmdMoveNorthEast, Direction: Direction{X: 1, Y: -1}}
	p.keyMap["b"] = Command{Type: CmdMoveSouthWest, Direction: Direction{X: -1, Y: 1}}
	p.keyMap["n"] = Command{Type: CmdMoveSouthEast, Direction: Direction{X: 1, Y: 1}}

	// Uppercase movement keys run in the same direction.
	p.keyMap["H"] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap["J"] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap["K"] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap["L"] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}
	p.keyMap["Y"] = Command{Type: CmdMoveNorthWest, Direction: Direction{X: -1, Y: -1}}
	p.keyMap["U"] = Command{Type: CmdMoveNorthEast, Direction: Direction{X: 1, Y: -1}}
	p.keyMap["B"] = Command{Type: CmdMoveSouthWest, Direction: Direction{X: -1, Y: 1}}
	p.keyMap["N"] = Command{Type: CmdMoveSouthEast, Direction: Direction{X: 1, Y: 1}}

	// Movement commands - arrow keys
	p.keyMap[gruid.KeyArrowLeft] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap[gruid.KeyArrowDown] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap[gruid.KeyArrowUp] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap[gruid.KeyArrowRight] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}

	// Movement commands - numpad (for when numlock is off)
	p.keyMap["Left"] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap["Down"] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap["Up"] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap["Right"] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}

	// Action commands. Each key has one meaning; movement keys are not reused.
	p.keyMap["i"] = Command{Type: CmdInventory}
	p.keyMap[","] = Command{Type: CmdPickUp}
	p.keyMap["g"] = Command{Type: CmdPickUp}
	p.keyMap["d"] = Command{Type: CmdDrop}
	p.keyMap["a"] = Command{Type: CmdUse}
	p.keyMap["z"] = Command{Type: CmdUse}
	p.keyMap["q"] = Command{Type: CmdQuaff}
	p.keyMap["r"] = Command{Type: CmdRead}
	p.keyMap["w"] = Command{Type: CmdWield}
	p.keyMap["t"] = Command{Type: CmdTakeOff}
	p.keyMap["e"] = Command{Type: CmdEat}
	p.keyMap["o"] = Command{Type: CmdOpen}
	p.keyMap["c"] = Command{Type: CmdClose}
	p.keyMap["s"] = Command{Type: CmdSearch}
	p.keyMap["D"] = Command{Type: CmdDisarm}
	p.keyMap["f"] = Command{Type: CmdFight}
	p.keyMap["x"] = Command{Type: CmdLook}
	p.keyMap[" "] = Command{Type: CmdWait}
	p.keyMap["."] = Command{Type: CmdWait}
	p.keyMap["^L"] = Command{Type: CmdLook}
	p.keyMap["^R"] = Command{Type: CmdLook}

	// Stair commands.
	p.keyMap["<"] = Command{Type: CmdGoUpstairs}   // Go up
	p.keyMap[">"] = Command{Type: CmdGoDownstairs} // Go down

	// System commands.
	p.keyMap["Q"] = Command{Type: CmdQuit}
	p.keyMap["?"] = Command{Type: CmdHelp}
	p.keyMap["/"] = Command{Type: CmdSymbol}
	p.keyMap[gruid.KeyEscape] = Command{Type: CmdEscape}
	p.keyMap["^W"] = Command{Type: CmdWizard}
	p.keyMap[":"] = Command{Type: CmdCLI}
}

// Parse converts a key input to a command
func (p *Parser) Parse(key gruid.Key) Command {
	if cmd, ok := p.keyMap[key]; ok {
		cmd.Key = string(key)
		return cmd
	}
	return Command{Type: CmdUnknown, Key: string(key)}
}

// GetKeyBindings returns all key bindings for help display.
func (p *Parser) GetKeyBindings() map[string]string {
	bindings := make(map[string]string)

	// Movement
	bindings["h,j,k,l"] = "Move west, south, north, east"
	bindings["y,u,b,n"] = "Move diagonally (NW, NE, SW, SE)"
	bindings["H,J,K,L"] = "Run in direction (until wall/object)"
	bindings["Y,U,B,N"] = "Run diagonally"
	bindings["Arrow keys"] = "Move in four directions"
	bindings["Numpad"] = "Move with keys 1-9 (including diagonals)"

	// Actions
	bindings["i"] = "Inventory - show what you are carrying"
	bindings[","] = "Pick up object(s) (PyRogue style)"
	bindings["g"] = "Get/pick up object(s) (alternative)"
	bindings["a"] = "Apply/use an item"
	bindings["z"] = "Apply/use an item (alternative)"
	bindings["q"] = "Quaff a potion"
	bindings["r"] = "Read a scroll"
	bindings["w"] = "Wield/wear an item"
	bindings["t"] = "Take off an item"
	bindings["e"] = "Eat food"
	bindings["d"] = "Drop an item"
	bindings["o"] = "Open a door"
	bindings["c"] = "Close a door"
	bindings["s"] = "Search for traps/doors"
	bindings["D"] = "Disarm a trap"
	bindings["f"] = "Fight (attack adjacent monster)"
	bindings["x"] = "Look/examine surroundings"
	bindings["."] = "Rest for a turn"
	bindings["Space"] = "Rest for a turn"
	bindings["Ctrl+L"] = "Redraw the screen"
	bindings["Ctrl+R"] = "Repeat last message"

	// Stairs
	bindings["<"] = "Go up a staircase"
	bindings[">"] = "Go down a staircase"

	// System
	bindings["Q"] = "Quit the game"
	bindings["?"] = "Show this help"
	bindings["/"] = "Show symbol explanation"
	bindings["ESC"] = "Cancel command"
	bindings["Ctrl+W"] = "Toggle wizard mode"
	bindings[":"] = "Enter CLI debug mode"

	return bindings
}

// GetCommandForKey returns the command type for a given key
func (p *Parser) GetCommandForKey(key gruid.Key) Type {
	if cmd, ok := p.keyMap[key]; ok {
		return cmd.Type
	}
	return CmdUnknown
}
