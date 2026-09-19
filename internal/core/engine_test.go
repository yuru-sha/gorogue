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
	monster.Attack = 100
	level.Monsters = []*actor.Monster{monster}
	engine.gameScreen.SetLevel(level)

	engine.Update(gruid.MsgKeyDown{Key: "d"})
	if effect := engine.Update(gruid.MsgKeyDown{Key: "a"}); effect != nil {
		t.Fatal("fatal inventory action returned an end effect before rendering game over")
	}
	if got := engine.stateManager.GetCurrentState(); got != state.StateGameOver {
		t.Fatalf("current state = %v, want StateGameOver", got)
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
