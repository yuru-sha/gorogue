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

func TestSurfaceExitWithAmuletWins(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}

	player := actor.NewPlayerWithSeed(1, 1, 42)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, 42)
	ctx := &Context{Player: player, Dungeon: dungeonManager, Level: dungeonManager.GetCurrentLevel()}
	upStairs, _ := dungeon.NewStairsManager(ctx.Level).GetStairPositions()
	if len(upStairs) != 1 {
		t.Fatalf("surface has %d upstairs, want one", len(upStairs))
	}
	player.Position.X, player.Position.Y = upStairs[0].X, upStairs[0].Y
	player.Inventory.AddItem(gameitem.NewAmulet(0, 0))

	result := Execute(ctx, Command{Type: CmdGoUpstairs})
	if result.Error || !result.Victory {
		t.Fatalf("surface exit result = %+v, want victory", result)
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

func TestSteppingOnRustTrapRevealsItAndDamagesArmor(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 42)
	armor := gameitem.NewItem(0, 0, gameitem.ItemArmor, "leather armor", 1)
	if !player.Equipment.EquipItem(armor) {
		t.Fatal("failed to equip test armor")
	}
	level := emptyTestLevel()
	trap := &dungeon.Trap{Type: dungeon.TrapRust, Position: dungeon.Position{X: 2, Y: 1}}
	level.Traps = []*dungeon.Trap{trap}
	level.Seed = 42
	if err := level.SetRandomDraws(0); err != nil {
		t.Fatal(err)
	}
	originalDefense := armor.Defense

	result := Execute(&Context{Player: player, Level: level}, Command{
		Type:      CmdMoveEast,
		Direction: Direction{X: 1},
	})

	if !result.TurnConsumed || result.PlayerDied {
		t.Fatalf("trap movement result = %+v, want a live consumed turn", result)
	}
	if !trap.Discovered {
		t.Fatal("stepping on trap did not reveal it")
	}
	if armor.Defense != originalDefense-1 {
		t.Fatalf("armor defense = %d, want %d after rust trap", armor.Defense, originalDefense-1)
	}
}

func TestCallCommandNamesWithoutIdentifyingItem(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 42)
	potion := gameitem.NewItem(0, 0, gameitem.ItemPotion, "healing", 130)
	if !player.Inventory.AddItem(potion) {
		t.Fatal("failed to add test potion")
	}

	result := Execute(&Context{Player: player, Level: emptyTestLevel()},
		Command{Type: CmdCall}, "a", "blue fizz")

	if result.Error || result.TurnConsumed {
		t.Fatalf("call result = %+v, want successful free command", result)
	}
	if player.IdentifyMgr.IsIdentified(potion) {
		t.Fatal("calling an item revealed its identity")
	}
	if got := player.IdentifyMgr.GetDisplayName(potion); !strings.Contains(got, "blue fizz") {
		t.Fatalf("called item display = %q, want the call name", got)
	}
}

func TestIdentifyScrollWaitsForCategoryTargetBeforeConsuming(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 42)
	scroll := gameitem.NewItem(1, 1, gameitem.ItemScroll, "identify potion", 1)
	potion := gameitem.NewItem(1, 1, gameitem.ItemPotion, "poison", 1)
	player.Inventory.AddItem(scroll)
	player.Inventory.AddItem(potion)
	ctx := &Context{Player: player, Level: emptyTestLevel()}

	request := Execute(ctx, Command{Type: CmdRead}, "a")
	if !request.NeedsItemSelection || request.TurnConsumed || len(player.Inventory.Items) != 2 {
		t.Fatalf("identify selection request = %+v, inventory size %d", request, len(player.Inventory.Items))
	}
	result := Execute(ctx, Command{Type: CmdRead}, "a", "b")
	if result.Error || !result.TurnConsumed || len(player.Inventory.Items) != 1 ||
		player.Inventory.Items[0] != potion || !player.IdentifyMgr.IsIdentified(potion) {
		t.Fatalf("identify target result = %+v, inventory=%v, identified=%t",
			result, player.Inventory.Items, player.IdentifyMgr.IsIdentified(potion))
	}
}

func TestMagicDetectionPotionReadsCurrentFloorItems(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemPotion, "magic detection", 1))
	level := emptyTestLevel()
	level.Items = []*gameitem.Item{gameitem.NewItem(3, 3, gameitem.ItemWand, "light", 1)}

	result := Execute(&Context{Player: player, Level: level}, Command{Type: CmdQuaff}, "a")
	if result.Error || !result.TurnConsumed || !strings.Contains(result.Message, "sense magic") {
		t.Fatalf("magic detection result = %+v, want a successful current-floor detection", result)
	}
}

