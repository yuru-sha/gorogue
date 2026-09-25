package command

import (
	"testing"

	"github.com/anaseto/gruid"
)

func TestParser_BasicMovement(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		key       gruid.Key
		expected  Type
		direction Direction
	}{
		// Vi-style movement
		{"h", CmdMoveWest, Direction{X: -1, Y: 0}},
		{"j", CmdMoveSouth, Direction{X: 0, Y: 1}},
		{"k", CmdMoveNorth, Direction{X: 0, Y: -1}},
		{"l", CmdMoveEast, Direction{X: 1, Y: 0}},

		// Diagonal movement
		{"y", CmdMoveNorthWest, Direction{X: -1, Y: -1}},
		{"u", CmdMoveNorthEast, Direction{X: 1, Y: -1}},
		{"b", CmdMoveSouthWest, Direction{X: -1, Y: 1}},
		{"n", CmdMoveSouthEast, Direction{X: 1, Y: 1}},

		// Arrow keys
		{gruid.KeyArrowLeft, CmdMoveWest, Direction{X: -1, Y: 0}},
		{gruid.KeyArrowRight, CmdMoveEast, Direction{X: 1, Y: 0}},
		{gruid.KeyArrowUp, CmdMoveNorth, Direction{X: 0, Y: -1}},
		{gruid.KeyArrowDown, CmdMoveSouth, Direction{X: 0, Y: 1}},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			cmd := parser.Parse(tt.key)
			if cmd.Type != tt.expected {
				t.Errorf("Expected command type %v, got %v", tt.expected, cmd.Type)
			}
			if cmd.Direction != tt.direction {
				t.Errorf("Expected direction %v, got %v", tt.direction, cmd.Direction)
			}
		})
	}
}

func TestParser_OriginalActionBindings(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		key      gruid.Key
		expected Type
	}{
		{"i", CmdInventory},
		{",", CmdPickUp},
		{"d", CmdDrop},
		{"a", CmdRepeat},
		{"q", CmdQuaff},
		{"r", CmdRead},
		{"w", CmdWield},
		{"W", CmdWear},
		{"T", CmdTakeOff},
		{"P", CmdRingOn},
		{"R", CmdRingOff},
		{"e", CmdEat},
		{"s", CmdSearch},
		{"f", CmdFight},
		{"t", CmdThrow},
		{"z", CmdZap},
		{"@", CmdCharacter},
		{"^", CmdFindTrap},
		{" ", CmdUnknown},
		{".", CmdWait},
		{"g", CmdUnknown},
		{"o", CmdUnknown},
		{"c", CmdCall},
		{"x", CmdUnknown},
		{"D", CmdDiscover},
		{"<", CmdGoUpstairs},
		{">", CmdGoDownstairs},
		{"Q", CmdQuit},
		{"?", CmdHelp},
		{"/", CmdSymbol},
		{gruid.KeyEscape, CmdEscape},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			if got := parser.Parse(tt.key).Type; got != tt.expected {
				t.Errorf("Parse(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}
func TestParser_RunBindingsDoNotMoveOneTile(t *testing.T) {
	parser := NewParser()
	for key, direction := range map[gruid.Key]Direction{
		"H": {-1, 0}, "J": {0, 1}, "K": {0, -1}, "L": {1, 0},
		"Y": {-1, -1}, "U": {1, -1}, "B": {-1, 1}, "N": {1, 1},
	} {
		got := parser.Parse(key)
		if got.Type != CmdRun || got.Direction != direction {
			t.Errorf("Parse(%q) = (%v, %+v), want (%v, %+v)", key, got.Type, got.Direction, CmdRun, direction)
		}
	}
}

func TestParser_GetKeyBindingsReportsOriginalActions(t *testing.T) {
	bindings := NewParser().GetKeyBindings()
	for key, want := range map[string]string{
		"a": "Repeat last command",
		"t": "Throw an item",
		"z": "Zap a wand or staff",
		"W": "Wear armor",
		"T": "Take off armor",
		"P": "Put on a ring",
		"R": "Remove a ring",
		"@": "Show character information",
		"c": "Name an unidentified item",
	} {
		if got := bindings[key]; got != want {
			t.Errorf("binding %q = %q, want %q", key, got, want)
		}
	}
}

func TestParser_UnknownCommand(t *testing.T) {
	parser := NewParser()

	cmd := parser.Parse("~") // ~ is not mapped to anything
	if cmd.Type != CmdUnknown {
		t.Errorf("Expected CmdUnknown, got %v", cmd.Type)
	}
}

func TestParser_GetKeyBindings(t *testing.T) {
	parser := NewParser()
	bindings := parser.GetKeyBindings()

	// Check that we have the expected number of bindings
	if len(bindings) == 0 {
		t.Error("Expected key bindings, got empty map")
	}

	// Check for some essential bindings
	essential := []string{
		"h,j,k,l,y,u,b,n",
		"i",
		"Q",
		"?",
		"@",
		"c",
	}

	for _, key := range essential {
		if _, exists := bindings[key]; !exists {
			t.Errorf("Missing essential key binding: %s", key)
		}
	}
}

func TestCommandType_String(t *testing.T) {
	tests := []struct {
		cmdType  Type
		expected string
	}{
		{CmdMoveWest, "Move West"},
		{CmdInventory, "Inventory"},
		{CmdPickUp, "Pick Up"},
		{CmdQuit, "Quit"},
		{CmdHelp, "Help"},
		{CmdUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.cmdType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParser_GetCommandForKey(t *testing.T) {
	parser := NewParser()

	// Test known key
	cmdType := parser.GetCommandForKey("h")
	if cmdType != CmdMoveWest {
		t.Errorf("Expected CmdMoveWest, got %v", cmdType)
	}

	// Test character and unknown keys
	cmdType = parser.GetCommandForKey("@")
	if cmdType != CmdCharacter {
		t.Errorf("Expected CmdCharacter, got %v", cmdType)
	}

	cmdType = parser.GetCommandForKey("~")
	if cmdType != CmdUnknown {
		t.Errorf("Expected CmdUnknown, got %v", cmdType)
	}
}
