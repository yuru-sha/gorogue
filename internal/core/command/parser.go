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
	// Movement commands use Rogue's vi keys.
	p.keyMap["h"] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap["j"] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap["k"] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap["l"] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}
	p.keyMap["y"] = Command{Type: CmdMoveNorthWest, Direction: Direction{X: -1, Y: -1}}
	p.keyMap["u"] = Command{Type: CmdMoveNorthEast, Direction: Direction{X: 1, Y: -1}}
	p.keyMap["b"] = Command{Type: CmdMoveSouthWest, Direction: Direction{X: -1, Y: 1}}
	p.keyMap["n"] = Command{Type: CmdMoveSouthEast, Direction: Direction{X: 1, Y: 1}}
	p.keyMap[gruid.KeyArrowLeft] = Command{Type: CmdMoveWest, Direction: Direction{X: -1, Y: 0}}
	p.keyMap[gruid.KeyArrowDown] = Command{Type: CmdMoveSouth, Direction: Direction{X: 0, Y: 1}}
	p.keyMap[gruid.KeyArrowUp] = Command{Type: CmdMoveNorth, Direction: Direction{X: 0, Y: -1}}
	p.keyMap[gruid.KeyArrowRight] = Command{Type: CmdMoveEast, Direction: Direction{X: 1, Y: 0}}
	p.keyMap["Left"] = p.keyMap[gruid.KeyArrowLeft]
	p.keyMap["Down"] = p.keyMap[gruid.KeyArrowDown]
	p.keyMap["Up"] = p.keyMap[gruid.KeyArrowUp]
	p.keyMap["Right"] = p.keyMap[gruid.KeyArrowRight]
	p.keyMap["H"] = Command{Type: CmdRun, Direction: Direction{X: -1, Y: 0}}
	p.keyMap["J"] = Command{Type: CmdRun, Direction: Direction{X: 0, Y: 1}}
	p.keyMap["K"] = Command{Type: CmdRun, Direction: Direction{X: 0, Y: -1}}
	p.keyMap["L"] = Command{Type: CmdRun, Direction: Direction{X: 1, Y: 0}}
	p.keyMap["Y"] = Command{Type: CmdRun, Direction: Direction{X: -1, Y: -1}}
	p.keyMap["U"] = Command{Type: CmdRun, Direction: Direction{X: 1, Y: -1}}
	p.keyMap["B"] = Command{Type: CmdRun, Direction: Direction{X: -1, Y: 1}}
	p.keyMap["N"] = Command{Type: CmdRun, Direction: Direction{X: 1, Y: 1}}

	p.keyMap["i"] = Command{Type: CmdInventory}
	p.keyMap[","] = Command{Type: CmdPickUp}
	p.keyMap["d"] = Command{Type: CmdDrop}
	p.keyMap["a"] = Command{Type: CmdRepeat}
	p.keyMap["q"] = Command{Type: CmdQuaff}
	p.keyMap["r"] = Command{Type: CmdRead}
	p.keyMap["w"] = Command{Type: CmdWield}
	p.keyMap["W"] = Command{Type: CmdWear}
	p.keyMap["T"] = Command{Type: CmdTakeOff}
	p.keyMap["P"] = Command{Type: CmdRingOn}
	p.keyMap["R"] = Command{Type: CmdRingOff}
	p.keyMap["e"] = Command{Type: CmdEat}
	p.keyMap["."] = Command{Type: CmdWait}
	p.keyMap["s"] = Command{Type: CmdSearch}
	p.keyMap["^"] = Command{Type: CmdFindTrap}
	p.keyMap["f"] = Command{Type: CmdFight}
	p.keyMap["t"] = Command{Type: CmdThrow}
	p.keyMap["z"] = Command{Type: CmdZap}
	p.keyMap["@"] = Command{Type: CmdCharacter}
	p.keyMap["c"] = Command{Type: CmdCall}
	p.keyMap["D"] = Command{Type: CmdDiscover}

	p.keyMap["<"] = Command{Type: CmdGoUpstairs}
	p.keyMap[">"] = Command{Type: CmdGoDownstairs}
	p.keyMap["Q"] = Command{Type: CmdQuit}
	p.keyMap["?"] = Command{Type: CmdHelp}
	p.keyMap["/"] = Command{Type: CmdSymbol}
	p.keyMap[gruid.KeyEscape] = Command{Type: CmdEscape}
	p.keyMap["^W"] = Command{Type: CmdWizard}
	p.keyMap[":"] = Command{Type: CmdCLI}
}

// Only command keys from Rogue are assigned gameplay meanings here;
// nonconflicting frontend conveniences remain available through named CLI commands.

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

	bindings["h,j,k,l,y,u,b,n"] = "Move one square in eight directions"
	bindings["H,J,K,L,Y,U,B,N"] = "Run in eight directions"
	bindings["Arrow keys"] = "Move in four directions"

	bindings["i"] = "Show inventory"
	bindings[","] = "Pick up an item"
	bindings["d"] = "Drop an item"
	bindings["a"] = "Repeat last command"
	bindings["q"] = "Quaff a potion"
	bindings["r"] = "Read a scroll"
	bindings["w"] = "Wield a weapon"
	bindings["W"] = "Wear armor"
	bindings["T"] = "Take off armor"
	bindings["P"] = "Put on a ring"
	bindings["R"] = "Remove a ring"
	bindings["e"] = "Eat food"
	bindings["."] = "Rest"
	bindings["s"] = "Search nearby for traps"
	bindings["^"] = "Search for a trap in a direction"
	bindings["f"] = "Fight in a direction"
	bindings["t"] = "Throw an item"
	bindings["z"] = "Zap a wand or staff"
	bindings["@"] = "Show character information"
	bindings["D"] = "Show discovered items"
	bindings["c"] = "Name an unidentified item"

	bindings["<"] = "Go up a staircase"
	bindings[">"] = "Go down a staircase"
	bindings["Q"] = "Quit the game"
	bindings["?"] = "Show this help"
	bindings["/"] = "Show symbol explanation"
	bindings["ESC"] = "Cancel command"
	bindings["^W"] = "Toggle wizard mode"
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
