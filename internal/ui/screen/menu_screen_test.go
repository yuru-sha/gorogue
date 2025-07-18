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
	if len(screen.menuItems) != 5 {
		t.Errorf("Expected 5 menu items, got %d", len(screen.menuItems))
	}
}

func TestMenuScreen_HandleInput(t *testing.T) {
	logger.Setup()
	screen := NewMenuScreen(80, 50)
	
	// Test Up key
	msg := gruid.MsgKeyDown{Key: "Up"}
	result := screen.HandleInput(msg)
	if result != state.StateMenu {
		t.Errorf("Expected StateMenu, got %v", result)
	}
	if screen.selected != 4 { // Should wrap around
		t.Errorf("Expected selected 4, got %d", screen.selected)
	}
	
	// Test Down key
	msg = gruid.MsgKeyDown{Key: "Down"}
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
	
	// Test New Game
	screen.selected = 0
	result := screen.handleMenuSelection()
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	
	// Test Load Game
	screen.selected = 1
	result = screen.handleMenuSelection()
	if result != state.StateSaveLoad {
		t.Errorf("Expected StateSaveLoad, got %v", result)
	}
	
	// Test Scores
	screen.selected = 2
	result = screen.handleMenuSelection()
	if result != state.StateMenu { // TODO: Should be StateScores when implemented
		t.Errorf("Expected StateMenu, got %v", result)
	}
	
	// Test Help
	screen.selected = 3
	result = screen.handleMenuSelection()
	if result != state.StateHelp {
		t.Errorf("Expected StateHelp, got %v", result)
	}
	
	// Test Quit
	screen.selected = 4
	result = screen.handleMenuSelection()
	if result != state.StateGameOver {
		t.Errorf("Expected StateGameOver, got %v", result)
	}
}