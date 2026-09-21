package command

import (
	"strings"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestInventoryCommandsReportPlayerDeathAfterMonsterTurn(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		command   Command
		args      []string
		setupItem func(*actor.Player)
	}{
		{
			name:    "drop",
			command: Command{Type: CmdDrop},
			args:    []string{"a"},
			setupItem: func(player *actor.Player) {
				player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemFood, "ration", 1))
			},
		},
		{
			name:    "equip",
			command: Command{Type: CmdEquip},
			args:    []string{"a"},
			setupItem: func(player *actor.Player) {
				player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemWeapon, "sword", 1))
			},
		},
		{
			name:    "unequip",
			command: Command{Type: CmdUnequip},
			args:    []string{"weapon"},
			setupItem: func(player *actor.Player) {
				player.Equipment.EquipItem(gameitem.NewItem(1, 1, gameitem.ItemWeapon, "sword", 1))
			},
		},
		{
			name:    "eat",
			command: Command{Type: CmdEat},
			args:    []string{"a"},
			setupItem: func(player *actor.Player) {
				player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemFood, "ration", 1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := actor.NewPlayerWithSeed(1, 1, 42)
			player.HP = 1
			tt.setupItem(player)
			level := fatalMonsterLevel()

			result := Execute(&Context{Player: player, Level: level}, tt.command, tt.args...)

			if !result.PlayerDied {
				t.Fatal("result.PlayerDied = false, want true")
			}
			if player.IsAlive() {
				t.Fatal("player survived the fatal monster turn")
			}
			if !strings.Contains(result.Message, "You died.") {
				t.Fatalf("result message = %q, want death message", result.Message)
			}
		})
	}
}

func TestInventoryCommandSuccessDoesNotReportDeath(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemFood, "ration", 1))
	result := Execute(&Context{Player: player, Level: emptyTestLevel()}, Command{Type: CmdDrop}, "a")

	if result.PlayerDied {
		t.Fatal("successful drop reported player death")
	}
	if result.Message != "Dropped ration." {
		t.Fatalf("result message = %q, want %q", result.Message, "Dropped ration.")
	}
}

func TestStairCommandsLandOnOppositeStairs(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	const seed = 28
	player := actor.NewPlayerWithSeed(1, 1, seed)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)
	ctx := &Context{Player: player, Dungeon: dungeonManager, Level: dungeonManager.GetCurrentLevel()}

	_, downStairs := dungeon.NewStairsManager(ctx.Level).GetStairPositions()
	if len(downStairs) != 1 {
		t.Fatalf("expected one down stair on floor 1, got %d", len(downStairs))
	}
	player.Position.X = downStairs[0].X
	player.Position.Y = downStairs[0].Y

	descend := Execute(ctx, Command{Type: CmdGoDownstairs})
	if descend.Error {
		t.Fatalf("descending failed: %s", descend.Message)
	}
	if dungeonManager.GetCurrentFloor() != 2 {
		t.Fatalf("current floor = %d, want 2", dungeonManager.GetCurrentFloor())
	}
	assertPlayerAtStair(t, ctx, dungeon.TileStairsUp)

	ascend := Execute(ctx, Command{Type: CmdGoUpstairs})
	if ascend.Error {
		t.Fatalf("ascending failed: %s", ascend.Message)
	}
	if dungeonManager.GetCurrentFloor() != 1 {
		t.Fatalf("current floor = %d, want 1", dungeonManager.GetCurrentFloor())
	}
	assertPlayerAtStair(t, ctx, dungeon.TileStairsDown)
}

func TestReturnToSurfaceAndWinThroughGameplayCommands(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	const seed = 28
	player := actor.NewPlayerWithSeed(1, 1, seed)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)
	ctx := &Context{Player: player, Dungeon: dungeonManager, Level: dungeonManager.GetCurrentLevel()}

	for floor := 1; floor < dungeon.MaxFloors; floor++ {
		walkToStair(t, ctx, dungeon.TileStairsDown)
		result := Execute(ctx, Command{Type: CmdGoDownstairs})
		if result.Error {
			t.Fatalf("failed to descend from floor %d: %s", floor, result.Message)
		}
		if got := dungeonManager.GetCurrentFloor(); got != floor+1 {
			t.Fatalf("current floor = %d after descending from %d, want %d", got, floor, floor+1)
		}
		assertPlayerAtStair(t, ctx, dungeon.TileStairsUp)
	}
	player.Inventory.AddItem(gameitem.NewAmulet(0, 0))

	for floor := dungeon.MaxFloors; floor > 1; floor-- {
		if floor < dungeon.MaxFloors {
			walkToStair(t, ctx, dungeon.TileStairsUp)
		}

		result := Execute(ctx, Command{Type: CmdGoUpstairs})
		if result.Error {
			t.Fatalf("failed to ascend from floor %d: %s", floor, result.Message)
		}
		if result.Victory {
			t.Fatalf("won before reaching the surface exit from floor %d", floor)
		}
		if got := dungeonManager.GetCurrentFloor(); got != floor-1 {
			t.Fatalf("current floor = %d after ascending from %d, want %d", got, floor, floor-1)
		}
		assertPlayerAtStair(t, ctx, dungeon.TileStairsDown)
	}

	walkToStair(t, ctx, dungeon.TileStairsUp)
	result := Execute(ctx, Command{Type: CmdGoUpstairs})
	if result.Error || !result.Victory {
		t.Fatalf("surface exit result = %+v, want victory", result)
	}
}

