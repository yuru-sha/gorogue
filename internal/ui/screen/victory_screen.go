package screen

import (
	"fmt"
	"time"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/score"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// VictoryScreen represents the victory screen
type VictoryScreen struct {
	width      int
	height     int
	selected   int
	scoreEntry *score.ScoreEntry
	showStats  bool
	menuItems  []string
}

// NewVictoryScreen creates a new victory screen
func NewVictoryScreen(width, height int, scoreEntry *score.ScoreEntry) *VictoryScreen {
	menuItems := []string{"R) Restart", "M) Main Menu", "Q) Quit"}
	return &VictoryScreen{
		width:      width,
		height:     height,
		selected:   0,
		scoreEntry: scoreEntry,
		showStats:  false,
		menuItems:  menuItems,
	}
}

// HandleInput handles input events
func (s *VictoryScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case "Up":
			s.selected = (s.selected - 1 + len(s.menuItems)) % len(s.menuItems)
		case "Down":
			s.selected = (s.selected + 1) % len(s.menuItems)
		case "Enter":
			return s.handleMenuSelection()
		case "Space":
			s.showStats = !s.showStats
		default:
			// キーによる直接選択
			switch msg.Key {
			case "r", "R":
				s.selected = 0
				return s.handleMenuSelection()
			case "m", "M":
				s.selected = 1
				return s.handleMenuSelection()
			case "q", "Q":
				s.selected = 2
				return s.handleMenuSelection()
			}
		}
	}

	return state.StateVictory
}

// handleMenuSelection handles menu selection
func (s *VictoryScreen) handleMenuSelection() state.GameState {
	switch s.selected {
	case 0: // Restart
		logger.Info("Restart selected from victory")
		return state.StateGame
	case 1: // Main Menu
		logger.Info("Main Menu selected from victory")
		return state.StateMenu
	case 2: // Quit
		logger.Info("Quit selected from victory")
		return state.StateGameOver
	}
	return state.StateVictory
}

