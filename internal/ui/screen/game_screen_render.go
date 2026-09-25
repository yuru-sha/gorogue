package screen

import (
	"fmt"
	"reflect"

	"github.com/anaseto/gruid"
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

	// Draw status lines (top 2 rows)
	s.drawStatusLines(grid)

	s.drawDisplayCells(grid)

	// Draw message log (bottom 7 rows)
	s.drawMessageLog(grid)

	// CLIモードの表示
	if s.inputMode == ModeCLI {
		s.drawCLIPrompt(grid)
	}
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

// drawStatusLines draws the status information at the top
func (s *GameScreen) drawStatusLines(grid *gruid.Grid) {
	currentFloor := 1
	floorInfo := map[string]any{}

	if s.dungeonManager != nil {
		currentFloor = s.dungeonManager.GetCurrentFloor()
		floorInfo = s.dungeonManager.GetFloorInfo()
	}

	// 第1行: プレイヤーステータス
	statusLine1 := fmt.Sprintf(
		"Lv:%d  HP:%d/%d  Atk:%d  Def:%d  Hunger:%d%%  Exp:%d  Gold:%d",
		s.player.Level,
		s.player.HP,
		s.player.MaxHP,
		s.player.Attack,
		s.player.Defense,
		s.player.Hunger,
		s.player.Exp,
		s.player.Gold,
	)
	s.drawText(grid, 0, 0, statusLine1, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})

	// 右上に詳細階層表示を追加
	floorDisplay := s.formatFloorDisplay(currentFloor, floorInfo)
	s.drawText(grid, s.width-len(floorDisplay), 0, floorDisplay, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})

	// 第2行: 装備情報
	s.drawEquipmentLine(grid)
}

// formatFloorDisplay formats the floor display with additional information
func (s *GameScreen) formatFloorDisplay(currentFloor int, floorInfo map[string]any) string {
	baseDisplay := fmt.Sprintf("B%dF/26", currentFloor)

	// 特別な階層の場合はマーカーを追加
	if isSpecial, ok := floorInfo["is_special"].(bool); ok && isSpecial {
		baseDisplay += " [MAZE]"
	}

	// 最終階層の場合
	if isFinal, ok := floorInfo["is_final"].(bool); ok && isFinal {
		baseDisplay += " [FINAL]"
	}

	// 魔除けを持っている場合
	if hasAmulet, ok := floorInfo["player_has_amulet"].(bool); ok && hasAmulet {
		baseDisplay += " [AMULET]"
	}

	// 勝利可能な場合
	if canEscape, ok := floorInfo["can_escape"].(bool); ok && canEscape {
		baseDisplay += " [ESCAPE!]"
	}

	return baseDisplay
}

// drawEquipmentLine draws the equipment status line
func (s *GameScreen) drawEquipmentLine(grid *gruid.Grid) {
	weapon, armor, ringLeft, ringRight := s.player.Equipment.GetEquippedNames()
	statusLine2 := fmt.Sprintf(
		"Weapon: %-15s  Armor: %-15s  Ring: (L): %-10s  Ring: (R): %-10s",
		weapon,
		armor,
		ringLeft,
		ringRight,
	)

	// ウィザードモードの表示を追加
	if s.wizardMode != nil && s.wizardMode.IsActive {
		statusLine2 += "  [WIZARD MODE]"
	}

	s.drawText(grid, 0, 1, statusLine2, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
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
		grid.Set(gruid.Point{X: cell.X, Y: cell.Y + 2}, gruid.Cell{
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

// drawMessageLog draws the message log at the bottom
func (s *GameScreen) drawMessageLog(grid *gruid.Grid) {
	for i, msg := range s.messages {
		s.drawText(grid, 0, s.height-7+i, msg, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
	}
}

// drawCLIPrompt draws the CLI prompt when in CLI mode
func (s *GameScreen) drawCLIPrompt(grid *gruid.Grid) {
	cliPrompt := fmt.Sprintf("CLI> %s_", s.cliBuffer)
	s.drawText(grid, 0, s.height-1, cliPrompt, gruid.Style{Fg: 0x00FF00, Bg: 0x000000}) // 緑色で表示
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
