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

// initializeKeyMap sets up the key to command mappings - PyRogue style
func (p *Parser) initializeKeyMap() {
	// Movement commands - vi keys (PyRogue standard)
	p.keyMap["h"] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap["j"] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap["k"] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap["l"] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}
	p.keyMap["y"] = Command{Type: CmdMoveNorthWest, Direction: Direction{X: -1, Y: -1}}
	p.keyMap["u"] = Command{Type: CmdMoveNorthEast, Direction: Direction{X: 1, Y: -1}}
	p.keyMap["b"] = Command{Type: CmdMoveSouthWest, Direction: Direction{X: -1, Y: 1}}
	p.keyMap["n"] = Command{Type: CmdMoveSouthEast, Direction: Direction{X: 1, Y: 1}}

	// Movement commands - uppercase for running (PyRogue style)
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

	// Action commands - PyRogue style
	p.keyMap["i"] = Command{Type: CmdInventory} // Inventory
	p.keyMap[","] = Command{Type: CmdPickUp}    // Pick up (PyRogue style)
	p.keyMap["g"] = Command{Type: CmdPickUp}    // Pick up (also g for compatibility)
	p.keyMap["u"] = Command{Type: CmdUse}       // Use item (PyRogue unified interface)
	p.keyMap["e"] = Command{Type: CmdEquip}     // Equip item (PyRogue style)
	p.keyMap["r"] = Command{Type: CmdUnequip}   // Unequip item (PyRogue style)
	p.keyMap["d"] = Command{Type: CmdDisarm}    // Disarm trap (PyRogue style)
	p.keyMap["o"] = Command{Type: CmdOpen}      // Open door
	p.keyMap["c"] = Command{Type: CmdClose}     // Close door
	p.keyMap["s"] = Command{Type: CmdSearch}    // Search
	p.keyMap["z"] = Command{Type: CmdUse}       // Spellbook (PyRogue style)
	p.keyMap["f"] = Command{Type: CmdFight}     // Fight
	p.keyMap["x"] = Command{Type: CmdLook}      // Look/examine
	p.keyMap[" "] = Command{Type: CmdWait}      // Space bar to rest/wait
	p.keyMap["."] = Command{Type: CmdWait}      // Period to rest (when not on stairs)
	p.keyMap[gruid.KeyTab] = Command{Type: CmdToggleFOV} // Toggle FOV (PyRogue style)
	p.keyMap["^L"] = Command{Type: CmdLook}     // Ctrl+L to redraw screen
	p.keyMap["^R"] = Command{Type: CmdLook}     // Ctrl+R to repeat last message

	// Stair commands - PyRogue style
	p.keyMap["<"] = Command{Type: CmdGoUpstairs}   // Go up
	p.keyMap[">"] = Command{Type: CmdGoDownstairs} // Go down

	// System commands - PyRogue style
	p.keyMap["Q"] = Command{Type: CmdQuit}               // Quit
	p.keyMap["S"] = Command{Type: CmdQuit}               // Save and quit (PyRogue)
	p.keyMap["?"] = Command{Type: CmdHelp}               // Help
	p.keyMap["/"] = Command{Type: CmdLook}               // Identify object (PyRogue)
	p.keyMap[gruid.KeyEscape] = Command{Type: CmdEscape} // Escape/cancel
	p.keyMap["^W"] = Command{Type: CmdWizard}            // Ctrl+W for wizard mode
	p.keyMap[":"] = Command{Type: CmdCLI}                // CLI mode (our addition)
}

// Parse converts a key input to a command
func (p *Parser) Parse(key gruid.Key) Command {
	if cmd, ok := p.keyMap[key]; ok {
		cmd.Key = string(key)
		return cmd
	}
	return Command{Type: CmdUnknown, Key: string(key)}
}

// GetKeyBindings returns all key bindings for help display - PyRogue style
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
	bindings["u"] = "Use item (unified interface)"
	bindings["e"] = "Equip item"
	bindings["r"] = "Unequip item"
	bindings["d"] = "Disarm trap"
	bindings["o"] = "Open a door"
	bindings["c"] = "Close a door"
	bindings["s"] = "Search for traps/doors"
	bindings["z"] = "Spellbook"
	bindings["f"] = "Fight (attack adjacent monster)"
	bindings["x"] = "Look/examine surroundings"
	bindings["."] = "Rest for a turn"
	bindings["Space"] = "Rest for a turn"
	bindings["Tab"] = "Toggle field of view display"
	bindings["Ctrl+L"] = "Redraw the screen"
	bindings["Ctrl+R"] = "Repeat last message"

	// Stairs
	bindings["<"] = "Go up a staircase"
	bindings[">"] = "Go down a staircase"

	// System
	bindings["Q"] = "Quit the game"
	bindings["S"] = "Save and quit"
	bindings["?"] = "Show this help"
	bindings["ESC"] = "Cancel command"
	bindings["Ctrl+W"] = "Toggle wizard mode"
	bindings["Ctrl+S"] = "Save game"
	bindings["Ctrl+L"] = "Load game"
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
