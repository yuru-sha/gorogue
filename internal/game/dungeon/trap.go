package dungeon

// TrapType identifies one of Rogue 5.4.4's eight traps.
type TrapType int

const (
	TrapDoor TrapType = iota
	TrapArrow
	TrapSleep
	TrapBear
	TrapTeleport
	TrapDart
	TrapRust
	TrapMystery
)

// Trap stores a hidden trap and whether the player has discovered it.
type Trap struct {
	Type       TrapType
	Position   Position
	Discovered bool
}

// String returns the Rogue name shown when a trap is discovered.
func (t TrapType) String() string {
	switch t {
	case TrapDoor:
		return "trap door"
	case TrapArrow:
		return "arrow trap"
	case TrapSleep:
		return "sleep trap"
	case TrapBear:
		return "bear trap"
	case TrapTeleport:
		return "teleport trap"
	case TrapDart:
		return "dart trap"
	case TrapRust:
		return "rust trap"
	case TrapMystery:
		return "mystery trap"
	default:
		return "unknown trap"
	}
}

// TrapAt returns the trap at a position, if one exists.
func (l *Level) TrapAt(x, y int) *Trap {
	for _, trap := range l.Traps {
		if trap.Position.X == x && trap.Position.Y == y {
			return trap
		}
	}
	return nil
}

// SearchTraps searches the eight surrounding cells using Rogue's odds:
// one chance in 2, worsened by blindness and hallucination.
func (l *Level) SearchTraps(x, y int, blind, hallucinating bool) []*Trap {
	penalty := trapSearchPenalty(blind, hallucinating)
	var found []*Trap
	for searchY := y - 1; searchY <= y+1; searchY++ {
		for searchX := x - 1; searchX <= x+1; searchX++ {
			if searchX == x && searchY == y {
				continue
			}
			trap := l.TrapAt(searchX, searchY)
			if trap == nil || trap.Discovered || l.random().Intn(2+penalty) != 0 {
				continue
			}
			trap.Discovered = true
			found = append(found, trap)
		}
	}
	return found
}

func trapSearchPenalty(blind, hallucinating bool) int {
	penalty := 0
	if hallucinating {
		penalty += 3
	}
	if blind {
		penalty += 2
	}
	return penalty
}
