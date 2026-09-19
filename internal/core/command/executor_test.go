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