func TestThrowingOneItemFromStackDropsOnlyOne(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	arrows := gameitem.NewItem(1, 1, gameitem.ItemWeapon, "arrow", 10)
	arrows.Quantity = 10
	player.Inventory.AddItem(arrows)

	result := Execute(&Context{Player: player, Level: emptyTestLevel()},
		Command{Type: CmdThrow, Direction: Direction{X: 1}}, "a")
	if result.Error || !result.TurnConsumed || arrows.Quantity != 9 {
		t.Fatalf("throw result = %+v, remaining stack quantity = %d", result, arrows.Quantity)
	}
	dropped := result.Level.GetItemAt(4, 1)
	if dropped == nil || dropped.Quantity != 1 {
		t.Fatalf("dropped stack = %v, want exactly one projectile", dropped)
	}
}

func TestHasteSkipsEveryOtherMonsterTurn(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Inventory.AddItem(gameitem.NewItem(1, 1, gameitem.ItemPotion, "haste self", 1))
	monster := actor.NewMonster(3, 1, 'O')
	monster.IsSlowed = true
	monster.TurnCount = 0
	monster.IsRunning = true
	level := emptyTestLevel()
	level.Monsters = []*actor.Monster{monster}
	ctx := &Context{Player: player, Level: level}

	Execute(ctx, Command{Type: CmdQuaff}, "a")
	if monster.TurnCount != 1 || player.HasteTurns == 0 || !player.HasteSkipMonsterTurn {
		t.Fatalf("state after haste potion: monster turn=%d haste=%d skip=%t",
			monster.TurnCount, player.HasteTurns, player.HasteSkipMonsterTurn)
	}
	turnsLeft := player.HasteTurns

	firstWait := Execute(ctx, Command{Type: CmdWait})
	if !firstWait.TurnConsumed || monster.TurnCount != 1 ||
		player.HasteTurns != turnsLeft-1 || player.HasteSkipMonsterTurn {
		t.Fatalf("haste extra action state: result=%+v monster turn=%d haste=%d skip=%t",
			firstWait, monster.TurnCount, player.HasteTurns, player.HasteSkipMonsterTurn)
	}

	secondWait := Execute(ctx, Command{Type: CmdWait})
	if !secondWait.TurnConsumed || monster.TurnCount != 0 {
		t.Fatalf("next monster turn: turn=%d haste=%d skip=%t active=%t running=%t held=%t pos=%v HP=%d alive=%t",
			monster.TurnCount, player.HasteTurns, player.HasteSkipMonsterTurn, monster.IsActive,
			monster.IsRunning, monster.IsHeld, monster.Position, player.HP, player.IsAlive())
	}
}

func TestZapAwardsExperienceWhenWandKillsMonster(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()

	player := actor.NewPlayerWithSeed(1, 1, 42)
	wand := gameitem.NewItem(1, 1, gameitem.ItemWand, "magic missile", 1)
	wand.Charges = 1
	player.Inventory.AddItem(wand)
	monster := actor.NewMonster(2, 1, 'B')
	monster.HP = 1
	level := emptyTestLevel()
	level.Monsters = []*actor.Monster{monster}

	result := Execute(&Context{Player: player, Level: level},
		Command{Type: CmdZap, Direction: Direction{X: 1}}, "a")

	if !result.TurnConsumed || player.Exp != monster.Type.Experience {
		t.Fatalf("zap result=%+v experience=%d, want kill reward %d",
			result, player.Exp, monster.Type.Experience)
	}
	if strings.Contains(result.Message, "You gain") == false {
		t.Fatalf("zap kill message %q omits the experience award", result.Message)
	}
}