func assertPlayerAtStair(t *testing.T, ctx *Context, expected dungeon.TileType) {
	t.Helper()
	tile := ctx.Level.GetTile(ctx.Player.Position.X, ctx.Player.Position.Y)
	if tile == nil || tile.Type != expected {
		if tile == nil {
			t.Fatalf("player landed off the map on floor %d, want stair type %d", ctx.Level.FloorNumber, expected)
		}
		t.Fatalf("player landed on tile type %d at (%d, %d) on floor %d, want stair type %d", tile.Type, ctx.Player.Position.X, ctx.Player.Position.Y, ctx.Level.FloorNumber, expected)
	}
}

func walkToStair(t *testing.T, ctx *Context, stairType dungeon.TileType) {
	t.Helper()
	level := ctx.Dungeon.GetCurrentLevel()
	level.Monsters = nil
	level.Items = nil
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile.Type == dungeon.TileDoor || tile.Type == dungeon.TileDoorClosed {
				level.SetTile(x, y, dungeon.TileOpenDoor)
			}
		}
	}

	upStairs, downStairs := dungeon.NewStairsManager(level).GetStairPositions()
	stairs := upStairs
	if stairType == dungeon.TileStairsDown {
		stairs = downStairs
	}
	if len(stairs) != 1 {
		t.Fatalf("floor %d has %d stairs of type %v, want one", level.FloorNumber, len(stairs), stairType)
	}

	start := dungeon.Position{X: ctx.Player.Position.X, Y: ctx.Player.Position.Y}
	target := stairs[0]
	moves := [...]Command{
		{Type: CmdMoveWest, Direction: Direction{X: -1}},
		{Type: CmdMoveEast, Direction: Direction{X: 1}},
		{Type: CmdMoveNorth, Direction: Direction{Y: -1}},
		{Type: CmdMoveSouth, Direction: Direction{Y: 1}},
	}
	type previousMove struct {
		position dungeon.Position
		command  Command
	}
	previous := map[dungeon.Position]previousMove{start: {}}
	queue := []dungeon.Position{start}
	for len(queue) > 0 {
		position := queue[0]
		queue = queue[1:]
		if position == target {
			break
		}
		for _, move := range moves {
			next := dungeon.Position{X: position.X + move.Direction.X, Y: position.Y + move.Direction.Y}
			if _, seen := previous[next]; seen || !level.IsInBounds(next.X, next.Y) || !level.GetTile(next.X, next.Y).Walkable() {
				continue
			}
			previous[next] = previousMove{position: position, command: move}
			queue = append(queue, next)
		}
	}
	if _, reached := previous[target]; !reached {
		t.Fatalf("no walkable route from %+v to stair %+v on floor %d", start, target, level.FloorNumber)
	}

	path := make([]Command, 0)
	for position := target; position != start; {
		step := previous[position]
		path = append(path, step.command)
		position = step.position
	}
	for i := len(path) - 1; i >= 0; i-- {
		result := Execute(ctx, path[i])
		if result.Error || result.PlayerDied {
			t.Fatalf("failed to walk to stair on floor %d: %+v", level.FloorNumber, result)
		}
	}
	assertPlayerAtStair(t, ctx, stairType)
}

func fatalMonsterLevel() *dungeon.Level {
	level := emptyTestLevel()
	monster := actor.NewMonster(2, 1, 'E')
	monster.Type.Speed = 1
	monster.Attack = 100
	level.Monsters = []*actor.Monster{monster}
	return level
}

func emptyTestLevel() *dungeon.Level {
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
	return level
}
