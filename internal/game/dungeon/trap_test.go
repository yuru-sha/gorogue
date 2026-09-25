package dungeon

import (
	"math/rand"
	"testing"
)

func TestSearchTrapsFindsAdjacentTrapButNotPlayerTile(t *testing.T) {
	seed := seedWithFirstSearchHit(t)
	level := &Level{
		Seed: seed,
		Traps: []*Trap{
			{Type: TrapDoor, Position: Position{X: 2, Y: 2}},
			{Type: TrapArrow, Position: Position{X: 1, Y: 1}},
		},
	}
	level.rngSource = newTrackedRandomSource(seed)
	level.rng = level.rngSource.rand()

	found := level.SearchTraps(2, 2, false, false)
	if len(found) != 1 || found[0].Type != TrapArrow || !found[0].Discovered {
		t.Fatalf("SearchTraps() = %#v, want only the discovered adjacent arrow trap", found)
	}
	if level.Traps[0].Discovered {
		t.Fatal("search discovered a trap under the player")
	}

	draws := level.RandomDraws()
	if found := level.SearchTraps(2, 2, false, false); len(found) != 0 {
		t.Fatalf("repeat search found already discovered trap: %#v", found)
	}
	if got := level.RandomDraws(); got != draws {
		t.Fatalf("repeat search consumed %d additional random draws", got-draws)
	}
}

func TestSearchTrapsUsesRogueStatusPenalties(t *testing.T) {
	for _, test := range []struct {
		name          string
		blind         bool
		hallucinating bool
		denominator   int
		found         bool
	}{
		{name: "clear sight", denominator: 2, found: true},
		{name: "blind", blind: true, denominator: 4, found: false},
		{name: "hallucinating", hallucinating: true, denominator: 5, found: false},
		{name: "blind and hallucinating", blind: true, hallucinating: true, denominator: 7, found: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			seed := seedForSearchOutcome(t, test.denominator, test.found)
			level := &Level{
				Seed:  seed,
				Traps: []*Trap{{Type: TrapDart, Position: Position{X: 1, Y: 1}}},
			}
			level.rngSource = newTrackedRandomSource(seed)
			level.rng = level.rngSource.rand()

			found := level.SearchTraps(2, 2, test.blind, test.hallucinating)
			if got := len(found) > 0; got != test.found {
				t.Fatalf("SearchTraps() found trap = %t, want %t", got, test.found)
			}
		})
	}
}

func seedForSearchOutcome(t *testing.T, denominator int, found bool) int64 {
	t.Helper()
	for seed := int64(0); ; seed++ {
		got := rand.New(rand.NewSource(seed)).Intn(denominator) == 0
		if got == found {
			return seed
		}
	}
}

func seedWithFirstSearchHit(t *testing.T) int64 {
	t.Helper()
	for seed := int64(0); ; seed++ {
		if rand.New(rand.NewSource(seed)).Intn(2) == 0 {
			return seed
		}
	}
}

func TestRogueTrapTypesHaveStableNames(t *testing.T) {
	for _, test := range []struct {
		trap TrapType
		want string
	}{
		{TrapDoor, "trap door"},
		{TrapArrow, "arrow trap"},
		{TrapSleep, "sleep trap"},
		{TrapBear, "bear trap"},
		{TrapTeleport, "teleport trap"},
		{TrapDart, "dart trap"},
		{TrapRust, "rust trap"},
		{TrapMystery, "mystery trap"},
	} {
		if got := test.trap.String(); got != test.want {
			t.Errorf("TrapType(%d).String() = %q, want %q", test.trap, got, test.want)
		}
	}
}
