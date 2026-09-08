package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
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

func newTestFloor(width, height int) *dungeon.Level {
	level := &dungeon.Level{
		Width:       width,
		Height:      height,
		FloorNumber: 1,
		Tiles:       make([][]*dungeon.Tile, height),
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*dungeon.Tile, width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = dungeon.NewTile(dungeon.TileFloor)
		}
	}
	return level
}

func TestOpenDoorAcceptsGeneratedDoor(t *testing.T) {
	player := actor.NewPlayer(1, 1)
	level := newTestFloor(3, 3)
	level.SetTile(2, 1, dungeon.TileDoor)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(level)

	screen.doOpenDoor(1, 0)

	if got := level.GetTile(2, 1).Type; got != dungeon.TileOpenDoor {
		t.Fatalf("door type = %v, want open door", got)
	}
}

func TestEatingAdvancesMonsterTurn(t *testing.T) {
	player := actor.NewPlayer(1, 1)
	level := newTestFloor(5, 5)
	monster := actor.NewMonster(3, 3, 'B')
	monster.Type.Speed = 100
	level.Monsters = []*actor.Monster{monster}
	food := item.NewFood(1, 1)
	player.Inventory.AddItem(food)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(level)
	screen.inputMode = ModeEat

	screen.handleEatInput(gruid.Key("a"))

	if monster.TurnCount != 1 {
		t.Fatalf("monster turn count = %d, want 1", monster.TurnCount)
	}
	if player.Inventory.GetItem(0) != nil {
		t.Fatal("eaten food was not removed")
	}
}
