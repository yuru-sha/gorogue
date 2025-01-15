package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
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

// MenuScreen represents the menu screen
type MenuScreen struct {
	width    int
	height   int
	selected int
}

// NewMenuScreen creates a new menu screen
func NewMenuScreen(width, height int) *MenuScreen {
	return &MenuScreen{
		width:    width,
		height:   height,
		selected: 0,
	}
}

// Update updates the menu screen state
func (s *MenuScreen) Update() state.GameState {
	// 上下キーでメニュー項目の選択
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		s.selected = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		s.selected = 1
	}

	// Enterキーで選択項目の実行
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if s.selected == 0 {
			logger.Info("Game started from menu")
			return state.StateGame
		} else {
			logger.Info("Game quit from menu")
			return state.StateGameOver
		}
	}

	return state.StateMenu
}

// Draw draws the menu screen
func (s *MenuScreen) Draw(screen *ebiten.Image) {
	// 背景を黒で塗りつぶす
	screen.Fill(color.Black)

	// タイトルの描画
	titleY := s.height/4 - len(titleArt)/2
	for i, line := range titleArt {
		titleX := (s.width*24 - len(line)*12) / 2
		text.Draw(screen, line, GetFont(), titleX, (titleY+i)*24, color.RGBA{R: 255, G: 215, B: 0, A: 255}) // ゴールド
	}

	// メニューの描画
	menuY := titleY + len(titleArt) + 4
	for i, line := range menuBox {
		menuX := (s.width*24 - len(line)*24) / 2                // 文字サイズを2倍に
		textColor := color.RGBA{R: 128, G: 128, B: 128, A: 255} // デフォルトはグレー

		// 選択中の項目をハイライト
		if (i == 1 && s.selected == 0) || (i == 2 && s.selected == 1) {
			textColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // 選択中は白
		}

		text.Draw(screen, line, GetFont(), menuX, (menuY+i)*24, textColor)
	}

	// バージョン情報の描画
	versionText := "Version " + version
	versionX := 10
	versionY := s.height*24 - 10
	text.Draw(screen, versionText, GetFont(), versionX, versionY, color.RGBA{R: 64, G: 64, B: 64, A: 255})

	// 操作説明の描画
	controlsText := "↑↓:Select  Enter:Decide"
	controlsX := (s.width*24 - len(controlsText)*12) / 2
	controlsY := menuY*24 + len(menuBox)*24 + 24
	text.Draw(screen, controlsText, GetFont(), controlsX, controlsY, color.RGBA{R: 128, G: 128, B: 128, A: 255})

	logger.Trace("Menu screen drawn")
}
