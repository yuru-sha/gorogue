// GoRogue - SDL2 ASCII文字表示版メイン

package main

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid-sdl"
	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

// ASCIITileManager implements a simple ASCII font-based tile manager
type ASCIITileManager struct {
	cellWidth  int
	cellHeight int
}

// Color constants for consistent display
const (
	BackgroundColor = 0x000000 // Pure black
	TextColor       = 0xFFFFFF // Pure white
)

// NewASCIITileManager creates a new ASCII tile manager
func NewASCIITileManager(cellWidth, cellHeight int) *ASCIITileManager {
	return &ASCIITileManager{
		cellWidth:  cellWidth,
		cellHeight: cellHeight,
	}
}

// TileSize returns the size of tiles
func (tm *ASCIITileManager) TileSize() gruid.Point {
	return gruid.Point{X: tm.cellWidth, Y: tm.cellHeight}
}

// GetImage returns an image for a given cell
func (tm *ASCIITileManager) GetImage(cell gruid.Cell) image.Image {
	// Create a new image for the cell
	img := image.NewRGBA(image.Rect(0, 0, tm.cellWidth, tm.cellHeight))

	// Background color - ensure consistent black
	bgColor := color.RGBA{0, 0, 0, 255} // Black background
	if cell.Style.Bg != 0 {
		bgColor = tm.gruidColorToRGBA(cell.Style.Bg)
	}

	// Fill background - force alpha to 255 for consistent display
	bgColor.A = 255
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Text color - ensure consistent alpha
	textColor := color.RGBA{255, 255, 255, 255} // White text
	if cell.Style.Fg != 0 {
		textColor = tm.gruidColorToRGBA(cell.Style.Fg)
	}
	textColor.A = 255 // Force alpha to 255 for consistent display

	// Draw the character
	if cell.Rune != 0 && cell.Rune != ' ' {
		tm.drawCharacter(img, cell.Rune, textColor)
	}

	return img
}

// gruidColorToRGBA converts gruid.Color to color.RGBA with consistent alpha
func (tm *ASCIITileManager) gruidColorToRGBA(c gruid.Color) color.RGBA {
	r := uint8((c >> 16) & 0xFF)
	g := uint8((c >> 8) & 0xFF)
	b := uint8(c & 0xFF)
	// Force alpha to 255 for consistent display
	return color.RGBA{r, g, b, 255}
}

// drawCharacter draws a character on the image
func (tm *ASCIITileManager) drawCharacter(img *image.RGBA, r rune, textColor color.RGBA) {
	// Use the larger Inconsolata font for better readability
	face := inconsolata.Regular8x16

	// Create a font drawer with adjusted positioning for Inconsolata font
	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{textColor},
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(1 << 6), Y: fixed.Int26_6(tm.cellHeight-2) << 6},
	}

	// Draw the character
	drawer.DrawString(string(r))
}

func main() {
	// ロガーの初期化
	if err := logger.Setup(); err != nil {
		panic(err)
	}
	defer logger.Cleanup()

	// 設定の読み込み確認
	if config.GetDebugMode() {
		logger.Info("Debug mode enabled", "config_loaded", true)
		config.PrintConfig()
	}

	logger.Info("Starting GoRogue", 
		"render_mode", "sdl2_ascii",
		"debug_mode", config.GetDebugMode(),
		"log_level", config.GetLogLevel(),
	)

	// ゲームエンジンの初期化
	engine := core.NewEngine()
	if engine == nil {
		logger.Fatal("Failed to initialize game engine")
		os.Exit(1)
	}

	// SDL2ドライバーの設定 - 固定サイズ
	sdlConfig := sdl.Config{
		TileManager: NewASCIITileManager(10, 16), // 10x16のInconsolataフォント用サイズ
		Width:       80,
		Height:      50,
		WindowTitle: "GoRogue - ASCII Roguelike",
		Fullscreen:  false,
	}

	driver := sdl.NewDriver(sdlConfig)

	// ドライバーの初期化
	if err := driver.Init(); err != nil {
		logger.Fatal("Failed to initialize SDL driver", "error", err.Error())
		os.Exit(1)
	}
	defer driver.Close()

	// アプリケーションの作成と実行
	app := gruid.NewApp(gruid.AppConfig{
		Driver: driver,
		Model:  engine,
	})

	// アプリケーションの実行
	if err := app.Start(context.Background()); err != nil {
		logger.Fatal("Game terminated with error", "error", err.Error())
		os.Exit(1)
	}
}
