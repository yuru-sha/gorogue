package core

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestNewEngineStartsInGame(t *testing.T) {
	logger.Setup()
	engine := NewEngineWithSeed(12345)
	if got := engine.stateManager.GetCurrentState(); got != state.StateGame {
		t.Fatalf("initial state = %v, want StateGame", got)
	}
	if !engine.runActive {
		t.Fatal("new game is not marked active for SIGHUP recovery")
	}
}

func TestEngineHandlesHangupOnUpdateThreadWithoutAdvancingTurn(t *testing.T) {
	logger.Setup()
	t.Setenv("HOME", t.TempDir())
	engine := NewEngineWithSeed(12345)
	beforeTurn := engine.saveIntegration.GetGameStats().GetTurnCount()

	if effect := engine.Update(hangupMessage{}); effect == nil {
		t.Fatal("SIGHUP update did not request application shutdown")
	}
	if !engine.saveIntegration.GetSaveManager().FileExists() {
		t.Fatal("SIGHUP update did not persist recovery save")
	}
	if got := engine.saveIntegration.GetGameStats().GetTurnCount(); got != beforeTurn {
		t.Fatalf("SIGHUP advanced turn count from %d to %d", beforeTurn, got)
	}
}

func TestEngineHangupAfterRunEndedPreservesExistingSave(t *testing.T) {
	logger.Setup()
	t.Setenv("HOME", t.TempDir())
	engine := NewEngineWithSeed(12345)
	if err := engine.saveIntegration.SaveGame(); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}
	if effect := engine.Update(gruid.MsgKeyDown{Key: "Q"}); effect == nil {
		t.Fatal("quit did not end the active run")
	}
	saveManager := engine.saveIntegration.GetSaveManager()
	before, err := saveManager.LoadGame()
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	engine.player.Position.X += 1

	if effect := engine.Update(hangupMessage{}); effect == nil {
		t.Fatal("SIGHUP update did not request application shutdown")
	}
	after, err := saveManager.LoadGame()
	if err != nil {
		t.Fatalf("LoadGame() after SIGHUP error = %v", err)
	}
	if after.PlayerData.X != before.PlayerData.X {
		t.Fatalf("SIGHUP overwrote saved player position: got %d, want %d", after.PlayerData.X, before.PlayerData.X)
	}
}

func TestEngineHangupSavesAuthoritativeRestoredState(t *testing.T) {
	logger.Setup()
	t.Setenv("HOME", t.TempDir())
	engine := NewEngineWithSeed(12345)
	engine.stateManager.SetState(state.StateHelp)
	engine.runActive = true
	player := actor.NewPlayerWithSeed(7, 8, 99)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, 99)
	player.Position.X, player.Position.Y = 7, 8
	engine.saveIntegration.SetGameState(player, dungeonManager)

	if effect := engine.Update(hangupMessage{}); effect == nil {
		t.Fatal("SIGHUP update did not request application shutdown")
	}
	saved, err := engine.saveIntegration.GetSaveManager().LoadGame()
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	if saved.PlayerData.X != 7 || saved.PlayerData.Y != 8 {
		t.Fatalf("SIGHUP saved stale player position (%d, %d)", saved.PlayerData.X, saved.PlayerData.Y)
	}
}

