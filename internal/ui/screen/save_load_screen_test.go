package screen

import (
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestSaveLoadScreen_NewSaveLoadScreen(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	if screen.width != 80 {
		t.Errorf("Expected width 80, got %d", screen.width)
	}
	if screen.height != 50 {
		t.Errorf("Expected height 50, got %d", screen.height)
	}
	if screen.selected != 0 {
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}
	if screen.saveManager != saveManager {
		t.Error("Expected saveManager to be set")
	}
	if len(screen.menuOptions) != 3 {
		t.Errorf("Expected 3 menu options, got %d", len(screen.menuOptions))
	}
}

func TestSaveLoadScreen_HandleInput(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	// Test Up key
	msg := gruid.MsgKeyDown{Key: "Up"}
	result := screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	if screen.selected != 2 { // Should wrap around
		t.Errorf("Expected selected 2, got %d", screen.selected)
	}
	
	// Test Down key
	msg = gruid.MsgKeyDown{Key: "Down"}
	result = screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	if screen.selected != 0 { // Should wrap around
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}
	
	// Test Enter key (Save Game)
	msg = gruid.MsgKeyDown{Key: "Enter"}
	result = screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	
	// Test direct save key
	msg = gruid.MsgKeyDown{Key: "s"}
	result = screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	
	// Test direct load key
	msg = gruid.MsgKeyDown{Key: "l"}
	result = screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
}

func TestSaveLoadScreen_MenuSelection(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	// Test Save Game
	screen.selected = 0
	result := screen.handleSelection()
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	
	// Test Load Game
	screen.selected = 1
	result = screen.handleSelection()
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
	
	// Test Back
	screen.selected = 2
	result = screen.handleSelection()
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
}

func TestSaveLoadScreen_Draw(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
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

func TestSaveLoadScreen_Validation(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	// Test valid state
	err := screen.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Test invalid save manager
	screen.saveManager = nil
	err = screen.Validate()
	if err == nil {
		t.Error("Expected error for nil save manager")
	}
	
	// Test invalid selected option
	screen.saveManager = saveManager
	screen.selected = -1
	err = screen.Validate()
	if err == nil {
		t.Error("Expected error for invalid selected option")
	}
}

func TestSaveLoadScreen_ActionDescription(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	// Test Save Game
	screen.selected = 0
	description := screen.GetActionDescription()
	if description != "Save game" {
		t.Errorf("Expected 'Save game', got '%s'", description)
	}
	
	// Test Load Game
	screen.selected = 1
	description = screen.GetActionDescription()
	if description != "Cannot load - no save file" {
		t.Errorf("Expected 'Cannot load - no save file', got '%s'", description)
	}
	
	// Test Back
	screen.selected = 2
	description = screen.GetActionDescription()
	if description != "Back to game" {
		t.Errorf("Expected 'Back to game', got '%s'", description)
	}
}

func TestSaveLoadScreen_CanPerformAction(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	// Test Save Game (always possible)
	screen.selected = 0
	if !screen.CanPerformAction() {
		t.Error("Expected save action to be possible")
	}
	
	// Test Load Game (not possible without save file)
	screen.selected = 1
	if screen.CanPerformAction() {
		t.Error("Expected load action to be impossible without save file")
	}
	
	// Test Back (always possible)
	screen.selected = 2
	if !screen.CanPerformAction() {
		t.Error("Expected back action to be possible")
	}
}

func TestSaveLoadScreen_GetStatus(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	status := screen.GetStatus()
	
	if status["mode"] != screen.mode {
		t.Error("Expected mode to be set in status")
	}
	if status["selected_option"] != screen.selected {
		t.Error("Expected selected_option to be set in status")
	}
	if status["has_save_file"] != false {
		t.Error("Expected has_save_file to be false")
	}
	if len(status["menu_options"].([]string)) != 3 {
		t.Error("Expected 3 menu options in status")
	}
}

func TestSaveLoadScreen_GetAvailableActions(t *testing.T) {
	logger.Setup()
	saveManager := save.NewSaveManager()
	screen := NewSaveLoadScreen(80, 50, saveManager)
	
	actions := screen.GetAvailableActions()
	
	expectedActions := []string{"Navigate", "Select", "Back", "Save", "Load", "Help"}
	if len(actions) != len(expectedActions) {
		t.Errorf("Expected %d actions, got %d", len(expectedActions), len(actions))
	}
	
	for i, expected := range expectedActions {
		if i < len(actions) && actions[i] != expected {
			t.Errorf("Expected action %d to be '%s', got '%s'", i, expected, actions[i])
		}
	}
}