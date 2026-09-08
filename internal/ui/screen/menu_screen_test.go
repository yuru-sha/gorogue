package screen

import (
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestMenuScreen_NewMenuScreen(t *testing.T) {
	logger.Setup()
	screen := NewMenuScreen(80, 50)

	if screen.width != 80 {
		t.Errorf("Expected width 80, got %d", screen.width)
	}
	if screen.height != 50 {
		t.Errorf("Expected height 50, got %d", screen.height)
	}
	if screen.selected != 0 {
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}
	expectedItems := 3
	if screen.saveManager.FileExists() {
		expectedItems++
	}
	if len(screen.menuItems) != expectedItems {
		t.Errorf("Expected %d menu items, got %d", expectedItems, len(screen.menuItems))
	}
}

func TestMenuScreen_HandleInput(t *testing.T) {
	logger.Setup()
	screen := NewMenuScreen(80, 50)

	// Test Up key
	msg := gruid.MsgKeyDown{Key: gruid.KeyArrowUp}
	result := screen.HandleInput(msg)
	if result != state.StateMenu {
		t.Errorf("Expected StateMenu, got %v", result)
	}
	if screen.selected != len(screen.menuItems)-1 { // Should wrap around
		t.Errorf("Expected selected %d, got %d", len(screen.menuItems)-1, screen.selected)
	}

	// Test Down key
	msg = gruid.MsgKeyDown{Key: gruid.KeyArrowDown}
	result = screen.HandleInput(msg)
	if result != state.StateMenu {
		t.Errorf("Expected StateMenu, got %v", result)
	}
	if screen.selected != 0 { // Should wrap around
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}

	// Test Enter key (New Game)
	msg = gruid.MsgKeyDown{Key: "Enter"}
	result = screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}

	// Test direct key selection
	msg = gruid.MsgKeyDown{Key: "H"}
	result = screen.HandleInput(msg)
	if result != state.StateHelp {
		t.Errorf("Expected StateHelp, got %v", result)
	}

	// Test quit key
	msg = gruid.MsgKeyDown{Key: "q"}
	result = screen.HandleInput(msg)
	if result != state.StateGameOver {
		t.Errorf("Expected StateGameOver, got %v", result)
	}
}

func TestMenuScreen_Draw(t *testing.T) {
	logger.Setup()
	screen := NewMenuScreen(80, 50)
	grid := gruid.NewGrid(80, 50)

	// Test that drawing doesn't panic
	screen.Draw(&grid)

	// Check that the grid is not empty
	hasContent := false
	for y := 0; y < 50; y++ {
		for x := 0; x < 80; x++ {
			pos := gruid.Point{X: x, Y: y}
			if pos.X >= 0 && pos.Y >= 0 && pos.X < grid.Size().X && pos.Y < grid.Size().Y {
				cell := grid.At(pos)
				if cell.Rune != ' ' {
					hasContent = true
					break
				}
			}
		}
		if hasContent {
			break
		}
	}

	if !hasContent {
		t.Error("Expected grid to have content after drawing")
	}
}

func TestMenuScreen_MenuSelection(t *testing.T) {
	logger.Setup()
	screen := NewMenuScreen(80, 50)

	expectedStates := map[string]state.GameState{
		"New Game":  state.StateGame,
		"Load Game": state.StateSaveLoad,
		"Help":      state.StateHelp,
		"Quit":      state.StateGameOver,
	}
	for i, menuItem := range screen.menuItems {
		screen.selected = i
		result := screen.handleMenuSelection()
		if result != expectedStates[menuItem] {
			t.Errorf("%s: expected %v, got %v", menuItem, expectedStates[menuItem], result)
		}
	}
}
