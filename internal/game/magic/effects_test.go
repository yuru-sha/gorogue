package magic

import (
	"github.com/yuru-sha/gorogue/internal/utils/logger"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestPoisonPotionReducesStrengthUnlessSustained(t *testing.T) {
	logger.Setup()
	player := actor.NewPlayerWithSeed(0, 0, 29)
	originalStrength, originalHP := player.Strength, player.HP
	result := UsePotion("poison", player)
	if !result.Success || !result.Identified {
		t.Fatalf("poison potion result = %+v, want successful identified source effect", result)
	}
	if loss := originalStrength - player.Strength; loss < 1 || loss > 3 {
		t.Errorf("poison strength loss = %d, want source roll 1..3", loss)
	}
	if player.HP != originalHP {
		t.Errorf("poison changed HP from %d to %d; Rogue poison reduces strength", originalHP, player.HP)
	}

	protected := actor.NewPlayerWithSeed(0, 0, 29)
	protected.Equipment.RingLeft = item.NewItem(0, 0, item.ItemRing, "sustain strength", 280)
	originalStrength, originalHP = protected.Strength, protected.HP
	result = UsePotion("poison", protected)
	if !result.Success || protected.Strength != originalStrength || protected.HP != originalHP {
		t.Errorf("sustained poison result = %+v, strength %d/%d, HP %d/%d; want no stat or HP loss", result, protected.Strength, originalStrength, protected.HP, originalHP)
	}
}

func TestIdentifyScrollExposesTargetSelectionAndIdentifiesChosenType(t *testing.T) {
	player := actor.NewPlayerWithSeed(0, 0, 19)
	request := UseScrollOnItem("identify potion", player, nil, nil)
	if !request.Success || !request.Identified || !request.NeedsItemSelection || len(request.TargetTypes) != 1 || request.TargetTypes[0] != item.ItemPotion {
		t.Fatalf("identify potion request = %+v, want successful selection for potions", request)
	}

	selected := item.NewItem(0, 0, item.ItemPotion, "poison", 5)
	result := UseScrollOnItem("identify potion", player, nil, selected)
	if !result.Success || result.NeedsItemSelection || !selected.IsIdentified || !player.IdentifyMgr.IsIdentified(selected) {
		t.Fatalf("identify potion result = %+v, item identified=%t", result, selected.IsIdentified)
	}
	if !player.IdentifyMgr.IsIdentified(item.NewItem(0, 0, item.ItemPotion, "poison", 5)) {
		t.Fatal("identify scroll did not reveal the source type globally")
	}
}

func TestSourcePotionStatusesAndStrengthEffects(t *testing.T) {
	player := actor.NewPlayerWithSeed(0, 0, 48)
	strength := player.Strength
	if result := UsePotion("gain strength", player); !result.Success || player.Strength != strength+1 {
		t.Fatalf("gain-strength result = %+v, strength %d -> %d; want +1", result, strength, player.Strength)
	}
	if result := UsePotion("restore strength", player); !result.Success || player.Strength != player.MaxStrength {
		t.Fatalf("restore-strength result = %+v, strength %d/%d", result, player.Strength, player.MaxStrength)
	}
	if result := UsePotion("hallucination", player); !result.Success || player.HallucinationTurns != 850 {
		t.Fatalf("hallucination result = %+v, turns = %d, want 850", result, player.HallucinationTurns)
	}
	if result := UsePotion("blindness", player); !result.Success || player.BlindTurns != 850 {
		t.Fatalf("blindness result = %+v, turns = %d, want 850", result, player.BlindTurns)
	}
	if result := UsePotion("see invisible", player); !result.Success || player.SeeInvisibleTurns != 850 {
		t.Fatalf("see-invisible result = %+v, turns = %d, want 850", result, player.SeeInvisibleTurns)
	}
	if result := UsePotion("levitation", player); !result.Success || player.LevitationTurns != 30 {
		t.Fatalf("levitation result = %+v, turns = %d, want 30", result, player.LevitationTurns)
	}
}

func TestCuringPotionsEndHallucination(t *testing.T) {
	for _, potion := range []string{"extra healing", "poison"} {
		t.Run(potion, func(t *testing.T) {
			if err := logger.Setup(); err != nil {
				t.Fatal(err)
			}
			defer logger.Cleanup()
			player := actor.NewPlayerWithSeed(0, 0, 17)
			player.HallucinationTurns = 5
			level := &dungeon.Level{
				Width:  1,
				Height: 1,
				Tiles: [][]*dungeon.Tile{{
					dungeon.NewTile(dungeon.TileStairsDown),
				}},
			}

			UsePotionOnLevel(potion, player, level)

			if player.HallucinationTurns != 0 {
				t.Fatalf("%s left hallucination active for %d turns", potion, player.HallucinationTurns)
			}
			if !level.GetTile(0, 0).HallucinationKnown {
				t.Fatalf("%s did not remember visible stairs after curing hallucination", potion)
			}
		})
	}
}

func TestSourceScrollEffectsChangeGameplayState(t *testing.T) {
	player := actor.NewPlayerWithSeed(0, 0, 55)
	if result := UseScroll("monster confusion", player, nil); !result.Success || !player.CanConfuse {
		t.Fatalf("monster-confusion result = %+v, ability = %t", result, player.CanConfuse)
	}
	if result := UseScroll("sleep", player, nil); !result.Success || player.NoCommandTurns < 4 || player.NoCommandTurns > 8 {
		t.Fatalf("sleep result = %+v, no-command turns = %d, want 4..8", result, player.NoCommandTurns)
	}

	armor := item.NewItem(0, 0, item.ItemArmor, "chain mail", 75)
	player.Equipment.Armor = armor
	result := UseScroll("protect armor", player, nil)
	if !result.Success || !armor.IsProtected {
		t.Fatalf("protect-armor result = %+v, armor protection = %t", result, armor.IsProtected)
	}
	weapon := item.NewItem(0, 0, item.ItemWeapon, "mace", 8)
	armor.IsCursed, weapon.IsCursed = true, true
	player.Equipment.Weapon = weapon
	loose := item.NewItem(0, 0, item.ItemRing, "teleportation", 30)
	loose.IsCursed = true
	player.Inventory.AddItem(loose)
	result = UseScroll("remove curse", player, nil)
	if !result.Success || armor.IsCursed || weapon.IsCursed || !loose.IsCursed {
		t.Fatalf("remove-curse result = %+v, equipment curses %t/%t, carried curse %t; Rogue only uncurses equipped objects", result, armor.IsCursed, weapon.IsCursed, loose.IsCursed)
	}
}

func TestLightWandIlluminatesRoomAndUsesCharge(t *testing.T) {
	player := actor.NewPlayerWithSeed(0, 0, 61)
	player.Position.X, player.Position.Y = 2, 2
	room := &dungeon.Room{X: 0, Y: 0, Width: 5, Height: 5, IsDark: true}
	level := &dungeon.Level{Width: 5, Height: 5, Rooms: []*dungeon.Room{room}}
	level.Tiles = make([][]*dungeon.Tile, level.Height)
	for y := range level.Height {
		level.Tiles[y] = make([]*dungeon.Tile, level.Width)
		for x := range level.Width {
			level.Tiles[y][x] = &dungeon.Tile{Type: dungeon.TileFloor, IsWalkable: true}
		}
	}
	level.GetTile(3, 3).Type = dungeon.TileStairsDown
	wand := item.NewItem(0, 0, item.ItemWand, "light", 250)
	wand.Charges = 1
	result := UseWand(wand, player, level, 0, 0)
	if !result.Success || room.IsDark || wand.Charges != 0 || !player.IdentifyMgr.IsIdentified(wand) {
		t.Fatalf("light-wand result = %+v, dark=%t, charges=%d, identified=%t", result, room.IsDark, wand.Charges, player.IdentifyMgr.IsIdentified(wand))
	}
	for y := range room.Height {
		for x := range room.Width {
			if !level.GetTile(x, y).Explored {
				t.Errorf("light wand did not reveal room tile (%d,%d)", x, y)
			}
		}
	}
	if !level.GetTile(3, 3).HallucinationKnown {
		t.Fatal("lighting a sober room did not remember newly visible stairs")
	}
}

func TestBoltWandHitsFirstMonsterAndStopsAtWall(t *testing.T) {
	player := actor.NewPlayerWithSeed(2, 2, 73)
	level := openTestLevel(9, 5)
	level.FloorNumber = 1
	monster := actor.NewMonsterWithRandAndFloor(5, 2, 'S', 1, player.RandomSource())
	monster.HP, monster.MaxHP = 100, 100
	level.Monsters = []*actor.Monster{monster}
	wand := item.NewItem(0, 0, item.ItemWand, "magic missile", 170)
	wand.Charges = 2

	result := UseWand(wand, player, level, 1, 0)
	damage := 100 - monster.HP
	if !result.Success || !result.Identified || damage < 2 || damage > 5 || wand.Charges != 1 {
		t.Fatalf("magic missile result = %+v, damage=%d, charges=%d; want identified 2..5 damage and one charge consumed", result, damage, wand.Charges)
	}

	level.SetTile(4, 2, dungeon.TileWall)
	result = UseWand(wand, player, level, 1, 0)
	if !result.Success || monster.HP != 100-damage || wand.Charges != 0 {
		t.Errorf("blocked bolt result = %+v, monster HP=%d, charges=%d; want wall to stop damage while consuming charge", result, monster.HP, wand.Charges)
	}
}

func TestMagicMappingRevealsEveryTileAndTracksKnownStairs(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	for _, testCase := range []struct {
		name          string
		hallucinating bool
	}{
		{name: "sober"},
		{name: "hallucinating", hallucinating: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			level := openTestLevel(7, 4)
			level.GetTile(1, 1).Type = dungeon.TileStairsDown
			player := actor.NewPlayerWithSeed(0, 0, 83)
			if testCase.hallucinating {
				player.HallucinationTurns = 5
			}

			result := UseScroll("magic mapping", player, level)
			if !result.Success || !result.Identified {
				t.Fatalf("magic-mapping result = %+v", result)
			}
			for y := range level.Height {
				for x := range level.Width {
					tile := level.GetTile(x, y)
					if !tile.Visible || !tile.Explored {
						t.Errorf("mapped tile (%d,%d) visibility = %t/%t, want visible and explored", x, y, tile.Visible, tile.Explored)
					}
				}
			}
			if got := level.GetTile(1, 1).HallucinationKnown; got == testCase.hallucinating {
				t.Fatalf("stair known state = %t while hallucinating=%t", got, testCase.hallucinating)
			}
		})
	}
}

