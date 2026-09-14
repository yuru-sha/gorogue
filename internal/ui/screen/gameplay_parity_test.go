package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/core/command"
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
			if len(guiLevel.Items) != len(cliLevel.Items) || len(guiLevel.Monsters) != len(cliLevel.Monsters) {
				t.Fatalf("level entity count mismatch: GUI=(items=%d monsters=%d) CLI=(items=%d monsters=%d)",
					len(guiLevel.Items), len(guiLevel.Monsters), len(cliLevel.Items), len(cliLevel.Monsters))
			}
			for i := range guiLevel.Monsters {
				guiMonster, cliMonster := guiLevel.Monsters[i], cliLevel.Monsters[i]
				if guiMonster.HP != cliMonster.HP || guiMonster.TurnCount != cliMonster.TurnCount {
					t.Fatalf("monster state mismatch: GUI=(hp=%d turns=%d) CLI=(hp=%d turns=%d)",
						guiMonster.HP, guiMonster.TurnCount, cliMonster.HP, cliMonster.TurnCount)
				}
			}
			if got := gui.messages[len(gui.messages)-1]; got != cliResult {
				t.Fatalf("message mismatch: GUI=%q CLI=%q\nall GUI messages: %s", got, cliResult, strings.Join(gui.messages, " | "))
			}
		})
	}
}

func TestPickupParityTracksTurnOnlyOnSuccess(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	run := func(capacity int) (int, int, int, int) {
		newLevel := func(player *actor.Player) *dungeon.Level {
			level := newTestFloor(5, 5)
			level.Items = []*item.Item{item.NewItem(1, 1, item.ItemFood, "parity food", 1)}
			monster := actor.NewMonster(3, 1, 'B')
			monster.Type.Speed = 100
			level.Monsters = []*actor.Monster{monster}
			return level
		}

		guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
		guiPlayer.Inventory.Capacity = capacity
		guiLevel := newLevel(guiPlayer)
		gui := NewGameScreen(80, 50, guiPlayer)
		gui.SetLevel(guiLevel)
		gui.HandleInput(gruid.MsgKeyDown{Key: ","})

		cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
		cliPlayer.Inventory.Capacity = capacity
		cliLevel := newLevel(cliPlayer)
		cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
		cliMode.IsActive = true
		cliMode.ExecuteCommand("pickup")

		return guiLevel.Monsters[0].TurnCount, cliLevel.Monsters[0].TurnCount,
			len(guiPlayer.Inventory.Items), len(cliPlayer.Inventory.Items)
	}

	guiTurn, cliTurn, guiItems, cliItems := run(1)
	if guiTurn != 1 || cliTurn != 1 || guiItems != 1 || cliItems != 1 {
		t.Fatalf("successful pickup mismatch: GUI=(turns=%d items=%d) CLI=(turns=%d items=%d)",
			guiTurn, guiItems, cliTurn, cliItems)
	}

	guiTurn, cliTurn, guiItems, cliItems = run(0)
	if guiTurn != 0 || cliTurn != 0 || guiItems != 0 || cliItems != 0 {
		t.Fatalf("failed pickup consumed a turn or changed inventory: GUI=(turns=%d items=%d) CLI=(turns=%d items=%d)",
			guiTurn, guiItems, cliTurn, cliItems)
	}
}

