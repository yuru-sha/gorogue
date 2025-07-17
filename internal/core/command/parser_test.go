package command

import (
	"testing"

	"github.com/anaseto/gruid"
)

func TestParser_BasicMovement(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		key      gruid.Key
		expected Type
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

func TestParser_ActionCommands(t *testing.T) {
	parser := NewParser()
	
	tests := []struct {
		key      gruid.Key
		expected Type
	}{
		// Basic actions
		{"i", CmdInventory},
		{"g", CmdPickUp},
		{",", CmdPickUp},
		{"d", CmdDrop},
		{"q", CmdQuaff},
		{"r", CmdRead},
		{"w", CmdWield},
		{"W", CmdWield},
		{"T", CmdTakeOff},
		{"s", CmdSearch},
		
		// Movement-related
		{" ", CmdWait},
		{".", CmdWait},
		
		// Stairs
		{"<", CmdGoUpstairs},
		{">", CmdGoDownstairs},
		
		// System
		{"Q", CmdQuit},
		{"S", CmdQuit},
		{"?", CmdHelp},
		{gruid.KeyEscape, CmdEscape},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			cmd := parser.Parse(tt.key)
			if cmd.Type != tt.expected {
				t.Errorf("Expected command type %v, got %v", tt.expected, cmd.Type)
			}
		})
	}
}

func TestParser_UnknownCommand(t *testing.T) {
	parser := NewParser()
	
	cmd := parser.Parse("@") // @ is not mapped to anything
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
		"h,j,k,l",
		"i",
		"Q",
		"?",
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
	
	// Test unknown key
	cmdType = parser.GetCommandForKey("@")
	if cmdType != CmdUnknown {
		t.Errorf("Expected CmdUnknown, got %v", cmdType)
	}
}