func openTestLevel(width, height int) *dungeon.Level {
	level := &dungeon.Level{Width: width, Height: height, FloorNumber: 1}
	level.Tiles = make([][]*dungeon.Tile, height)
	for y := range height {
		level.Tiles[y] = make([]*dungeon.Tile, width)
		for x := range width {
			level.Tiles[y][x] = &dungeon.Tile{Type: dungeon.TileFloor, IsWalkable: true}
		}
	}
	return level
}

func TestDrainLifeNoMonsterDoesNotHurtPlayer(t *testing.T) {
	player := actor.NewPlayerWithSeed(2, 2, 97)
	player.HP = 10
	wand := item.NewItem(0, 0, item.ItemWand, "drain life", 300)
	wand.Charges = 1

	result := UseWand(wand, player, openTestLevel(5, 5), 1, 0)
	if !result.Success || player.HP != 10 || wand.Charges != 0 {
		t.Fatalf("empty drain-life result = %+v, HP=%d, charges=%d; want no HP loss but one charge consumed", result, player.HP, wand.Charges)
	}
}

func TestWandHasteAndSlowCancelOppositeStatus(t *testing.T) {
	player := actor.NewPlayerWithSeed(2, 2, 101)
	level := openTestLevel(6, 5)
	monster := actor.NewMonsterWithRandAndFloor(3, 2, 'S', 1, player.RandomSource())
	monster.IsSlowed = true
	level.Monsters = []*actor.Monster{monster}

	haste := item.NewItem(0, 0, item.ItemWand, "haste monster", 5)
	haste.Charges = 1
	if result := UseWand(haste, player, level, 1, 0); !result.Success || monster.IsSlowed || monster.IsHasted {
		t.Fatalf("haste-on-slowed result = %+v, slowed=%t, hasted=%t; want only slow cleared", result, monster.IsSlowed, monster.IsHasted)
	}

	slow := item.NewItem(0, 0, item.ItemWand, "slow monster", 350)
	slow.Charges = 1
	monster.IsHasted = true
	if result := UseWand(slow, player, level, 1, 0); !result.Success || monster.IsHasted || monster.IsSlowed {
		t.Fatalf("slow-on-hasted result = %+v, slowed=%t, hasted=%t; want only haste cleared", result, monster.IsSlowed, monster.IsHasted)
	}
}