func TestGameplayCombatAdvancesMonstersAfterKill(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	newLevel := func(player *actor.Player) *dungeon.Level {
		level := newTestFloor(5, 5)
		front := actor.NewMonster(player.Position.X+1, player.Position.Y, 'B')
		front.HP = 1
		other := actor.NewMonster(player.Position.X+2, player.Position.Y, 'B')
		other.Type.Speed = 100
		level.Monsters = []*actor.Monster{front, other}
		return level
	}

	guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	guiLevel := newLevel(guiPlayer)
	gui := NewGameScreen(80, 50, guiPlayer)
	gui.SetLevel(guiLevel)
	gui.HandleInput(gruid.MsgKeyDown{Key: "l"})

	cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	cliLevel := newLevel(cliPlayer)
	cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
	cliMode.IsActive = true
	cliMode.ExecuteCommand("move east")

	if len(guiLevel.Monsters) != 1 || len(cliLevel.Monsters) != 1 {
		t.Fatalf("dead monster was not removed: GUI=%d CLI=%d", len(guiLevel.Monsters), len(cliLevel.Monsters))
	}
	if guiLevel.Monsters[0].TurnCount != 1 || cliLevel.Monsters[0].TurnCount != 1 {
		t.Fatalf("remaining monster did not receive a turn: GUI=%d CLI=%d", guiLevel.Monsters[0].TurnCount, cliLevel.Monsters[0].TurnCount)
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

func TestGameplayItemUseHasGUICLIParity(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	guiPlayer.Hunger = 40
	guiLevel := newTestFloor(5, 5)
	guiPlayer.Inventory.AddItem(item.NewItem(1, 1, item.ItemFood, "parity food", 12))
	gui := NewGameScreen(80, 50, guiPlayer)
	gui.SetLevel(guiLevel)
	gui.HandleInput(gruid.MsgKeyDown{Key: "e"})
	gui.HandleInput(gruid.MsgKeyDown{Key: "a"})

	cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	cliPlayer.Hunger = 40
	cliLevel := newTestFloor(5, 5)
	cliPlayer.Inventory.AddItem(item.NewItem(1, 1, item.ItemFood, "parity food", 12))
	cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
	cliMode.IsActive = true
	cliResult := cliMode.ExecuteCommand("use a")

	if guiPlayer.Hunger != cliPlayer.Hunger || len(guiPlayer.Inventory.Items) != len(cliPlayer.Inventory.Items) {
		t.Fatalf("item state mismatch: GUI=(hunger=%d inventory=%d) CLI=(hunger=%d inventory=%d)",
			guiPlayer.Hunger, len(guiPlayer.Inventory.Items), cliPlayer.Hunger, len(cliPlayer.Inventory.Items))
	}
	if got := gui.messages[len(gui.messages)-1]; got != cliResult {
		t.Fatalf("message mismatch: GUI=%q CLI=%q", got, cliResult)
	}
}

func TestGameplayPotionUseConsumesATurnForBothEntryPoints(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	newLevel := func(player *actor.Player) *dungeon.Level {
		level := newTestFloor(5, 5)
		monster := actor.NewMonster(player.Position.X+2, player.Position.Y, 'B')
		monster.Type.Speed = 100
		level.Monsters = []*actor.Monster{monster}
		return level
	}

	guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	guiPlayer.HP = 1
	guiPlayer.Inventory.AddItem(item.NewItem(1, 1, item.ItemPotion, "healing", 1))
	guiLevel := newLevel(guiPlayer)
	gui := NewGameScreen(80, 50, guiPlayer)
	gui.SetLevel(guiLevel)
	gui.HandleInput(gruid.MsgKeyDown{Key: "a"})
	gui.HandleInput(gruid.MsgKeyDown{Key: "a"})

	cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	cliPlayer.HP = 1
	cliPlayer.Inventory.AddItem(item.NewItem(1, 1, item.ItemPotion, "healing", 1))
	cliLevel := newLevel(cliPlayer)
	cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
	cliMode.IsActive = true
	cliMode.ExecuteCommand("use a")

	if guiPlayer.HP != cliPlayer.HP || guiLevel.Monsters[0].TurnCount != cliLevel.Monsters[0].TurnCount {
		t.Fatalf("potion state mismatch: GUI=(hp=%d turns=%d) CLI=(hp=%d turns=%d)",
			guiPlayer.HP, guiLevel.Monsters[0].TurnCount, cliPlayer.HP, cliLevel.Monsters[0].TurnCount)
	}
	if guiLevel.Monsters[0].TurnCount != 1 {
		t.Fatalf("potion use did not consume a turn: %d", guiLevel.Monsters[0].TurnCount)
	}
}

func TestAttackRejectsDistantMonster(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	player := actor.NewPlayerWithSeed(1, 1, 42)
	level := newTestFloor(5, 5)
	level.Monsters = []*actor.Monster{actor.NewMonster(3, 1, 'B')}
	cliMode := cli.NewCLIMode(level, player)
	cliMode.IsActive = true

	if got := cliMode.ExecuteCommand("attack 3 1"); got != "You can only attack an adjacent monster." {
		t.Fatalf("distant attack result = %q", got)
	}
}

func TestGameScreenSplitsMultilineMessages(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	screen := NewGameScreen(80, 50, actor.NewPlayerWithSeed(1, 1, 42))
	screen.messages = nil
	screen.AddMessage("first line\nsecond line")

	if got := strings.Join(screen.messages, "|"); got != "first line|second line" {
		t.Fatalf("message log = %q", got)
	}
}

func TestGameplayValidationHasGUICLIParity(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	guiLevel := newTestFloor(5, 5)
	guiLevel.SetTile(2, 1, dungeon.TileWall)
	gui := NewGameScreen(80, 50, guiPlayer)
	gui.SetLevel(guiLevel)
	gui.HandleInput(gruid.MsgKeyDown{Key: "l"})

	cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	cliLevel := newTestFloor(5, 5)
	cliLevel.SetTile(2, 1, dungeon.TileWall)
	cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
	cliMode.IsActive = true
	cliResult := cliMode.ExecuteCommand("move east")

	if got := gui.messages[len(gui.messages)-1]; got != cliResult {
		t.Fatalf("blocked movement message mismatch: GUI=%q CLI=%q", got, cliResult)
	}
	if guiPlayer.Position.X != cliPlayer.Position.X || guiPlayer.Position.Y != cliPlayer.Position.Y {
		t.Fatalf("blocked movement changed position: GUI=%v CLI=%v", guiPlayer.Position, cliPlayer.Position)
	}
}

func TestEquipmentReplacementPreservesStateThroughEitherEntryPoint(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name     string
		itemType item.ItemType
		oldName  string
		newName  string
		fillPack bool
	}{
		{name: "weapon", itemType: item.ItemWeapon, oldName: "old sword", newName: "new sword"},
		{name: "armor", itemType: item.ItemArmor, oldName: "old armor", newName: "new armor"},
		{name: "weapon with full inventory", itemType: item.ItemWeapon, oldName: "old sword", newName: "new sword", fillPack: true},
		{name: "armor with full inventory", itemType: item.ItemArmor, oldName: "old armor", newName: "new armor", fillPack: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setup := func(player *actor.Player) *item.Item {
				old := item.NewItem(1, 1, tc.itemType, tc.oldName, 1)
				candidate := item.NewItem(1, 1, tc.itemType, tc.newName, 1)
				player.Equipment.EquipItem(old)
				if tc.fillPack {
					for i := 0; i < player.Inventory.Capacity-1; i++ {
						player.Inventory.AddItem(item.NewItem(1, 1, item.ItemPotion, "filler", 1))
					}
				}
				player.Inventory.AddItem(candidate)
				return candidate
			}

			guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			guiCandidate := setup(guiPlayer)
			gui := NewGameScreen(80, 50, guiPlayer)
			gui.SetLevel(newTestFloor(5, 5))
			gui.HandleInput(gruid.MsgKeyDown{Key: "w"})
			gui.HandleInput(gruid.MsgKeyDown{Key: "a"})

			cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			cliCandidate := setup(cliPlayer)
			cliMode := cli.NewCLIMode(newTestFloor(5, 5), cliPlayer)
			cliMode.IsActive = true
			cliResult := cliMode.ExecuteCommand("equip " + string(rune('a'+len(cliPlayer.Inventory.Items)-1)))

			check := func(player *actor.Player, candidate *item.Item) {
				var equipped *item.Item
				if tc.itemType == item.ItemWeapon {
					equipped = player.Equipment.Weapon
				} else {
					equipped = player.Equipment.Armor
				}
				if equipped == nil || equipped.Name != tc.oldName {
					t.Fatalf("equipped item changed: got %v", equipped)
				}
				preserved := false
				for _, inventoryItem := range player.Inventory.Items {
					if inventoryItem == candidate {
						preserved = true
						break
					}
				}
				if !preserved {
					t.Fatal("replacement item was removed from inventory")
				}
				expectedSize := 1
				if tc.fillPack {
					expectedSize = player.Inventory.Capacity
				}
				if len(player.Inventory.Items) != expectedSize {
					t.Fatalf("inventory changed: got %d, want %d", len(player.Inventory.Items), expectedSize)
				}
			}
			check(guiPlayer, guiCandidate)
			check(cliPlayer, cliCandidate)
			if got := gui.messages[len(gui.messages)-1]; got != cliResult {
				t.Fatalf("message mismatch: GUI=%q CLI=%q", got, cliResult)
			}
		})
	}
}

