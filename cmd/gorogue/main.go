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
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// ASCIITileManager implements a simple ASCII font-based tile manager
type ASCIITileManager struct {
	cellWidth  int
	cellHeight int
}

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

	// Background color
	bgColor := color.RGBA{0, 0, 0, 255} // Black background
	if cell.Style.Bg != 0 {
		bgColor = tm.gruidColorToRGBA(cell.Style.Bg)
	}

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Text color
	textColor := color.RGBA{255, 255, 255, 255} // White text
	if cell.Style.Fg != 0 {
		textColor = tm.gruidColorToRGBA(cell.Style.Fg)
	}

	// Draw the character
	if cell.Rune != 0 && cell.Rune != ' ' {
		tm.drawCharacter(img, cell.Rune, textColor)
	}

	return img
}

// gruidColorToRGBA converts gruid.Color to color.RGBA
func (tm *ASCIITileManager) gruidColorToRGBA(c gruid.Color) color.RGBA {
	r := uint8((c >> 16) & 0xFF)
	g := uint8((c >> 8) & 0xFF)
	b := uint8(c & 0xFF)
	return color.RGBA{r, g, b, 255}
}

// drawCharacter draws a character on the image
func (tm *ASCIITileManager) drawCharacter(img *image.RGBA, r rune, textColor color.RGBA) {
	// Use a font optimized for PyRogue-style display
	face := basicfont.Face7x13

	// Create a font drawer with PyRogue-style positioning
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

	logger.Info("Starting GoRogue", "render_mode", "sdl2_ascii")

	// ゲームエンジンの初期化
	engine := core.NewEngine()
	if engine == nil {
		logger.Fatal("Failed to initialize game engine")
		os.Exit(1)
	}

	// SDL2ドライバーの設定 - ASCII文字表示用（PyRogue風）
	config := sdl.Config{
		TileManager: NewASCIITileManager(10, 14), // 10x14のPyRogue風サイズ
		Width:       80,
		Height:      50,
		WindowTitle: "GoRogue - ASCII Roguelike",
		Fullscreen:  false,
	}

	driver := sdl.NewDriver(config)

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
