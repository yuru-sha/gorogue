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
