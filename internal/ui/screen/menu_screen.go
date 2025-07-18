package screen

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// PyRogue準拠のタイトルアート
var titleArt = []string{
	"",
	"  @@@@@@   @@@@@@  @@@@@@@   @@@@@@   @@@@@@  @@@  @@@  @@@@@@@",
	" @@@@@@@  @@@@@@@ @@@@@@@@ @@@@@@@@  @@@@@@@  @@@  @@@  @@@@@@@@",
	" !@@      @@!  @@ @@!  @@@  @@!  @@@  @@!  @@@  @@!  @@@  @@!",
	" !@!      !@!  @!@ !@!  @!@  !@!  @!@  !@!  @!@  !@!  @!@  !@!",
	" !@! @!@!@ @!@  !@! @!@!!@!  @!@  !@!  @!@  !@!  @!@  !@!  @!@!!@!",
	" !!! !!@! !@!  !!! !!@!@!   !@!  !!!  !@!  !!!  !@!  !!!  !!@!@!",
	" :!!   !:  !!:  !!! !!: :!!  !!:  !!!  !!:  !!!  !!:  !!!  !!: :!!",
	" :!:   !:  :!:  !:!  :!:  :!:  :!:  :!:  :!:  :!:  :!:  :!:  :!:  :!:",
	"  ::: ::::  ::::: ::  ::   :::  ::::: ::  ::::: ::  ::::: ::  ::   :::",
	"  :: :: :    : :  :   :    :    : :  :   : :  :   : :  :   :    :",
	"",
}

var menuBox = []string{
	"",
	"N) New Game",
	"L) Load Game",
	"S) Scores",
	"H) Help",
	"Q) Quit",
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
	menuItems []string
}

// NewMenuScreen creates a new menu screen
func NewMenuScreen(width, height int) *MenuScreen {
	menuItems := []string{"N) New Game", "L) Load Game", "S) Scores", "H) Help", "Q) Quit"}
	return &MenuScreen{
		width:     width,
		height:    height,
		selected:  0,
		grid:      gruid.NewGrid(width, height),
		menuItems: menuItems,
	}
}

// HandleInput handles input events
func (s *MenuScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case "Up":
			s.selected = (s.selected - 1 + len(s.menuItems)) % len(s.menuItems)
		case "Down":
			s.selected = (s.selected + 1) % len(s.menuItems)
		case "Enter":
			return s.handleMenuSelection()
		default:
			// キーによる直接選択
			switch msg.Key {
			case "n", "N":
				s.selected = 0
				return s.handleMenuSelection()
			case "l", "L":
				s.selected = 1
				return s.handleMenuSelection()
			case "s", "S":
				s.selected = 2
				return s.handleMenuSelection()
			case "h", "H":
				s.selected = 3
				return s.handleMenuSelection()
			case "q", "Q":
				s.selected = 4
				return s.handleMenuSelection()
			}
		}
	}

	return state.StateMenu
}

// handleMenuSelection handles menu selection
func (s *MenuScreen) handleMenuSelection() state.GameState {
	switch s.selected {
	case 0: // New Game
		logger.Info("New Game selected from menu")
		return state.StateGame
	case 1: // Load Game
		logger.Info("Load Game selected from menu")
		return state.StateSaveLoad
	case 2: // Scores
		logger.Info("Scores selected from menu")
		// TODO: Implement scores screen
		return state.StateMenu
	case 3: // Help
		logger.Info("Help selected from menu")
		return state.StateHelp
	case 4: // Quit
		logger.Info("Quit selected from menu")
		return state.StateGameOver
	}
	return state.StateMenu
}

// Draw draws the menu screen
func (s *MenuScreen) Draw(grid *gruid.Grid) {
	// グリッドをクリア
	grid.Fill(gruid.Cell{Rune: ' '})

	// タイトルの描画
	titleY := 2
	for i, line := range titleArt {
		if len(line) > 0 { // 空行以外のみ描画
			titleX := (s.width - len(line)) / 2
			if titleX < 0 {
				titleX = 0
			}
			s.drawText(grid, titleX, titleY+i, line, colorYellow)
		}
	}

	// メニューの描画
	menuY := titleY + len(titleArt) + 2
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

	// バージョン情報の描画
	versionText := "Version " + version
	versionX := 1
	versionY := s.height - 1
	s.drawText(grid, versionX, versionY, versionText, colorDarkGray)

	// 操作説明の描画
	controlsText := "↑↓:Select  Enter:Decide  or Press Key"
	controlsX := (s.width - len(controlsText)) / 2
	if controlsX < 0 {
		controlsX = 0
	}
	controlsY := menuY + len(s.menuItems) + 2
	s.drawText(grid, controlsX, controlsY, controlsText, colorGray)

	// PyRogue風のクレジット表示
	creditText := "A faithful recreation of the classic Rogue"
	creditX := (s.width - len(creditText)) / 2
	if creditX < 0 {
		creditX = 0
	}
	creditY := controlsY + 2
	s.drawText(grid, creditX, creditY, creditText, colorDarkGray)

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