func TestEngineRendersGameOverAfterFatalInventoryAction(t *testing.T) {
	logger.Setup()
	engine := NewEngineWithSeed(12345)
	engine.stateManager.SetState(state.StateGame)
	engine.player.Position.X = 1
	engine.player.Position.Y = 1
	engine.player.HP = 1
	engine.player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemFood, "ration", 1))

	level := fatalEngineLevel()
	engine.gameScreen.SetLevel(level)

	engine.Update(gruid.MsgKeyDown{Key: "d"})
	if !strings.Contains(engine.Draw().String(), "Drop which item?") {
		t.Fatal("drop selection prompt was not rendered")
	}
	if effect := engine.Update(gruid.MsgKeyDown{Key: "a"}); effect != nil {
		t.Fatal("fatal inventory action returned an end effect before rendering game over")
	}
	if got := engine.stateManager.GetCurrentState(); got != state.StateGameOver {
		t.Fatalf("current state = %v, HP=%d, inventory=%d, dropped=%d; want StateGameOver",
			got, engine.player.HP, engine.player.Inventory.Size(), len(level.Items))
	}

	grid := engine.Draw()
	var rendered strings.Builder
	for y := 0; y < 50; y++ {
		for x := 0; x < 80; x++ {
			rendered.WriteRune(grid.At(gruid.Point{X: x, Y: y}).Rune)
		}
	}
	if !strings.Contains(rendered.String(), "You died.") {
		t.Fatal("game-over sequence was not rendered after a fatal inventory action")
	}
	engine.Update(gruid.MsgKeyDown{Key: gruid.KeySpace})
	if got := engine.stateManager.GetCurrentState(); got != state.StateGameOver {
		t.Fatalf("state after death acknowledgement = %v, want StateGameOver", got)
	}
	if !strings.Contains(engine.Draw().String(), "Score:") {
		t.Fatal("death score page was not rendered")
	}
	if effect := engine.Update(gruid.MsgKeyDown{Key: gruid.KeySpace}); effect == nil {
		t.Fatal("final death acknowledgement did not end the run")
	}
	if got := engine.stateManager.GetCurrentState(); got != state.StateQuit {
		t.Fatalf("state after death sequence = %v, want StateQuit", got)
	}
}
func TestEngineShowsVictorySequenceAfterSurfaceExit(t *testing.T) {
	logger.Setup()
	engine := NewEngineWithSeed(12345)
	upStairs, _ := dungeon.NewStairsManager(engine.dungeonManager.GetCurrentLevel()).GetStairPositions()
	if len(upStairs) != 1 {
		t.Fatalf("upstairs count = %d, want 1", len(upStairs))
	}
	engine.player.Position.X, engine.player.Position.Y = upStairs[0].X, upStairs[0].Y
	engine.player.Inventory.AddItem(gameitem.NewAmulet(0, 0))

	engine.Update(gruid.MsgKeyDown{Key: "<"})
	if got := engine.stateManager.GetCurrentState(); got != state.StateVictory {
		t.Fatalf("state after surface exit = %v, want StateVictory", got)
	}
	if !strings.Contains(engine.Draw().String(), "Congratulations") {
		t.Fatal("victory message was not rendered")
	}
	engine.Update(gruid.MsgKeyDown{Key: gruid.KeySpace})
	if !strings.Contains(engine.Draw().String(), "Score:") {
		t.Fatal("victory score stage was not rendered")
	}
	if effect := engine.Update(gruid.MsgKeyDown{Key: gruid.KeySpace}); effect == nil {
		t.Fatal("final victory acknowledgement did not end the run")
	}
}

func TestEngineSyncLoadedStateAfterSharedExecutor(t *testing.T) {
	logger.Setup()
	t.Setenv("HOME", t.TempDir())
	engine := NewEngineWithSeed(12345)
	engine.player.Position.X = 3
	engine.player.Position.Y = 4
	engine.saveIntegration.SetGameState(engine.player, engine.dungeonManager)
	if err := engine.saveIntegration.SaveGame(); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}
	engine.player.Position.X = 0
	engine.player.Position.Y = 0

	result := command.Execute(&command.Context{
		Player:  engine.player,
		Level:   engine.dungeonManager.GetCurrentLevel(),
		Dungeon: engine.dungeonManager,
		Save:    engine.saveIntegration,
	}, command.Command{Type: command.CmdLoad})
	if result.Error {
		t.Fatalf("shared CmdLoad reported error: %s", result.Message)
	}

	if err := engine.SyncLoadedState(); err != nil {
		t.Fatalf("SyncLoadedState() error = %v", err)
	}
	if engine.player.Position.X != 3 || engine.player.Position.Y != 4 {
		t.Fatalf("engine player position = (%d, %d), want (3, 4)", engine.player.Position.X, engine.player.Position.Y)
	}
	if !engine.runActive {
		t.Fatal("engine did not remain active after successful load")
	}
	if engine.saveIntegration.HasSave() {
		t.Fatal("successful engine load did not consume the save")
	}
}

func TestEngineSyncLoadedStateFailsWhenSaveMissing(t *testing.T) {
	logger.Setup()
	t.Setenv("HOME", t.TempDir())
	engine := NewEngineWithSeed(12345)
	original := engine.player
	result := command.Execute(&command.Context{
		Player:  engine.player,
		Level:   engine.dungeonManager.GetCurrentLevel(),
		Dungeon: engine.dungeonManager,
		Save:    engine.saveIntegration,
	}, command.Command{Type: command.CmdLoad})
	if !result.Error {
		t.Fatal("shared CmdLoad without a save should report an error")
	}
	if !strings.Contains(result.Message, "Load failed:") {
		t.Fatalf("missing-save result = %q, want Load failed prefix", result.Message)
	}
	if engine.player != original {
		t.Fatal("failed load replaced the active player")
	}
}

func fatalEngineLevel() *dungeon.Level {
	level := &dungeon.Level{
		Width:       5,
		Height:      5,
		FloorNumber: 1,
		Tiles:       make([][]*dungeon.Tile, 5),
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*dungeon.Tile, 5)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = dungeon.NewTile(dungeon.TileFloor)
		}
	}
	monster := actor.NewMonster(2, 1, 'E')
	monster.Type.Speed = 1
	monster.Type.Level = 1000
	monster.Type.Damage = "100x100"
	monster.IsRunning = true
	level.Monsters = []*actor.Monster{monster}
	return level
}
