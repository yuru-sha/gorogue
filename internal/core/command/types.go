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

	// Action commands
	CmdLook      // Look around (x)
	CmdInventory // Show inventory (i)
	CmdPickUp    // Pick up item (,)
	CmdDrop      // Drop item (d)
	CmdUse       // Use/Apply item (a)
	CmdQuaff     // Quaff potion (q)
	CmdRead      // Read scroll (r)
	CmdWield     // Wield/wear item (w)
	CmdTakeOff   // Take off item (t)
	CmdEat       // Eat food (e)
	CmdWait      // Wait/Rest (.)
	CmdSearch    // Search (s)
	CmdOpen      // Open door (o)
	CmdClose     // Close door (c)
	CmdFight     // Fight/Attack (f)
	CmdDisarm    // Disarm trap (d)
	CmdEquip     // Equip item (e)
	CmdUnequip   // Unequip item (r)

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
		return "Wield/Wear"
	case CmdTakeOff:
		return "Take Off"
	case CmdEat:
		return "Eat"
	case CmdWait:
		return "Wait/Rest"
	case CmdSearch:
		return "Search"
	case CmdOpen:
		return "Open"
	case CmdClose:
		return "Close"
	case CmdFight:
		return "Fight"
	case CmdDisarm:
		return "Disarm"
	case CmdEquip:
		return "Equip"
	case CmdUnequip:
		return "Unequip"
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
