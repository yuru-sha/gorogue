package screen

import (
	"testing"
	"time"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/score"
)

func TestVictoryScreen_NewVictoryScreen(t *testing.T) {
	scoreEntry := &score.ScoreEntry{
		PlayerName:     "TestPlayer",
		Score:          5000,
		Level:          10,
		DeepestFloor:   26,
		PlayTime:       7200, // 2 hours
		TurnCount:      1000,
		MonstersKilled: 50,
		GoldCollected:  500,
		IsVictory:      true,
		DeathReason:    "",
		GameSeed:       12345,
		Timestamp:      time.Now(),
		Version:        "v0.1.0",
	}

	screen := NewVictoryScreen(80, 50, scoreEntry)

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

func TestVictoryScreen_HandleInput(t *testing.T) {
	screen := NewVictoryScreen(80, 50, nil)

	// Test Up key
	msg := gruid.MsgKeyDown{Key: "Up"}
	result := screen.HandleInput(msg)
	if result != state.StateVictory {
		t.Errorf("Expected StateVictory, got %v", result)
	}
	if screen.selected != 2 { // Should wrap around
		t.Errorf("Expected selected 2, got %d", screen.selected)
	}

	// Test Down key
	msg = gruid.MsgKeyDown{Key: "Down"}
	result = screen.HandleInput(msg)
	if result != state.StateVictory {
		t.Errorf("Expected StateVictory, got %v", result)
	}
	if screen.selected != 0 { // Should wrap around
		t.Errorf("Expected selected 0, got %d", screen.selected)
	}

	// Test Space key (toggle stats)
	initialShowStats := screen.showStats
	msg = gruid.MsgKeyDown{Key: "Space"}
	result = screen.HandleInput(msg)
	if result != state.StateVictory {
		t.Errorf("Expected StateVictory, got %v", result)
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

func TestVictoryScreen_MenuSelection(t *testing.T) {
	screen := NewVictoryScreen(80, 50, nil)

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

func TestVictoryScreen_BonusCalculations(t *testing.T) {
	scoreEntry := &score.ScoreEntry{
		PlayerName:     "TestPlayer",
		Score:          5000,
		Level:          10,
		DeepestFloor:   26,
		PlayTime:       3600, // 1 hour
		TurnCount:      1000,
		MonstersKilled: 50,
		GoldCollected:  500,
		IsVictory:      true,
		DeathReason:    "",
		GameSeed:       12345,
		Timestamp:      time.Now(),
		Version:        "v0.1.0",
	}

	screen := NewVictoryScreen(80, 50, scoreEntry)

	// Test survival bonus
	survivalBonus := screen.calculateSurvivalBonus()
	expected := 10 * 100 // Level 10 * 100
	if survivalBonus != expected {
		t.Errorf("Expected survival bonus %d, got %d", expected, survivalBonus)
	}

	// Test speed bonus
	speedBonus := screen.calculateSpeedBonus()
	expectedSpeed := 10000 - 1*1000 // 10000 - 1 hour * 1000
	if speedBonus != expectedSpeed {
		t.Errorf("Expected speed bonus %d, got %d", expectedSpeed, speedBonus)
	}

	// Test exploration bonus
	explorationBonus := screen.calculateExplorationBonus()
	expectedExploration := 26 * 50 // Floor 26 * 50
	if explorationBonus != expectedExploration {
		t.Errorf("Expected exploration bonus %d, got %d", expectedExploration, explorationBonus)
	}
}

func TestVictoryScreen_FormatPlayTime(t *testing.T) {
	screen := NewVictoryScreen(80, 50, nil)

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

func TestVictoryScreen_Draw(t *testing.T) {
	scoreEntry := &score.ScoreEntry{
		PlayerName:     "TestPlayer",
		Score:          5000,
		Level:          10,
		DeepestFloor:   26,
		PlayTime:       7200,
		TurnCount:      1000,
		MonstersKilled: 50,
		GoldCollected:  500,
		IsVictory:      true,
		DeathReason:    "",
		GameSeed:       12345,
		Timestamp:      time.Now(),
		Version:        "v0.1.0",
	}

	screen := NewVictoryScreen(80, 50, scoreEntry)
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
