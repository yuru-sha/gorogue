package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestGameplayCommandsHaveGUICLIParity(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		gui   func(*GameScreen) gruid.Key
		cli   string
		setup func(*actor.Player, *dungeon.Level)
	}{
		{
			name: "movement",
			gui:  func(*GameScreen) gruid.Key { return "l" },
			cli:  "move east",
		},
		{
			name: "pickup",
			gui:  func(*GameScreen) gruid.Key { return "," },
			cli:  "pickup",
			setup: func(player *actor.Player, level *dungeon.Level) {
				level.Items = []*item.Item{item.NewItem(player.Position.X, player.Position.Y, item.ItemWeapon, "parity sword", 1)}
			},
		},
		{
			name: "combat",
			gui:  func(*GameScreen) gruid.Key { return "l" },
			cli:  "move east",
			setup: func(player *actor.Player, level *dungeon.Level) {
				monster := actor.NewMonster(player.Position.X+1, player.Position.Y, 'B')
				monster.HP = 100
				monster.Type.Speed = 100
				level.Monsters = []*actor.Monster{monster}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			guiLevel := newTestFloor(5, 5)
			if tt.setup != nil {
				tt.setup(guiPlayer, guiLevel)
			}
			gui := NewGameScreen(80, 50, guiPlayer)
			gui.SetLevel(guiLevel)
			gui.HandleInput(gruid.MsgKeyDown{Key: tt.gui(gui)})

			cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			cliLevel := newTestFloor(5, 5)
			if tt.setup != nil {
				tt.setup(cliPlayer, cliLevel)
			}
			cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
			cliMode.IsActive = true
			cliResult := cliMode.ExecuteCommand(tt.cli)

			if guiPlayer.Position.X != cliPlayer.Position.X || guiPlayer.Position.Y != cliPlayer.Position.Y {
				t.Fatalf("position mismatch: GUI=%v CLI=%v", guiPlayer.Position, cliPlayer.Position)
			}
			if guiPlayer.HP != cliPlayer.HP || guiPlayer.Gold != cliPlayer.Gold || guiPlayer.Exp != cliPlayer.Exp {
				t.Fatalf("player state mismatch: GUI=(hp=%d gold=%d exp=%d) CLI=(hp=%d gold=%d exp=%d)",
					guiPlayer.HP, guiPlayer.Gold, guiPlayer.Exp, cliPlayer.HP, cliPlayer.Gold, cliPlayer.Exp)
			}
			if len(guiPlayer.Inventory.Items) != len(cliPlayer.Inventory.Items) {
				t.Fatalf("inventory size mismatch: GUI=%d CLI=%d", len(guiPlayer.Inventory.Items), len(cliPlayer.Inventory.Items))
			}
			if got := gui.messages[len(gui.messages)-1]; got != cliResult {
				t.Fatalf("message mismatch: GUI=%q CLI=%q\nall GUI messages: %s", got, cliResult, strings.Join(gui.messages, " | "))
			}
		})
	}
}

func TestGameplayStairsHaveGUICLIParity(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	guiPlayer := actor.NewPlayerWithSeed(0, 0, 42)
	guiManager := dungeon.NewDungeonManagerWithSeed(guiPlayer, 42)
	guiLevel := guiManager.GetCurrentLevel()
	placePlayerOnStairsDown(t, guiPlayer, guiLevel)
	gui := NewGameScreen(80, 50, guiPlayer)
	gui.SetLevel(guiLevel)
	gui.SetDungeonManager(guiManager)
	gui.HandleInput(gruid.MsgKeyDown{Key: ">"})

	cliPlayer := actor.NewPlayerWithSeed(0, 0, 42)
	cliManager := dungeon.NewDungeonManagerWithSeed(cliPlayer, 42)
	placePlayerOnStairsDown(t, cliPlayer, cliManager.GetCurrentLevel())
	cliMode := cli.NewCLIModeWithDungeonManager(cliManager, cliPlayer)
	cliMode.IsActive = true
	cliResult := cliMode.ExecuteCommand("stairs down")

	if guiManager.GetCurrentFloor() != cliManager.GetCurrentFloor() {
		t.Fatalf("floor mismatch: GUI=%d CLI=%d", guiManager.GetCurrentFloor(), cliManager.GetCurrentFloor())
	}
	if guiPlayer.Position.X != cliPlayer.Position.X || guiPlayer.Position.Y != cliPlayer.Position.Y {
		t.Fatalf("position mismatch: GUI=%v CLI=%v", guiPlayer.Position, cliPlayer.Position)
	}
	if got := gui.messages[len(gui.messages)-1]; got != cliResult {
		t.Fatalf("message mismatch: GUI=%q CLI=%q", got, cliResult)
	}
}

func TestSaveLoadCommandsHaveGUICLIParity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	player := actor.NewPlayerWithSeed(1, 1, 42)
	manager := dungeon.NewDungeonManagerWithSeed(player, 42)
	integration := save.NewSaveGameIntegration()
	if err := integration.Initialize(); err != nil {
		t.Fatal(err)
	}
	integration.SetGameState(player, manager)
	savedX, savedY := player.Position.X, player.Position.Y

	cliMode := cli.NewCLIModeWithDungeonManager(manager, player)
	cliMode.SetSaveIntegration(integration)
	cliMode.IsActive = true
	cliSaveMessage := cliMode.ExecuteCommand("save")
	if cliSaveMessage != "Game saved." {
		t.Fatalf("CLI save result = %q", cliSaveMessage)
	}

	guiSave := NewSaveLoadScreenWithIntegration(80, 50, integration)
	guiSave.SetSelectedOption(0)
	guiSave.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if guiSave.GetMessage() != cliSaveMessage {
		t.Fatalf("save message mismatch: GUI=%q CLI=%q", guiSave.GetMessage(), cliSaveMessage)
	}

	player.Position.X = 9
	if got := cliMode.ExecuteCommand("load"); got != "Game loaded." {
		t.Fatalf("CLI load result = %q", got)
	}
	if cliMode.Player.Position.X != savedX || cliMode.Player.Position.Y != savedY {
		t.Fatalf("CLI load position = (%d, %d)", cliMode.Player.Position.X, cliMode.Player.Position.Y)
	}

	var loadedPlayer *actor.Player
	guiLoad := NewSaveLoadScreenWithIntegration(80, 50, integration)
	guiLoad.SetOnLoad(func(player *actor.Player, _ *dungeon.DungeonManager) {
		loadedPlayer = player
	})
	guiLoad.SetSelectedOption(1)
	guiLoad.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if guiLoad.GetMessage() != "Game loaded." {
		t.Fatalf("GUI load result = %q", guiLoad.GetMessage())
	}
	if loadedPlayer == nil || loadedPlayer.Position.X != savedX || loadedPlayer.Position.Y != savedY {
		t.Fatalf("GUI load position = %+v", loadedPlayer)
	}
}

func placePlayerOnStairsDown(t *testing.T, player *actor.Player, level *dungeon.Level) {
	t.Helper()
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile != nil && tile.Type == dungeon.TileStairsDown {
				player.Position.X = x
				player.Position.Y = y
				return
			}
		}
	}
	t.Fatal("generated level has no down stairs")
}
