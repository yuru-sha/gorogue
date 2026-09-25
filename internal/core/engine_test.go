package core

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestNewEngineRegistersSaveLoadState(t *testing.T) {
	logger.Setup()
	engine := NewEngineWithSeed(12345)
	engine.stateManager.SetState(state.StateSaveLoad)
	grid := gruid.NewGrid(80, 50)
	engine.stateManager.Draw(&grid)

	var rows []string
	for y := 0; y < 50; y++ {
		var row strings.Builder
		for x := 0; x < 80; x++ {
			row.WriteRune(grid.At(gruid.Point{X: x, Y: y}).Rune)
		}
		rows = append(rows, row.String())
	}
	if !strings.Contains(strings.Join(rows, "\n"), "SAVE/LOAD GAME") {
		t.Fatal("save/load state is not registered")
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
	if !strings.Contains(rendered.String(), "R) Restart") {
		t.Fatal("game-over screen was not rendered after a fatal inventory action")
	}
}

func TestEngineRestartAfterFatalInventoryActionStartsFreshGame(t *testing.T) {
	logger.Setup()
	const seed int64 = 12345
	engine := NewEngineWithSeed(seed)
	engine.stateManager.SetState(state.StateGame)
	engine.player.Position.X = 1
	engine.player.Position.Y = 1
	engine.player.HP = 1
	engine.player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemFood, "ration", 1))
	engine.gameScreen.SetLevel(fatalEngineLevel())

	deadPlayer := engine.player
	engine.Update(gruid.MsgKeyDown{Key: "d"})
	engine.Update(gruid.MsgKeyDown{Key: "a"})

	var rendered strings.Builder
	grid := engine.Draw()
	for y := 0; y < 50; y++ {
		for x := 0; x < 80; x++ {
			rendered.WriteRune(grid.At(gruid.Point{X: x, Y: y}).Rune)
		}
	}
	if !strings.Contains(rendered.String(), "Death: Player died during a turn.") {
		t.Fatal("game-over score entry has no death reason")
	}

	engine.Update(gruid.MsgKeyDown{Key: "R"})
	if got := engine.stateManager.GetCurrentState(); got != state.StateGame {
		t.Fatalf("current state after restart = %v, want StateGame", got)
	}
	if engine.player == deadPlayer {
		t.Fatal("restart reused the dead player")
	}
	if !engine.player.IsAlive() {
		t.Fatal("restarted player is dead")
	}
	if got := engine.saveIntegration.GetGameInfo().Seed; got != seed {
		t.Fatalf("restart seed = %d, want %d", got, seed)
	}

	engine.Update(gruid.MsgKeyDown{Key: "."})
	if got := engine.stateManager.GetCurrentState(); got != state.StateGame {
		t.Fatalf("current state after first restarted input = %v, want StateGame", got)
	}
}

func TestEngineNewGameAfterGameOverDoesNotResumeDeadPlayer(t *testing.T) {
	logger.Setup()
	engine := NewEngineWithSeed(12345)
	engine.stateManager.SetState(state.StateGame)
	engine.player.Position.X = 1
	engine.player.Position.Y = 1
	engine.player.HP = 1
	engine.player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemFood, "ration", 1))
	engine.gameScreen.SetLevel(fatalEngineLevel())

	engine.Update(gruid.MsgKeyDown{Key: "d"})
	engine.Update(gruid.MsgKeyDown{Key: "a"})
	deadPlayer := engine.player
	engine.Update(gruid.MsgKeyDown{Key: "M"})
	if got := engine.stateManager.GetCurrentState(); got != state.StateMenu {
		t.Fatalf("current state after game over menu selection = %v, want StateMenu", got)
	}

	engine.Update(gruid.MsgKeyDown{Key: "n"})
	if got := engine.stateManager.GetCurrentState(); got != state.StateGame {
		t.Fatalf("current state after new game selection = %v, want StateGame", got)
	}
	if engine.player == deadPlayer {
		t.Fatal("new game reused the dead player")
	}
	if !engine.player.IsAlive() {
		t.Fatal("new game player is dead")
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
