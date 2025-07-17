package screen

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

var titleArt = []string{
	"",
	"GoRogue",
	"",
}

var menuBox = []string{
	"",
	"NEW GAME",
	"QUIT",
	"",
}

var version = "v0.1.0"

// Colors
var (
	colorYellow   = gruid.Style{Fg: 3}  // 黄色
	colorGray     = gruid.Style{Fg: 8}  // グレー
	colorWhite    = gruid.Style{Fg: 15} // 白
	colorDarkGray = gruid.Style{Fg: 7}  // 暗いグレー
)

// MenuScreen represents the menu screen
type MenuScreen struct {
	width    int
	height   int
	selected int
	grid     gruid.Grid
}

// NewMenuScreen creates a new menu screen
func NewMenuScreen(width, height int) *MenuScreen {
	return &MenuScreen{
		width:    width,
		height:   height,
		selected: 0,
		grid:     gruid.NewGrid(width, height),
	}
}

// HandleInput handles input events
func (s *MenuScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case "Up":
			s.selected = 0
		case "Down":
			s.selected = 1
		case "Enter":
			if s.selected == 0 {
				logger.Info("Game started from menu")
				return state.StateGame
			} else {
				logger.Info("Game quit from menu")
				return state.StateGameOver
			}
		}
	}

	return state.StateMenu
}

// Draw draws the menu screen
func (s *MenuScreen) Draw(grid *gruid.Grid) {
	// グリッドをクリア
	grid.Fill(gruid.Cell{Rune: ' '})

	// タイトルの描画
	titleY := s.height/4 - len(titleArt)/2
	for i, line := range titleArt {
		titleX := (s.width - len(line)) / 2
		s.drawText(grid, titleX, titleY+i, line, colorYellow)
	}

	// メニューの描画
	menuY := titleY + len(titleArt) + 4
	for i, line := range menuBox {
		menuX := (s.width - len(line)) / 2
		style := colorGray

		// 選択中の項目をハイライト
		if (i == 1 && s.selected == 0) || (i == 2 && s.selected == 1) {
			style = colorWhite
		}

		s.drawText(grid, menuX, menuY+i, line, style)
	}

	// バージョン情報の描画
	versionText := "Version " + version
	versionX := 1
	versionY := s.height - 1
	s.drawText(grid, versionX, versionY, versionText, colorDarkGray)

	// 操作説明の描画
	controlsText := "↑↓:Select  Enter:Decide"
	controlsX := (s.width - len(controlsText)) / 2
	controlsY := menuY + len(menuBox) + 2
	s.drawText(grid, controlsX, controlsY, controlsText, colorGray)

	logger.Trace("Menu screen drawn")
}

// drawText draws text at the specified position with the given style
func (s *MenuScreen) drawText(grid *gruid.Grid, x, y int, text string, style gruid.Style) {
	for i, r := range text {
		pos := gruid.Point{X: x + i, Y: y}
		if pos.X >= grid.Size().X {
			break
		}
		grid.Set(pos, gruid.Cell{Rune: r, Style: style})
	}
}
