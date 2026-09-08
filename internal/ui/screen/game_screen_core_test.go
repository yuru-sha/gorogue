package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
)

func TestSetDungeonManagerBindsCLIToSharedState(t *testing.T) {
	player := actor.NewPlayerWithSeed(0, 0, 12345)
	manager := dungeon.NewDungeonManagerWithSeed(player, 12345)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(manager.GetCurrentLevel())
	screen.SetDungeonManager(manager)
	screen.cliMode.IsActive = true
	screen.inputMode = ModeCLI

	for _, key := range "level 2" {
		screen.handleCLIInput(gruid.Key(string(key)))
	}
	screen.handleCLIInput(gruid.KeyEnter)
	if manager.GetCurrentFloor() != 2 || screen.level != manager.GetCurrentLevel() {
		t.Fatalf("GUI CLI did not switch shared dungeon state: floor=%d", manager.GetCurrentFloor())
	}
	if !strings.Contains(strings.Join(screen.messages, "\n"), "Moved to floor 2") {
		t.Fatalf("unexpected CLI result: %v", screen.messages)
	}
}
