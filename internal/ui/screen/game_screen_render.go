package screen

import (
	"fmt"
	"reflect"

	"github.com/anaseto/gruid"
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

	// Draw dungeon
	s.drawDungeon(grid)

	// Draw entities (items, monsters, player)
	s.drawEntities(grid)

	// Draw message log (bottom 7 rows)
	s.drawMessageLog(grid)

	// CLIモードの表示
	if s.inputMode == ModeCLI {
		s.drawCLIPrompt(grid)
	}
}

// collectCurrentStats collects current player stats for change detection
func (s *GameScreen) collectCurrentStats() map[string]interface{} {
	return map[string]interface{}{
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
	if s.dungeonManager != nil {
		currentFloor = s.dungeonManager.GetCurrentFloor()
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

	// 右上に階層表示を追加
	floorDisplay := fmt.Sprintf("B%dF", currentFloor)
	s.drawText(grid, s.width-len(floorDisplay), 0, floorDisplay, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})

	// 第2行: 装備情報
	s.drawEquipmentLine(grid)
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

// drawDungeon draws the dungeon tiles
func (s *GameScreen) drawDungeon(grid *gruid.Grid) {
	for y := 0; y < s.level.Height; y++ {
		for x := 0; x < s.level.Width; x++ {
			tile := s.level.GetTile(x, y)
			grid.Set(gruid.Point{X: x, Y: y + 2}, gruid.Cell{
				Rune:  tile.Rune,
				Style: gruid.Style{Fg: tile.Color, Bg: 0x000000},
			})
		}
	}
}

// drawEntities draws all entities (items, monsters, player)
func (s *GameScreen) drawEntities(grid *gruid.Grid) {
	// アイテムの描画（最初に描画）
	for _, item := range s.level.Items {
		grid.Set(gruid.Point{X: item.Position.X, Y: item.Position.Y + 2}, gruid.Cell{
			Rune:  item.Symbol,
			Style: gruid.Style{Fg: item.Color, Bg: 0x000000},
		})
	}

	// モンスターの描画（アイテムの上に描画）
	for _, monster := range s.level.Monsters {
		if monster.IsAlive() {
			grid.Set(gruid.Point{X: monster.Position.X, Y: monster.Position.Y + 2}, gruid.Cell{
				Rune:  monster.Type.Symbol,
				Style: gruid.Style{Fg: monster.Color, Bg: 0x000000},
			})
		}
	}

	// プレイヤーの描画（最上位に描画）
	grid.Set(gruid.Point{X: s.player.Position.X, Y: s.player.Position.Y + 2}, gruid.Cell{
		Rune:  s.player.Symbol,
		Style: gruid.Style{Fg: s.player.Color, Bg: 0x000000},
	})
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