func TestCursedEquipmentCannotBeUnequippedThroughEitherEntryPoint(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	guiLevel := newTestFloor(5, 5)
	guiCursedWeapon := item.NewItem(1, 1, item.ItemWeapon, "cursed sword", 1)
	guiCursedWeapon.IsCursed = true
	guiPlayer.Equipment.EquipItem(guiCursedWeapon)
	gui := NewGameScreen(80, 50, guiPlayer)
	gui.SetLevel(guiLevel)
	gui.HandleInput(gruid.MsgKeyDown{Key: "t"})
	gui.HandleInput(gruid.MsgKeyDown{Key: "w"})

	cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
	cliLevel := newTestFloor(5, 5)
	cliCursedWeapon := item.NewItem(1, 1, item.ItemWeapon, "cursed sword", 1)
	cliCursedWeapon.IsCursed = true
	cliPlayer.Equipment.EquipItem(cliCursedWeapon)
	cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
	cliMode.IsActive = true
	cliResult := cliMode.ExecuteCommand("unequip weapon")

	if guiPlayer.Equipment.Weapon == nil || cliPlayer.Equipment.Weapon == nil {
		t.Fatal("cursed equipment was removed")
	}
	if got := gui.messages[len(gui.messages)-1]; got != cliResult {
		t.Fatalf("cursed equipment message mismatch: GUI=%q CLI=%q", got, cliResult)
	}
}

