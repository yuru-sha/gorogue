// GoRogue - SDL2グラフィックス版メイン

package main

import (
	"context"
	"image"
	"image/color"
	"os"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid-sdl"
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// SimpleTileManager implements a basic tile manager for SDL2
type SimpleTileManager struct{}

// TileSize returns the size of individual tiles
func (tm SimpleTileManager) TileSize() gruid.Point {
	return gruid.Point{X: 16, Y: 16}
}

// GetImage returns an image for a specific cell
func (tm SimpleTileManager) GetImage(cell gruid.Cell) image.Image {
	// シンプルなASCII文字のレンダリング
	tileSize := tm.TileSize()
	img := image.NewRGBA(image.Rect(0, 0, tileSize.X, tileSize.Y))
	
	// 背景色の設定
	bg := color.RGBA{0, 0, 0, 255} // 黒背景
	if cell.Style.Bg != 0 {
		bg = convertGruidColor(cell.Style.Bg)
	}
	
	// 文字色の設定
	fg := color.RGBA{255, 255, 255, 255} // 白文字
	if cell.Style.Fg != 0 {
		fg = convertGruidColor(cell.Style.Fg)
	}
	
	// 背景を塗りつぶす
	for y := 0; y < tileSize.Y; y++ {
		for x := 0; x < tileSize.X; x++ {
			img.Set(x, y, bg)
		}
	}
	
	// 文字に応じた簡単な図形を描画
	switch cell.Rune {
	case '@': // プレイヤー
		drawCircle(img, tileSize.X/2, tileSize.Y/2, 6, fg)
	case '#': // 壁
		drawRectangle(img, 0, 0, tileSize.X, tileSize.Y, fg)
	case '.': // 床
		drawPixel(img, tileSize.X/2, tileSize.Y/2, fg)
	case '+': // 扉
		drawRectangle(img, 2, 2, tileSize.X-4, tileSize.Y-4, fg)
	case '<': // 上り階段
		drawTriangle(img, fg, true)
	case '>': // 下り階段
		drawTriangle(img, fg, false)
	case 'B', 'D', 'E', 'F', 'G', 'O', 'S', 'T': // モンスター
		drawMonster(img, cell.Rune, fg)
	default: // その他の文字
		drawChar(img, cell.Rune, fg)
	}
	
	return img
}

// convertGruidColor converts a gruid color to RGBA
func convertGruidColor(c gruid.Color) color.RGBA {
	// GruidのColorは数値なので、RGB値に変換
	r := uint8((c >> 16) & 0xFF)
	g := uint8((c >> 8) & 0xFF)
	b := uint8(c & 0xFF)
	
	// 色が0の場合はデフォルト色を使用
	if c == 0 {
		return color.RGBA{128, 128, 128, 255}
	}
	
	return color.RGBA{r, g, b, 255}
}

// drawCircle draws a simple circle
func drawCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for y := cy - radius; y <= cy + radius; y++ {
		for x := cx - radius; x <= cx + radius; x++ {
			if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
				dx := x - cx
				dy := y - cy
				if dx*dx + dy*dy <= radius*radius {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// drawRectangle draws a rectangle
func drawRectangle(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	for py := y; py < y + height; py++ {
		for px := x; px < x + width; px++ {
			if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
				img.Set(px, py, c)
			}
		}
	}
}

// drawPixel draws a single pixel
func drawPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
		img.Set(x, y, c)
	}
}

// drawTriangle draws a simple triangle
func drawTriangle(img *image.RGBA, c color.RGBA, up bool) {
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	
	if up {
		// 上向きの三角形
		for y := height/4; y < 3*height/4; y++ {
			lineWidth := (y - height/4) * width / height
			for x := width/2 - lineWidth/2; x < width/2 + lineWidth/2; x++ {
				if x >= 0 && x < width && y >= 0 && y < height {
					img.Set(x, y, c)
				}
			}
		}
	} else {
		// 下向きの三角形
		for y := height/4; y < 3*height/4; y++ {
			lineWidth := (3*height/4 - y) * width / height
			for x := width/2 - lineWidth/2; x < width/2 + lineWidth/2; x++ {
				if x >= 0 && x < width && y >= 0 && y < height {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// drawMonster draws a simple monster representation
func drawMonster(img *image.RGBA, monster rune, c color.RGBA) {
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	
	switch monster {
	case 'B': // バット - 小さな円
		drawCircle(img, width/2, height/2, 3, c)
	case 'D': // ドラゴン - 大きな四角
		drawRectangle(img, 2, 2, width-4, height-4, c)
	case 'G': // ゴブリン - 中程度の円
		drawCircle(img, width/2, height/2, 5, c)
	default: // その他のモンスター
		drawCircle(img, width/2, height/2, 4, c)
	}
}

// drawChar draws a simple character representation
func drawChar(img *image.RGBA, char rune, c color.RGBA) {
	// 文字の簡単な表現（中央に点）
	drawPixel(img, img.Bounds().Max.X/2, img.Bounds().Max.Y/2, c)
}

func main() {
	// ロガーの初期化
	if err := logger.Setup(); err != nil {
		panic(err)
	}
	defer logger.Cleanup()

	logger.Info("Starting GoRogue", "render_mode", "sdl2")

	// ゲームエンジンの初期化
	engine := core.NewEngine()
	if engine == nil {
		logger.Fatal("Failed to initialize game engine")
		os.Exit(1)
	}

	// SDL2ドライバーの設定
	config := sdl.Config{
		TileManager: SimpleTileManager{},
		Width:       80,
		Height:      50,
		WindowTitle: "GoRogue - A Roguelike Game",
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