// Draw draws the victory screen
func (s *VictoryScreen) Draw(grid *gruid.Grid) {
	// グリッドをクリア
	grid.Fill(gruid.Cell{Rune: ' '})

	// PyRogue風の勝利タイトル
	victoryArt := []string{
		"",
		"@@@  @@@  @@@  @@@@@@@  @@@@@@@  @@@@@@  @@@@@@@  @@@  @@@",
		"@@@  @@@  @@@  @@@@@@@@  @@@@@@@  @@@@@@@  @@@@@@@  @@@  @@@",
		"@@!  @@@  @@!  @@!       @@!  @@@ @@!  @@@ @@!  @@@ @@!  @@@",
		"!@!  @!@  !@!  !@!       !@!  @!@ !@!  @!@ !@!  @!@ !@!  @!@",
		"@!@  !@!  !!@  @!!!:!    @!@  !@! @!@  !@! @!@!!@!  @!@  !@!",
		"!@!  !!!  !!!  !!!!!:    !@!  !!! !@!  !!! !!@!@!   !@!  !!!",
		"!!:  !!!  !!:  !!:       !!:  !!! !!:  !!! !!: :!!  !!:  !!!",
		":!:  !:!  :!:  :!:       :!:  !:! :!:  !:! :!:  !:!  :!:  !:!",
		"::::: ::   ::   :: ::::   :::: ::  ::::: ::  ::   ::   :::: ::",
		" : :  :    :    : :: :     :: :     : :  :   :    :     :: : :",
		"",
		"  @@@@@@   @@@@@@   @@@@@@@   @@@@@@@   @@@@@@@   @@@@@@@   @@@@@@@",
		" @@@@@@@@  @@@@@@@  @@@@@@@@  @@@@@@@@  @@@@@@@@  @@@@@@@@  @@@@@@@@",
		" @@!  @@@  @@!  @@@ @@!  @@@  @@!  @@@  @@!  @@@  @@!  @@@  @@!",
		" !@!  @!@  !@!  @!@ !@!  @!@  !@!  @!@  !@!  @!@  !@!  @!@  !@!",
		" @!@  !@!  @!@  !@! @!@  !@!  @!@  !@!  @!@  !@!  @!@  !@!  @!!!:!",
		" !@!  !!!  !@!  !!! !@!  !!!  !@!  !!!  !@!  !!!  !@!  !!!  !!!!!:",
		" !!:  !!!  !!:  !!! !!:  !!!  !!:  !!!  !!:  !!!  !!:  !!!  !!:",
		" :!:  !:!  :!:  !:! :!:  !:!  :!:  !:!  :!:  !:!  :!:  !:!  :!:",
		" ::::: ::   :::: ::  ::::: ::   :::: ::   :::: ::   :::: ::   :: ::::",
		"  : :  :    :: :  :   : :  :     :: :     :: :     :: :  :   : :: ::",
		"",
	}

	// 勝利タイトルの描画
	titleY := 1
	for i, line := range victoryArt {
		if len(line) > 0 {
			titleX := (s.width - len(line)) / 2
			if titleX < 0 {
				titleX = 0
			}
			s.drawText(grid, titleX, titleY+i, line, gruid.Style{Fg: 2}) // 緑色
		}
	}

	// 勝利メッセージ
	victoryMsg := "Congratulations! You have retrieved the Amulet of Yendor!"
	msgX := (s.width - len(victoryMsg)) / 2
	if msgX < 0 {
		msgX = 0
	}
	msgY := titleY + len(victoryArt) + 1
	s.drawText(grid, msgX, msgY, victoryMsg, gruid.Style{Fg: 11}) // 明るい黄色

	// スコア情報の描画
	scoreY := msgY + 3
	if s.scoreEntry != nil {
		scoreLines := []string{
			fmt.Sprintf("Final Score: %d", s.scoreEntry.Score),
			fmt.Sprintf("Level: %d", s.scoreEntry.Level),
			fmt.Sprintf("Play Time: %s", s.formatPlayTime(s.scoreEntry.PlayTime)),
		}

		for i, line := range scoreLines {
			scoreX := (s.width - len(line)) / 2
			if scoreX < 0 {
				scoreX = 0
			}
			s.drawText(grid, scoreX, scoreY+i, line, colorWhite)
		}

		// 詳細統計の表示
		if s.showStats {
			statsY := scoreY + len(scoreLines) + 2
			statsLines := []string{
				fmt.Sprintf("Turn Count: %d", s.scoreEntry.TurnCount),
				fmt.Sprintf("Monsters Killed: %d", s.scoreEntry.MonstersKilled),
				fmt.Sprintf("Gold Collected: %d", s.scoreEntry.GoldCollected),
				fmt.Sprintf("Game Seed: %d", s.scoreEntry.GameSeed),
			}

			// 勝利ボーナスの表示
			bonusLines := []string{
				"",
				"=== VICTORY BONUS ===",
				fmt.Sprintf("Survival Bonus: %d", s.calculateSurvivalBonus()),
				fmt.Sprintf("Speed Bonus: %d", s.calculateSpeedBonus()),
				fmt.Sprintf("Exploration Bonus: %d", s.calculateExplorationBonus()),
			}

			allStats := append(statsLines, bonusLines...)
			for i, line := range allStats {
				statsX := (s.width - len(line)) / 2
				if statsX < 0 {
					statsX = 0
				}
				s.drawText(grid, statsX, statsY+i, line, colorGray)
			}
		}
	}

	// メニューの描画
	menuY := s.height - 8
	for i, item := range s.menuItems {
		menuX := (s.width - len(item)) / 2
		style := colorGray

		// 選択中の項目をハイライト
		if i == s.selected {
			style = colorWhite
			// 選択中の項目には矢印を表示
			arrowX := menuX - 3
			s.drawText(grid, arrowX, menuY+i, ">>", colorWhite)
		}

		s.drawText(grid, menuX, menuY+i, item, style)
	}

	// 操作説明の描画
	controlsText := "↑↓:Select  Enter:Decide  Space:Stats"
	controlsX := (s.width - len(controlsText)) / 2
	if controlsX < 0 {
		controlsX = 0
	}
	controlsY := menuY + len(s.menuItems) + 2
	s.drawText(grid, controlsX, controlsY, controlsText, colorGray)

	logger.Trace("Victory screen drawn")
}

// formatPlayTime formats play time in seconds to a readable format
func (s *VictoryScreen) formatPlayTime(seconds int64) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

// calculateSurvivalBonus calculates bonus points for surviving
func (s *VictoryScreen) calculateSurvivalBonus() int {
	if s.scoreEntry == nil {
		return 0
	}
	return s.scoreEntry.Level * 100 // レベルに応じたボーナス
}

// calculateSpeedBonus calculates bonus points for completion speed
func (s *VictoryScreen) calculateSpeedBonus() int {
	if s.scoreEntry == nil {
		return 0
	}
	// 早くクリアするほど高いボーナス
	maxBonus := 10000
	penaltyPerHour := 1000
	hours := s.scoreEntry.PlayTime / 3600
	bonus := maxBonus - int(hours)*penaltyPerHour
	if bonus < 0 {
		bonus = 0
	}
	return bonus
}

// calculateExplorationBonus calculates bonus points for exploration
func (s *VictoryScreen) calculateExplorationBonus() int {
	if s.scoreEntry == nil {
		return 0
	}
	// 深い階層まで到達したボーナス
	return s.scoreEntry.DeepestFloor * 50
}

// drawText draws text at the specified position with the given style
func (s *VictoryScreen) drawText(grid *gruid.Grid, x, y int, text string, style gruid.Style) {
	for i, r := range text {
		pos := gruid.Point{X: x + i, Y: y}
		if pos.X >= 0 && pos.X < grid.Size().X && pos.Y >= 0 && pos.Y < grid.Size().Y {
			grid.Set(pos, gruid.Cell{Rune: r, Style: style})
		}
	}
}
