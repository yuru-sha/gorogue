package screen

import (
	"testing"
	"time"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/score"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestGameOverScreen_NewGameOverScreen(t *testing.T) {
	logger.Setup()
	scoreEntry := &score.ScoreEntry{
		PlayerName:     "TestPlayer",
		Score:          1000,
		Level:          5,
		DeepestFloor:   10,
		PlayTime:       3600, // 1 hour
		TurnCount:      500,
		MonstersKilled: 25,
		GoldCollected:  200,
		IsVictory:      false,
		DeathReason:    "Killed by a dragon",
		GameSeed:       12345,
		Timestamp:      time.Now(),
		Version:        "v0.1.0",
	}

	screen := NewGameOverScreen(80, 50, scoreEntry)

	if screen.width != 80 {
		t.Errorf("Expected width 80, got %d", screen.width)
	}
	if screen.height != 50 {
		t.Errorf("Expected height 50, got %d", screen.height)
	}
	if screen.selected != 0 {
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}
	if screen.scoreEntry != scoreEntry {
		t.Error("Expected scoreEntry to be set")
	}
	if len(screen.menuItems) != 3 {
		t.Errorf("Expected 3 menu items, got %d", len(screen.menuItems))
	}
}

func TestGameOverScreen_HandleInput(t *testing.T) {
	logger.Setup()
	screen := NewGameOverScreen(80, 50, nil)

	// Test Up key
	msg := gruid.MsgKeyDown{Key: "Up"}
	result := screen.HandleInput(msg)
	if result != state.StateGameOver {
		t.Errorf("Expected StateGameOver, got %v", result)
	}
	if screen.selected != 2 { // Should wrap around
		t.Errorf("Expected selected 2, got %d", screen.selected)
	}

	// Test Down key
	msg = gruid.MsgKeyDown{Key: "Down"}
	result = screen.HandleInput(msg)
	if result != state.StateGameOver {
		t.Errorf("Expected StateGameOver, got %v", result)
	}
	if screen.selected != 0 { // Should wrap around
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}

	// Test Space key (toggle stats)
	initialShowStats := screen.showStats
	msg = gruid.MsgKeyDown{Key: "Space"}
	result = screen.HandleInput(msg)
	if result != state.StateGameOver {
		t.Errorf("Expected StateGameOver, got %v", result)
	}
	if screen.showStats == initialShowStats {
		t.Error("Expected showStats to toggle")
	}

	// Test Enter key (Restart)
	msg = gruid.MsgKeyDown{Key: "Enter"}
	result = screen.HandleInput(msg)
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}
}

func TestGameOverScreen_MenuSelection(t *testing.T) {
	logger.Setup()
	screen := NewGameOverScreen(80, 50, nil)

	// Test Restart
	screen.selected = 0
	result := screen.handleMenuSelection()
	if result != state.StateGame {
		t.Errorf("Expected StateGame, got %v", result)
	}

	// Test Main Menu
	screen.selected = 1
	result = screen.handleMenuSelection()
	if result != state.StateMenu {
		t.Errorf("Expected StateMenu, got %v", result)
	}

	// Test Quit
	screen.selected = 2
	result = screen.handleMenuSelection()
	if result != state.StateGameOver {
		t.Errorf("Expected StateGameOver, got %v", result)
	}
}

func TestGameOverScreen_FormatPlayTime(t *testing.T) {
	screen := NewGameOverScreen(80, 50, nil)

	// Test minutes only
	result := screen.formatPlayTime(90) // 1 minute 30 seconds
	expected := "01:30"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test hours
	result = screen.formatPlayTime(3661) // 1 hour 1 minute 1 second
	expected = "01:01:01"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test zero time
	result = screen.formatPlayTime(0)
	expected = "00:00"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGameOverScreen_Draw(t *testing.T) {
	scoreEntry := &score.ScoreEntry{
		PlayerName:     "TestPlayer",
		Score:          1000,
		Level:          5,
		DeepestFloor:   10,
		PlayTime:       3600,
		TurnCount:      500,
		MonstersKilled: 25,
		GoldCollected:  200,
		IsVictory:      false,
		DeathReason:    "Killed by a dragon",
		GameSeed:       12345,
		Timestamp:      time.Now(),
		Version:        "v0.1.0",
	}

	screen := NewGameOverScreen(80, 50, scoreEntry)
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
