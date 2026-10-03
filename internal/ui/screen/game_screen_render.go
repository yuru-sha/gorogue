package screen

import (
	"fmt"
	"reflect"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// Draw draws the game screen
func (s *GameScreen) Draw(grid *gruid.Grid) {
	// Collect current status information
	currentStats := s.collectCurrentStats()

	// Log output only when status has changed
	if !reflect.DeepEqual(s.lastStats, currentStats) {
		s.logStatsChange()
		s.lastStats = currentStats
	}

	// Output detailed drawing logs at TRACE level
	logger.Trace("Drawing game screen")

	// Clear grid - consistent black background with proper alpha
	blackCell := gruid.Cell{Rune: ' ', Style: gruid.Style{Fg: 0x000000, Bg: 0x000000}}
	grid.Fill(blackCell)
	if s.presentation == presentationHelp && s.helpStage == helpShowingList {
		pageSize := s.helpPageSize()
		start := s.helpPage * pageSize
		end := min(start+pageSize, len(s.messages))
		for i, line := range s.messages[start:end] {
			s.drawText(grid, 0, i, line, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
		}
		footer := "Press any key to return."
		if end < len(s.messages) {
			footer = "Press space for the next page."
		}
		s.drawText(grid, 0, s.height-1, footer, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
		return
	}

	if s.presentation == presentationDeath || s.presentation == presentationVictory {
		if s.sequenceStage < len(s.sequencePages) {
			for i, line := range s.sequencePages[s.sequenceStage] {
				s.drawText(grid, 0, i, line, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
			}
			s.drawText(grid, 0, len(s.sequencePages[s.sequenceStage])+1, "Press any key to continue.", gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
		}
		return
	}
	s.drawText(grid, 0, 0, s.lastMessage(), gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
	s.drawDisplayCells(grid)
	s.drawStatusLines(grid)
	if s.inputMode == ModeCLI {
		s.drawCLIPrompt(grid)
	}

}
func (s *GameScreen) lastMessage() string {
	if len(s.messages) == 0 {
		return ""
	}
	return s.messages[len(s.messages)-1]
}

// collectCurrentStats collects current player stats for change detection
func (s *GameScreen) collectCurrentStats() map[string]any {
	return map[string]any{
		"level":   s.player.Level,
		"hp":      s.player.HP,
		"max_hp":  s.player.MaxHP,
		"attack":  s.player.Attack,
		"defense": s.player.Defense,
		"hunger":  s.player.Hunger,
		"exp":     s.player.Exp,
		"gold":    s.player.Gold,
	}
}

// logStatsChange logs when player stats change
func (s *GameScreen) logStatsChange() {
	logger.Debug("Player stats changed",
		"level", s.player.Level,
		"hp", s.player.HP,
		"max_hp", s.player.MaxHP,
		"attack", s.player.Attack,
		"defense", s.player.Defense,
		"hunger", s.player.Hunger,
		"exp", s.player.Exp,
		"gold", s.player.Gold,
	)
}

func (s *GameScreen) drawStatusLines(grid *gruid.Grid) {
	if s.player == nil || s.height < 2 {
		return
	}
	floor := 1
	if s.dungeonManager != nil {
		floor = s.dungeonManager.GetCurrentFloor()
	}
	status := fmt.Sprintf("Level: %d Gold: %d Hp: %d(%d) Str: %d(%d) Arm: %d Exp: %d/%d%s",
		floor, s.player.Gold, s.player.HP, s.player.MaxHP, s.player.Strength, s.player.MaxStrength,
		10-s.player.GetTotalDefense(), s.player.Level, s.player.Exp, hungerLabel(s.player.HungerState))
	s.drawText(grid, 0, s.height-1, status, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
}

func hungerLabel(hunger int) string {
	switch hunger {
	case actor.HungerHungry:
		return " Hungry"
	case actor.HungerWeak:
		return " Weak"
	case actor.HungerFainting:
		return " Faint"
	default:
		return ""
	}
}

// drawDisplayCells converts game state into logical display cells, then maps
// those cells to terminal glyphs, colors, and gruid cells.
func (s *GameScreen) drawDisplayCells(grid *gruid.Grid) {
	if s.level == nil {
		return
	}
	s.displayCells = convertDisplayCells(s.displayCells, s.level, s.player)
	for _, cell := range s.displayCells {
		if !cell.Visible && !cell.Explored {
			continue
		}
		glyph, color := terrainAppearance(cell.Terrain)
		switch cell.Entity {
		case displayEntityItem:
			glyph, color = itemAppearance(cell.ItemType)
		case displayEntityTrap:
			glyph, color = '^', 0xFFFF00
		case displayEntityMonster:
			glyph, color = cell.MonsterCode, 0xFFFFFF
		case displayEntityPlayer:
			glyph, color = '@', 0xFFFFFF
		}
		if cell.Hallucinated {
			glyph = hallucinationGlyph(cell.X, cell.Y, s.player.HallucinationTurns, cell.Entity == displayEntityMonster, s.level.FloorNumber)
		}
		grid.Set(gruid.Point{X: cell.X, Y: cell.Y + 1}, gruid.Cell{
			Rune:  glyph,
			Style: gruid.Style{Fg: color, Bg: 0x000000},
		})
	}
}

func hallucinationGlyph(x, y, turns int, monster bool, floorNumber int) rune {
	value := int64(x+1)*6364136223846793005 ^ int64(y+1)*1442695040888963407 ^ int64(turns)*3202034522624059733
	value ^= value >> 29
	value *= 3037000493
	value ^= value >> 31
	value &= 0x7fffffffffffffff
	if monster {
		return rune('A' + int(value%26))
	}
	const objectGlyphs = "!?=/:)]%*,"
	choices := len(objectGlyphs)
	if floorNumber < 26 {
		choices--
	}
	return rune(objectGlyphs[int(value%int64(choices))])
}

func terrainAppearance(tileType dungeon.TileType) (rune, gruid.Color) {
	switch tileType {
	case dungeon.TileWall, dungeon.TileSecretDoor, dungeon.TileSecretPassage:
		return '#', 0x826E32
	case dungeon.TileFloor:
		return '.', 0x808080
	case dungeon.TileDoor, dungeon.TileDoorClosed:
		return '+', 0x8B4513
	case dungeon.TileDoorOpen, dungeon.TileOpenDoor:
		return '/', 0x8B4513
	case dungeon.TileStairsUp:
		return '<', 0xFFFFFF
	case dungeon.TileStairsDown:
		return '>', 0xFFFFFF
	case dungeon.TileWater:
		return '~', 0x00FFFF
	case dungeon.TileLava:
		return '^', 0xFF0000
	case dungeon.TilePassage:
		return '#', 0x808080
	default:
		return ' ', 0xFFFFFF
	}
}

func itemAppearance(itemType gameitem.ItemType) (rune, gruid.Color) {
	switch itemType {
	case gameitem.ItemWeapon:
		return ')', 0xC0C0C0
	case gameitem.ItemArmor:
		return ']', 0x8B4513
	case gameitem.ItemRing:
		return '=', 0xFFD700
	case gameitem.ItemScroll:
		return '?', 0xFFFFFF
	case gameitem.ItemPotion:
		return '!', 0xFF1493
	case gameitem.ItemWand:
		return '/', 0x8A2BE2
	case gameitem.ItemFood:
		return ':', 0xFFA500
	case gameitem.ItemGold:
		return '*', 0xFFD700
	case gameitem.ItemAmulet:
		return ',', 0x9400D3
	default:
		return '*', 0xFFFFFF
	}
}

// drawCLIPrompt draws the CLI prompt in the message row.
func (s *GameScreen) drawCLIPrompt(grid *gruid.Grid) {
	s.drawText(grid, 0, 0, fmt.Sprintf("CLI> %s_", s.cliBuffer), gruid.Style{Fg: 0x00FF00, Bg: 0x000000})
}

// drawText draws text at the specified position with the given style
func (s *GameScreen) drawText(grid *gruid.Grid, x, y int, text string, style gruid.Style) {
	for i, r := range text {
		pos := gruid.Point{X: x + i, Y: y}
		if pos.X >= grid.Size().X {
			break
		}
		grid.Set(pos, gruid.Cell{Rune: r, Style: style})
	}
}
