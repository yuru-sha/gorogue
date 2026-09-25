package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
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

func occludedTestFloor() *dungeon.Level {
	level := &dungeon.Level{
		Width:  7,
		Height: 5,
		Tiles:  make([][]*dungeon.Tile, 5),
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*dungeon.Tile, level.Width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = dungeon.NewTile(dungeon.TileWall)
		}
	}
	for _, position := range [][2]int{{1, 2}, {2, 2}, {4, 2}} {
		level.SetTile(position[0], position[1], dungeon.TileFloor)
	}
	return level
}

func TestDrawHidesUnexploredTerrainAndEntities(t *testing.T) {
	player := actor.NewPlayer(1, 2)
	level := occludedTestFloor()
	monster := actor.NewMonster(4, 2, 'B')
	level.Monsters = []*actor.Monster{monster}
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(level)
	grid := gruid.NewGrid(80, 50)

	screen.Draw(&grid)

	if got := grid.At(gruid.Point{X: 4, Y: 4}).Rune; got != ' ' {
		t.Fatalf("unexplored tile was drawn as %q", got)
	}
	if got := grid.At(gruid.Point{X: 1, Y: 4}).Rune; got != '@' {
		t.Fatalf("player was drawn as %q, want '@'", got)
	}
}

func TestDrawShowsExploredTerrainWithoutOutOfFOVEntities(t *testing.T) {
	player := actor.NewPlayer(1, 2)
	level := occludedTestFloor()
	monster := actor.NewMonster(4, 2, 'B')
	level.Monsters = []*actor.Monster{monster}
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(level)
	level.UpdateVisibility(4, 2)
	level.UpdateVisibility(1, 2)
	grid := gruid.NewGrid(80, 50)

	screen.Draw(&grid)

	farTile := level.GetTile(4, 2)
	if farTile.Visible || !farTile.Explored {
		t.Fatalf("explored tile state = visible:%t explored:%t", farTile.Visible, farTile.Explored)
	}
	wantGlyph, _ := terrainAppearance(farTile.Type)
	if got := grid.At(gruid.Point{X: 4, Y: 4}).Rune; got != wantGlyph {
		t.Fatalf("explored terrain was drawn as %q, want %q", got, wantGlyph)
	}
	if got := grid.At(gruid.Point{X: 4, Y: 4}).Rune; got == monster.Type.Code {
		t.Fatal("out-of-FOV monster was drawn")
	}
}

func TestInvisibleMonsterRequiresSeeInvisibleToRender(t *testing.T) {
	player := actor.NewPlayer(1, 2)
	level := occludedTestFloor()
	monster := actor.NewMonster(2, 2, 'P')
	monster.IsInvisible = true
	level.Monsters = []*actor.Monster{monster}
	level.UpdateVisibility(player.Position.X, player.Position.Y)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(level)
	grid := gruid.NewGrid(80, 50)

	screen.Draw(&grid)
	position := gruid.Point{X: monster.Position.X, Y: monster.Position.Y + 2}
	if got := grid.At(position).Rune; got == monster.Type.Code {
		t.Fatal("invisible monster rendered without see-invisible")
	}

	player.SeeInvisibleTurns = 1
	screen.Draw(&grid)
	if got := grid.At(position).Rune; got != monster.Type.Code {
		t.Fatalf("monster rendered as %q with see-invisible active, want %q", got, monster.Type.Code)
	}
}

func TestReadIdentifyScrollSelectsTargetInGUI(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	scroll := item.NewItem(1, 1, item.ItemScroll, "identify potion", 1)
	potion := item.NewItem(1, 1, item.ItemPotion, "poison", 1)
	player.Inventory.AddItem(scroll)
	player.Inventory.AddItem(potion)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(occludedTestFloor())

	screen.HandleInput(gruid.MsgKeyDown{Key: "r"})
	screen.HandleInput(gruid.MsgKeyDown{Key: "a"})
	if screen.inputMode != ModeReadTarget || len(player.Inventory.Items) != 2 {
		t.Fatalf("after selecting identify scroll: mode=%d inventory=%d", screen.inputMode, len(player.Inventory.Items))
	}
	screen.HandleInput(gruid.MsgKeyDown{Key: "b"})
	if screen.inputMode != ModeNormal || len(player.Inventory.Items) != 1 ||
		player.Inventory.Items[0] != potion || !player.IdentifyMgr.IsIdentified(potion) {
		t.Fatalf("identify target state: mode=%d inventory=%v identified=%t",
			screen.inputMode, player.Inventory.Items, player.IdentifyMgr.IsIdentified(potion))
	}
}

func TestTabIsNotBoundToFOVToggle(t *testing.T) {
	player := actor.NewPlayer(1, 2)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(occludedTestFloor())

	screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyTab})

	if strings.Contains(strings.Join(screen.messages, "\n"), "FOV display toggled") {
		t.Fatal("Tab advertised an unimplemented FOV toggle")
	}
}

func TestClosingDoorHidesEntitiesBehindIt(t *testing.T) {
	player := actor.NewPlayer(1, 2)
	level := occludedTestFloor()
	level.SetTile(2, 2, dungeon.TileOpenDoor)
	level.SetTile(3, 2, dungeon.TileFloor)
	monster := actor.NewMonster(3, 2, 'B')
	level.Monsters = []*actor.Monster{monster}
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(level)
	if !level.GetTile(3, 2).Visible {
		t.Fatal("test setup did not put the monster in view")
	}

	screen.doCloseDoor(1, 0)

	if level.GetTile(3, 2).Visible {
		t.Fatal("closing a door left the tile behind it visible")
	}
	grid := gruid.NewGrid(80, 50)
	screen.Draw(&grid)
	if got := grid.At(gruid.Point{X: 3, Y: 4}).Rune; got == monster.Type.Code {
		t.Fatal("monster behind a closed door was drawn")
	}
}

func TestGUIAndCLIShareRepeatCommandHistory(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()

	player := actor.NewPlayerWithSeed(2, 2, 12345)
	screen := NewGameScreen(80, 50, player)
	screen.SetLevel(newTestFloor(5, 5))
	screen.cliMode.IsActive = true

	screen.executeCommand(command.NewMoveCommand(command.Direction{X: 1}))
	screen.cliMode.ExecuteCommand("move west")
	if player.Position.X != 2 || player.Position.Y != 2 {
		t.Fatalf("CLI move position=(%d,%d), want (2,2)", player.Position.X, player.Position.Y)
	}
	result := screen.executeCommand(command.Command{Type: command.CmdRepeat})
	if result.Error || player.Position.X != 1 || player.Position.Y != 2 {
		t.Fatalf("GUI repeat after CLI move = %+v, position=(%d,%d), want last shared action west",
			result, player.Position.X, player.Position.Y)
	}
}