func TestInventoryCommandsAdvanceMonstersThroughEitherEntryPoint(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		guiKeys    []gruid.Key
		cliCommand string
		setup      func(*actor.Player)
		check      func(*testing.T, *actor.Player, *dungeon.Level)
	}{
		{
			name:       "drop",
			guiKeys:    []gruid.Key{"d", "a"},
			cliCommand: "drop a",
			setup: func(player *actor.Player) {
				player.Inventory.AddItem(item.NewItem(1, 1, item.ItemFood, "ration", 1))
			},
			check: func(t *testing.T, player *actor.Player, level *dungeon.Level) {
				if player.Inventory.Size() != 0 || len(level.Items) != 1 {
					t.Fatalf("drop state = inventory %d, level items %d", player.Inventory.Size(), len(level.Items))
				}
			},
		},
		{
			name:       "equip",
			guiKeys:    []gruid.Key{"w", "a"},
			cliCommand: "equip a",
			setup: func(player *actor.Player) {
				player.Inventory.AddItem(item.NewItem(1, 1, item.ItemWeapon, "sword", 1))
			},
			check: func(t *testing.T, player *actor.Player, level *dungeon.Level) {
				if player.Inventory.Size() != 0 || player.Equipment.Weapon == nil {
					t.Fatalf("equip state = inventory %d, weapon %v", player.Inventory.Size(), player.Equipment.Weapon)
				}
			},
		},
		{
			name:       "unequip",
			guiKeys:    []gruid.Key{"t", "w"},
			cliCommand: "unequip weapon",
			setup: func(player *actor.Player) {
				player.Equipment.EquipItem(item.NewItem(1, 1, item.ItemWeapon, "sword", 1))
			},
			check: func(t *testing.T, player *actor.Player, level *dungeon.Level) {
				if player.Inventory.Size() != 1 || player.Equipment.Weapon != nil {
					t.Fatalf("unequip state = inventory %d, weapon %v", player.Inventory.Size(), player.Equipment.Weapon)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			tt.setup(guiPlayer)
			guiLevel := newTestFloor(5, 5)
			guiMonster := actor.NewMonster(3, 1, 'O')
			guiMonster.Type.Speed = 2
			guiLevel.Monsters = []*actor.Monster{guiMonster}
			gui := NewGameScreen(80, 50, guiPlayer)
			gui.SetLevel(guiLevel)
			for _, key := range tt.guiKeys {
				gui.HandleInput(gruid.MsgKeyDown{Key: key})
			}

			cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			tt.setup(cliPlayer)
			cliLevel := newTestFloor(5, 5)
			cliMonster := actor.NewMonster(3, 1, 'O')
			cliMonster.Type.Speed = 2
			cliLevel.Monsters = []*actor.Monster{cliMonster}
			cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
			cliMode.IsActive = true
			cliResult := cliMode.ExecuteCommand(tt.cliCommand)

			if guiMonster.TurnCount != 1 || cliMonster.TurnCount != 1 {
				t.Fatalf("successful %s turn count = GUI %d, CLI %d; want 1", tt.name, guiMonster.TurnCount, cliMonster.TurnCount)
			}
			tt.check(t, guiPlayer, guiLevel)
			tt.check(t, cliPlayer, cliLevel)
			if got := gui.messages[len(gui.messages)-1]; got != cliResult {
				t.Fatalf("message mismatch: GUI=%q CLI=%q", got, cliResult)
			}
		})
	}
}