func TestThrowUsesMissileDamageAndConsumesHitProjectile(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()

	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Level = 100
	player.Strength = 31
	arrow := gameitem.NewItem(1, 1, gameitem.ItemWeapon, "arrow", 1)
	arrow.ItemID = 107
	arrow.Enchantment = 1
	player.Inventory.AddItem(arrow)
	bow := gameitem.NewItem(1, 1, gameitem.ItemWeapon, "short bow", 1)
	bow.ItemID = 104
	bow.Enchantment = 1
	player.Equipment.Weapon = bow
	monster := actor.NewMonster(2, 1, 'B')
	monster.HP = 100
	level := emptyTestLevel()
	level.Monsters = []*actor.Monster{monster}

	result := Execute(&Context{Player: player, Level: level},
		Command{Type: CmdThrow, Direction: Direction{X: 1}}, "a")

	damage := 100 - monster.HP
	if !result.TurnConsumed || damage < 10 || damage > 14 {
		t.Fatalf("throw result=%+v damage=%d; want source 2x3 plus strength and launcher bonuses", result, damage)
	}
	if !player.Inventory.IsEmpty() {
		t.Fatal("hit projectile remained in inventory")
	}
	if len(level.Items) != 0 {
		t.Fatalf("hit projectile remained on floor: %v", level.Items)
	}
}

func TestSearchingRingAutomaticallyFindsNearbyTrap(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Equipment.RingLeft = gameitem.NewItem(0, 0, gameitem.ItemRing, "searching", 0)
	level := emptyTestLevel()
	trap := &dungeon.Trap{Type: dungeon.TrapArrow, Position: dungeon.Position{X: 2, Y: 1}}
	level.Traps = []*dungeon.Trap{trap}
	ctx := &Context{Player: player, Level: level}

	for range 100 {
		Execute(ctx, Command{Type: CmdWait})
		if trap.Discovered {
			return
		}
	}
	t.Fatal("searching ring did not reveal adjacent trap")
}

func TestTeleportationRingEventuallyTeleportsPlayer(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Equipment.RingLeft = gameitem.NewItem(0, 0, gameitem.ItemRing, "teleportation", 0)
	level := emptyTestLevel()
	ctx := &Context{Player: player, Level: level}

	for range 1000 {
		before := *player.Position
		Execute(ctx, Command{Type: CmdWait})
		if *player.Position != before {
			return
		}
	}
	t.Fatal("teleportation ring did not move player")
}

func TestAggravateMonsterRingMakesMonstersRun(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	ring := gameitem.NewItem(0, 0, gameitem.ItemRing, "aggravate monster", 0)
	player.Inventory.AddItem(ring)
	level := emptyTestLevel()
	monster := actor.NewMonster(3, 3, 'O')
	level.Monsters = []*actor.Monster{monster}

	outcome := Execute(&Context{Player: player, Level: level}, Command{Type: CmdRingOn}, "a")
	if !outcome.TurnConsumed {
		t.Fatal("ring equip did not consume a turn")
	}
	if !monster.IsRunning {
		t.Fatal("aggravate monster ring did not make monster run")
	}
}

func TestAddStrengthRingChangesStrengthUntilRemoved(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	initialStrength := player.Strength
	ring := gameitem.NewItem(0, 0, gameitem.ItemRing, "add strength", 3)
	ring.Enchantment = 3
	player.Inventory.AddItem(ring)
	level := emptyTestLevel()
	ctx := &Context{Player: player, Level: level}

	equipped := Execute(ctx, Command{Type: CmdRingOn}, "a")
	if !equipped.TurnConsumed || player.Strength != initialStrength+3 {
		t.Fatalf("equip outcome=%+v strength=%d, want %d", equipped, player.Strength, initialStrength+3)
	}
	removed := Execute(ctx, Command{Type: CmdRingOff}, "left")
	if !removed.TurnConsumed || player.Strength != initialStrength {
		t.Fatalf("remove outcome=%+v strength=%d, want %d", removed, player.Strength, initialStrength)
	}
}

func TestDefeatingFlytrapReleasesHeldPlayer(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Held = true
	level := emptyTestLevel()
	flytrap := actor.NewMonster(2, 1, 'F')
	flytrap.TakeDamage(flytrap.HP)
	level.Monsters = []*actor.Monster{flytrap}

	Execute(&Context{Player: player, Level: level}, Command{Type: CmdWait})
	if player.Held {
		t.Fatal("dead flytrap continued holding the player")
	}
}

func TestAttackingMonsterWakesIt(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Level = 100
	player.Strength = 31
	level := emptyTestLevel()
	monster := actor.NewMonster(2, 1, 'O')
	monster.HP = 100
	monster.IsRunning = false
	monster.IsHeld = true
	level.Monsters = []*actor.Monster{monster}

	Execute(&Context{Player: player, Level: level}, Command{Type: CmdFight, Direction: Direction{X: 1}}, "east")
	if !monster.IsRunning {
		t.Fatal("attacked monster remained asleep")
	}
	if monster.IsHeld {
		t.Fatal("attacking monster did not release its held state")
	}
}
