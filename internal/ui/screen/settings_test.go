package screen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestOptionsScreenTogglesAndEditsValues(t *testing.T) {
	screen := NewGameScreen(80, 50, actor.NewPlayerWithSeed(1, 1, 1))
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "O"}); got != state.StateSettings {
		t.Fatalf("opening options state = %v, want StateSettings", got)
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeySpace}); got != state.StateSettings || !screen.options.Terse {
		t.Fatalf("terse toggle state/value = %v/%t, want StateSettings/true", got, screen.options.Terse)
	}
	for range 6 {
		screen.HandleInput(gruid.MsgKeyDown{Key: "j"})
	}
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if screen.options.InventoryStyle != config.InventorySlow {
		t.Fatalf("inventory style = %q, want %q", screen.options.InventoryStyle, config.InventorySlow)
	}
	screen.HandleInput(gruid.MsgKeyDown{Key: "j"})
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	for _, key := range []gruid.Key{"A", "d", "a"} {
		screen.HandleInput(gruid.MsgKeyDown{Key: key})
	}
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if screen.options.Name != "Ada" || !screen.options.NameConfigured {
		t.Fatalf("edited player name = %+v, want configured name Ada", screen.options)
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEscape}); got != state.StateGame {
		t.Fatalf("closing options state = %v, want StateGame", got)
	}
	if screen.options.InventoryStyle != config.InventorySlow {
		t.Fatalf("inventory style = %q, want %q", screen.options.InventoryStyle, config.InventorySlow)
	}
}

func TestOptionsScreenRetainsPlayerNameWhenEditedValueIsEmpty(t *testing.T) {
	screen := NewGameScreen(80, 50, actor.NewPlayerWithSeed(1, 1, 1))
	screen.HandleInput(gruid.MsgKeyDown{Key: "O"})
	for range 7 {
		screen.HandleInput(gruid.MsgKeyDown{Key: "j"})
	}
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyBackspace})
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if screen.options.Name != "Player" || screen.options.NameConfigured {
		t.Fatalf("empty player name edit = %+v, want original unconfigured name", screen.options)
	}
}

func TestOptionsEditorPreservesLoadedPlayerIdentity(t *testing.T) {
	screen := NewGameScreen(80, 50, actor.NewPlayerWithSeed(1, 1, 1))
	integration := save.NewSaveGameIntegration()
	integration.SetGameInfo(save.GameInfo{CharName: "Ada"})
	screen.SetSaveIntegration(integration)
	screen.ApplyLoadedState(screen.player, nil)
	if screen.options.Name != "Ada" || screen.options.NameConfigured {
		t.Fatalf("loaded identity options = %+v, want saved Ada without explicit option", screen.options)
	}

	screen.HandleInput(gruid.MsgKeyDown{Key: "O"})
	for range 7 {
		screen.HandleInput(gruid.MsgKeyDown{Key: "j"})
	}
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if screen.options.Name != "Ada" || screen.options.NameConfigured {
		t.Fatalf("unchanged name edit = %+v, want saved Ada without explicit option", screen.options)
	}
}

func TestEmbeddedCLIOutputAppliesFruitSubstitutionOnce(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 1)
	player.Inventory.AddItem(item.NewItem(0, 0, item.ItemFood, "slime-mold", 0))
	manager := dungeon.NewDungeonManagerWithSeed(player, 1)
	screen := NewGameScreen(80, 50, player)
	screen.SetDungeonManager(manager)
	options := config.DefaultOptions()
	options.Fruit = "slime-mold pie"
	screen.SetOptions(&options)
	screen.cliMode.IsActive = true
	screen.cliBuffer = "inventory"

	want := screen.cliMode.ExecuteCommand(screen.cliBuffer)
	screen.messages = nil
	screen.handleCLIInput(gruid.KeyEnter)
	if got, expected := strings.Join(screen.messages, "\n"), "> inventory\n"+want; got != expected {
		t.Fatalf("embedded CLI output = %q, want single-substitution result %q", got, expected)
	}
}

func TestEmbeddedCLILoadUpdatesUnconfiguredPlayerIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LOG_DIRECTORY", t.TempDir())
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(logger.Cleanup)
	path := filepath.Join(t.TempDir(), "saved.sav")
	options := config.DefaultOptions()
	options.File = path
	integration := save.NewSaveGameIntegration()
	integration.SetOptions(&options)
	if err := integration.Initialize(); err != nil {
		t.Fatal(err)
	}
	savedPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	savedDungeon := dungeon.NewDungeonManagerWithSeed(savedPlayer, 42)
	integration.SetGameState(savedPlayer, savedDungeon)
	integration.SetGameInfo(save.GameInfo{CharName: "Ada"})
	if err := integration.SaveGame(); err != nil {
		t.Fatal(err)
	}

	screen := NewGameScreen(80, 50, actor.NewPlayerWithSeed(1, 1, 1))
	screen.SetOptions(&options)
	screen.SetDungeonManager(dungeon.NewDungeonManagerWithSeed(screen.player, 1))
	screen.SetSaveIntegration(integration)
	screen.cliMode.IsActive = true
	screen.cliBuffer = "load"
	screen.handleCLIInput(gruid.KeyEnter)
	if screen.options.Name != "Ada" || screen.options.NameConfigured {
		t.Fatalf("embedded CLI loaded identity = %+v, want saved Ada without explicit option", screen.options)
	}
}