func TestFailedInventoryCommandsDoNotAdvanceMonstersThroughEitherEntryPoint(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		commandType command.Type
		args        []string
		cliCommand  string
		setup       func(*actor.Player)
	}{
		{
			name:        "drop missing item",
			commandType: command.CmdDrop,
			args:        []string{"a"},
			cliCommand:  "drop a",
		},
		{
			name:        "equip food",
			commandType: command.CmdEquip,
			args:        []string{"a"},
			cliCommand:  "equip a",
			setup: func(player *actor.Player) {
				player.Inventory.AddItem(item.NewItem(1, 1, item.ItemFood, "ration", 1))
			},
		},
		{
			name:        "unequip empty slot",
			commandType: command.CmdUnequip,
			args:        []string{"weapon"},
			cliCommand:  "unequip weapon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guiPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			if tt.setup != nil {
				tt.setup(guiPlayer)
			}
			guiLevel := newTestFloor(5, 5)
			guiMonster := actor.NewMonster(3, 1, 'O')
			guiMonster.Type.Speed = 2
			guiLevel.Monsters = []*actor.Monster{guiMonster}
			gui := NewGameScreen(80, 50, guiPlayer)
			gui.SetLevel(guiLevel)
			guiResult := gui.executeCommand(command.Command{Type: tt.commandType}, tt.args...)

			cliPlayer := actor.NewPlayerWithSeed(1, 1, 42)
			if tt.setup != nil {
				tt.setup(cliPlayer)
			}
			cliLevel := newTestFloor(5, 5)
			cliMonster := actor.NewMonster(3, 1, 'O')
			cliMonster.Type.Speed = 2
			cliLevel.Monsters = []*actor.Monster{cliMonster}
			cliMode := cli.NewCLIMode(cliLevel, cliPlayer)
			cliMode.IsActive = true
			cliResult := cliMode.ExecuteCommand(tt.cliCommand)

			if guiMonster.TurnCount != 0 || cliMonster.TurnCount != 0 {
				t.Fatalf("failed %s advanced monsters: GUI %d, CLI %d; want 0", tt.name, guiMonster.TurnCount, cliMonster.TurnCount)
			}
			if guiResult.Message != cliResult {
				t.Fatalf("message mismatch: GUI=%q CLI=%q", guiResult.Message, cliResult)
			}
		})
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
	cliMissingLoad := cliMode.ExecuteCommand("load")
	guiMissingLoad := NewSaveLoadScreenWithIntegration(80, 50, integration)
	guiMissingLoad.SetOnLoad(func(*actor.Player, *dungeon.DungeonManager) {})
	guiMissingLoad.SetSelectedOption(1)
	guiMissingLoad.HandleInput(gruid.MsgKeyDown{Key: gruid.KeyEnter})
	if guiMissingLoad.GetMessage() != cliMissingLoad {
		t.Fatalf("missing save message mismatch: GUI=%q CLI=%q", guiMissingLoad.GetMessage(), cliMissingLoad)
	}

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
