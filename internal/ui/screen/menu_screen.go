package screen

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// PyRogue準拠のタイトルアート（純ASCII文字）
var titleArt = []string{
	"",
	"  ####   ####  ####   ####   ####  #   # #####",
	" #      #    # #   # #    # #    # #   # #    ",
	" #  ### #    # ####  #    # #  ### #   # ####",
	" #    # #    # #   # #    # #    # #   # #    ",
	"  ####   ####  #   #  ####   ####   ###  #####",
	"",
	" A Go Roguelike Adventure",
	"",
}

// PyRogue準拠のシンプルなメニュー
var menuBox = []string{
	"",
	"New Game",
	"Help",
	"Quit",
	"",
}

var version = "v0.3.0"

// Colors (SDL2対応の16進数カラー)
var (
	colorYellow   = gruid.Style{Fg: 0xFFFF00} // 黄色
	colorGray     = gruid.Style{Fg: 0x808080} // グレー
	colorWhite    = gruid.Style{Fg: 0xFFFFFF} // 白
	colorDarkGray = gruid.Style{Fg: 0x404040} // 暗いグレー
)

// MenuScreen represents the menu screen
type MenuScreen struct {
	width       int
	height      int
	selected    int
	grid        gruid.Grid
	menuItems   []string
	saveManager *save.SaveManager
}

// NewMenuScreen creates a new menu screen
func NewMenuScreen(width, height int) *MenuScreen {
	// セーブマネージャーを初期化
	saveManager := save.NewSaveManager()
	saveManager.Initialize()

	// セーブデータの存在をチェック
	menuItems := []string{"New Game"}
	if saveManager.FileExists() {
		menuItems = append(menuItems, "Load Game")
	}
	menuItems = append(menuItems, "Help", "Quit")

	return &MenuScreen{
		width:       width,
		height:      height,
		selected:    0,
		grid:        gruid.NewGrid(width, height),
		menuItems:   menuItems,
		saveManager: saveManager,
	}
}

// HandleInput handles input events
func (s *MenuScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		logger.Debug("MenuScreen key pressed", "key", msg.Key, "mod", msg.Mod)
		switch msg.Key {
		case gruid.KeyArrowUp:
			s.selected = (s.selected - 1 + len(s.menuItems)) % len(s.menuItems)
			logger.Debug("Selected item changed", "selected", s.selected)
		case gruid.KeyArrowDown:
			s.selected = (s.selected + 1) % len(s.menuItems)
			logger.Debug("Selected item changed", "selected", s.selected)
		case gruid.KeyEnter:
			return s.handleMenuSelection()
		default:
			// キーによる直接選択（文字列での判定）
			keyStr := string(msg.Key)
			switch keyStr {
			case "n", "N":
				s.selected = 0
				return s.handleMenuSelection()
			case "l", "L":
				// Load Gameがメニューにあるかチェック
				for i, item := range s.menuItems {
					if item == "Load Game" {
						s.selected = i
						return s.handleMenuSelection()
					}
				}
			case "h", "H":
				// Helpの位置を動的に検索
				for i, item := range s.menuItems {
					if item == "Help" {
						s.selected = i
						return s.handleMenuSelection()
					}
				}
			case "q", "Q":
				// Quitの位置を動的に検索
				for i, item := range s.menuItems {
					if item == "Quit" {
						s.selected = i
						return s.handleMenuSelection()
					}
				}
			}
		}
	}

	return state.StateMenu
}

// handleMenuSelection handles menu selection
func (s *MenuScreen) handleMenuSelection() state.GameState {
	if s.selected < 0 || s.selected >= len(s.menuItems) {
		return state.StateMenu
	}

	selectedItem := s.menuItems[s.selected]
	switch selectedItem {
	case "New Game":
		logger.Info("New Game selected from menu")
		return state.StateGame
	case "Load Game":
		logger.Info("Load Game selected from menu")
		return state.StateSaveLoad
	case "Help":
		logger.Info("Help selected from menu")
		return state.StateHelp
	case "Quit":
		logger.Info("Quit selected from menu")
		return state.StateGameOver
	}
	return state.StateMenu
}

// Draw draws the menu screen
func (s *MenuScreen) Draw(grid *gruid.Grid) {
	// グリッドをクリア（背景色を明示的に設定）
	grid.Fill(gruid.Cell{Rune: ' ', Style: gruid.Style{Bg: 0x000000, Fg: 0xFFFFFF}})

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
	versionText := "GoRogue " + version
	versionX := 1
	versionY := s.height - 1
	s.drawText(grid, versionX, versionY, versionText, colorDarkGray)

	// PyRogue風の操作説明（下部）
	controlsText := "Use UP/DOWN arrows to navigate, ENTER to select, ESC to quit"
	controlsX := (s.width - len(controlsText)) / 2
	if controlsX < 0 {
		controlsX = 0
	}
	controlsY := s.height - 2
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
