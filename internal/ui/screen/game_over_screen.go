package screen

import (
	"fmt"
	"time"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/score"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// GameOverScreen represents the game over screen
type GameOverScreen struct {
	width      int
	height     int
	selected   int
	scoreEntry *score.ScoreEntry
	showStats  bool
	menuItems  []string
}

const (
	gameOverKeyDown  = "Down"
	gameOverKeyEnter = "Enter"
	gameOverKeySpace = "Space"
)

// NewGameOverScreen creates a new game over screen
func NewGameOverScreen(width, height int, scoreEntry *score.ScoreEntry) *GameOverScreen {
	menuItems := []string{"R) Restart", "M) Main Menu", "Q) Quit"}
	return &GameOverScreen{
		width:      width,
		height:     height,
		selected:   0,
		scoreEntry: scoreEntry,
		showStats:  false,
		menuItems:  menuItems,
	}
}

// HandleInput handles input events
func (s *GameOverScreen) HandleInput(msg gruid.Msg) state.GameState {
	keyMsg, ok := msg.(gruid.MsgKeyDown)
	if !ok {
		return state.StateGameOver
	}
	switch keyMsg.Key {
	case "Up":
		s.selected = (s.selected - 1 + len(s.menuItems)) % len(s.menuItems)
	case gameOverKeyDown:
		s.selected = (s.selected + 1) % len(s.menuItems)
	case gameOverKeyEnter:
		return s.handleMenuSelection()
	case gameOverKeySpace:
		s.showStats = !s.showStats
	default:
		// キーによる直接選択
		switch keyMsg.Key {
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

	return state.StateGameOver
}

// handleMenuSelection handles menu selection
func (s *GameOverScreen) handleMenuSelection() state.GameState {
	switch s.selected {
	case 0: // Restart
		logger.Info("Restart selected from game over")
		return state.StateGame
	case 1: // Main Menu
		logger.Info("Main Menu selected from game over")
		return state.StateMenu
	case 2: // Quit
		logger.Info("Quit selected from game over")
		return state.StateGameOver
	}
	return state.StateGameOver
}

// Draw draws the game over screen
func (s *GameOverScreen) Draw(grid *gruid.Grid) {
	// グリッドをクリア
	grid.Fill(gruid.Cell{Rune: ' '})

	// PyRogue風のゲームオーバータイトル
	gameOverArt := []string{
		"",
		"  @@@@@@   @@@@@@@@ @@@@@@@@@@@  @@@@@@@@@@@  ",
		" @@@@@@@  @@@@@@@@@ @@@@@@@@@@@@  @@@@@@@@@@@ ",
		"!@@       @@!   @@@ @@! @@! @@!  @@!         ",
		"!@!       !@!   @!@ !@! !@! !@!  !@!         ",
		"!@! @!@!@ @!@!@!@!@ @!! !!@ @!@  @!!!:!     ",
		"!!! !!@! !!!@!!!! !!@   ! !@!  !!!!!:      ",
		":!!   !!  !!:  !!! !!:     !!:  !!:         ",
		":!:   !:  :!:  !:! :!:     :!:  :!:         ",
		" ::: ::::  :: :::  :::     ::   :: ::::     ",
		" :: :: :   :: : :   :      :    : :: ::      ",
		"",
		"   @@@@@@  @@@  @@@  @@@@@@@  @@@@@@@  ",
		"  @@@@@@@ @@@@  @@@ @@@@@@@@  @@@@@@@@@ ",
		"  @@!  @@@ @@!  @@@ @@!       @@!  @@@ ",
		"  !@!  @!@  !@!  !@! !@!       !@!  @!@ ",
		"  @!@  !@!  @!@  !@@ @!!!:!    @!@!!@!  ",
		"  !@!  !!! !!@  !@! !!!!!:    !!@!@!   ",
		"  !!:  !!!  !!:  !!! !!:       !!: :!! ",
		"  :!:  !:!  :!:  !:!  :!:       :!:  !:!",
		"  ::::: ::   :::: ::  :: ::::   ::   ::::",
		"   : :  :    :: :  :   : :: :     :   : :",
		"",
	}

	// ゲームオーバータイトルの描画
	titleY := 2
	for i, line := range gameOverArt {
		if line != "" {
			titleX := (s.width - len(line)) / 2
			if titleX < 0 {
				titleX = 0
			}
			s.drawText(grid, titleX, titleY+i, line, gruid.Style{Fg: 1}) // 赤色
		}
	}

	// スコア情報の描画
	scoreY := titleY + len(gameOverArt) + 2
	if s.scoreEntry != nil {
		scoreLines := []string{
			fmt.Sprintf("Final Score: %d", s.scoreEntry.Score),
			fmt.Sprintf("Level: %d", s.scoreEntry.Level),
			fmt.Sprintf("Deepest Floor: %d", s.scoreEntry.DeepestFloor),
			fmt.Sprintf("Play Time: %s", s.formatPlayTime(s.scoreEntry.PlayTime)),
		}

		if s.scoreEntry.DeathReason != "" {
			scoreLines = append(scoreLines, fmt.Sprintf("Death: %s", s.scoreEntry.DeathReason))
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

			for i, line := range statsLines {
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

	logger.Trace("Game over screen drawn")
}

// formatPlayTime formats play time in seconds to a readable format
func (s *GameOverScreen) formatPlayTime(seconds int64) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

// drawText draws text at the specified position with the given style
func (s *GameOverScreen) drawText(grid *gruid.Grid, x, y int, text string, style gruid.Style) {
	for i, r := range text {
		pos := gruid.Point{X: x + i, Y: y}
		if pos.X >= 0 && pos.X < grid.Size().X && pos.Y >= 0 && pos.Y < grid.Size().Y {
			grid.Set(pos, gruid.Cell{Rune: r, Style: style})
		}
	}
}
