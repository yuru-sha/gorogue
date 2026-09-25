// Package command provides a structured command system for the game
package command

// Type represents the type of command
type Type int

const (
	// Movement commands
	CmdMoveWest Type = iota
	CmdMoveEast
	CmdMoveNorth
	CmdMoveSouth
	CmdMoveNorthWest
	CmdMoveNorthEast
	CmdMoveSouthWest
	CmdMoveSouthEast
	CmdRun

	// Action commands
	CmdLook      // Look around (CLI convenience)
	CmdInventory // Show inventory (i)
	CmdPickUp    // Pick up item (,)
	CmdDrop      // Drop item (d)
	CmdUse       // Use/Apply item (CLI convenience)
	CmdQuaff     // Quaff potion (q)
	CmdRead      // Read scroll (r)
	CmdWield     // Wield a weapon (w)
	CmdWear      // Wear armor (W)
	CmdTakeOff   // Take off armor/ring (T/R)
	CmdRingOn    // Put on a ring (P)
	CmdRingOff   // Remove a ring (R)
	CmdThrow     // Throw an item (t)
	CmdZap       // Zap a wand or staff (z)
	CmdEat       // Eat food (e)
	CmdWait      // Rest (.)
	CmdSearch    // Search nearby (s)
	CmdFindTrap  // Search for a trap in a direction (^)
	CmdOpen      // Open door (CLI convenience)
	CmdClose     // Close door (CLI convenience)
	CmdFight     // Fight/Attack (f)
	CmdEquip     // Equip item (CLI convenience)
	CmdUnequip   // Unequip item (CLI convenience)
	CmdRepeat    // Repeat the last command (a)
	CmdCharacter // Show character information (@)
	CmdCall      // Assign a call name to an unidentified item (c)
	CmdDiscover  // List discovered source item names (D)

	// Stair commands
	CmdGoUpstairs   // Go up stairs (<)
	CmdGoDownstairs // Go down stairs (>)

	// System commands
	CmdQuit    // Quit game (Q)
	CmdHelp    // Show help (?)
	CmdEscape  // Cancel/Back (ESC)
	CmdWizard  // Toggle wizard mode (^W)
	CmdCLI     // Enter CLI mode (:)
	CmdSymbol  // Show symbol explanation (/)
	CmdSave    // Save game
	CmdLoad    // Load game
	CmdUnknown // Unknown command
)

// Command represents a game command
type Command struct {
	Type      Type
	Key       string
	Direction Direction // For movement commands
}

// Direction represents movement direction
type Direction struct {
	X, Y int
}

// String returns the string representation of a command type
//
//nolint:gocyclo // The command enum is intentionally rendered in one exhaustive switch.
func (t Type) String() string {
	switch t {
	case CmdMoveWest:
		return "Move West"
	case CmdMoveEast:
		return "Move East"
	case CmdMoveNorth:
		return "Move North"
	case CmdMoveSouth:
		return "Move South"
	case CmdMoveNorthWest:
		return "Move North-West"
	case CmdMoveNorthEast:
		return "Move North-East"
	case CmdMoveSouthWest:
		return "Move South-West"
	case CmdMoveSouthEast:
		return "Move South-East"
	case CmdRun:
		return "Run"
	case CmdLook:
		return "Look"
	case CmdInventory:
		return "Inventory"
	case CmdPickUp:
		return "Pick Up"
	case CmdDrop:
		return "Drop"
	case CmdUse:
		return "Use/Apply"
	case CmdQuaff:
		return "Quaff"
	case CmdRead:
		return "Read"
	case CmdWield:
		return "Wield"
	case CmdWear:
		return "Wear"
	case CmdTakeOff:
		return "Take Off"
	case CmdRingOn:
		return "Put On Ring"
	case CmdRingOff:
		return "Remove Ring"
	case CmdThrow:
		return "Throw"
	case CmdZap:
		return "Zap"
	case CmdEat:
		return "Eat"
	case CmdWait:
		return "Wait/Rest"
	case CmdSearch:
		return "Search"
	case CmdFindTrap:
		return "Find Trap"
	case CmdOpen:
		return "Open"
	case CmdClose:
		return "Close"
	case CmdFight:
		return "Fight"
	case CmdEquip:
		return "Equip"
	case CmdUnequip:
		return "Unequip"
	case CmdRepeat:
		return "Repeat"
	case CmdCharacter:
		return "Character"
	case CmdDiscover:
		return "Discover"
	case CmdCall:
		return "Call"
	case CmdGoUpstairs:
		return "Go Upstairs"
	case CmdGoDownstairs:
		return "Go Downstairs"
	case CmdQuit:
		return "Quit"
	case CmdHelp:
		return "Help"
	case CmdEscape:
		return "Cancel"
	case CmdWizard:
		return "Wizard Mode"
	case CmdCLI:
		return "CLI Mode"
	case CmdSymbol:
		return "Symbol Explanation"
	case CmdSave:
		return "Save"
	case CmdLoad:
		return "Load"
	default:
		return "Unknown"
	}
}
