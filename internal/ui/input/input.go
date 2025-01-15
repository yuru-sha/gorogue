package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	initialDelay = 20 // フレーム数
	repeatDelay  = 5  // フレーム数
)

// shouldRepeat キーリピートの判定
func shouldRepeat(key ebiten.Key) bool {
	duration := inpututil.KeyPressDuration(key)
	if duration == 1 {
		return true
	}
	if duration >= initialDelay && (duration-initialDelay)%repeatDelay == 0 {
		return true
	}
	return false
}

// GetMovementDirection キー入力から移動方向を取得
func GetMovementDirection() (dx, dy int) {
	// vi キー
	if (ebiten.IsKeyPressed(ebiten.KeyH) && shouldRepeat(ebiten.KeyH)) ||
		(ebiten.IsKeyPressed(ebiten.KeyLeft) && shouldRepeat(ebiten.KeyLeft)) {
		logger.Debug("Left movement input detected", "key", "h/left", "dx", -1, "dy", 0)
		return -1, 0
	}
	if (ebiten.IsKeyPressed(ebiten.KeyL) && shouldRepeat(ebiten.KeyL)) ||
		(ebiten.IsKeyPressed(ebiten.KeyRight) && shouldRepeat(ebiten.KeyRight)) {
		logger.Debug("Right movement input detected", "key", "l/right", "dx", 1, "dy", 0)
		return 1, 0
	}
	if (ebiten.IsKeyPressed(ebiten.KeyK) && shouldRepeat(ebiten.KeyK)) ||
		(ebiten.IsKeyPressed(ebiten.KeyUp) && shouldRepeat(ebiten.KeyUp)) {
		logger.Debug("Up movement input detected", "key", "k/up", "dx", 0, "dy", -1)
		return 0, -1
	}
	if (ebiten.IsKeyPressed(ebiten.KeyJ) && shouldRepeat(ebiten.KeyJ)) ||
		(ebiten.IsKeyPressed(ebiten.KeyDown) && shouldRepeat(ebiten.KeyDown)) {
		logger.Debug("Down movement input detected", "key", "j/down", "dx", 0, "dy", 1)
		return 0, 1
	}
	// 斜め移動
	if ebiten.IsKeyPressed(ebiten.KeyY) && shouldRepeat(ebiten.KeyY) {
		logger.Debug("Diagonal movement input detected", "key", "y", "dx", -1, "dy", -1)
		return -1, -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyU) && shouldRepeat(ebiten.KeyU) {
		logger.Debug("Diagonal movement input detected", "key", "u", "dx", 1, "dy", -1)
		return 1, -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyB) && shouldRepeat(ebiten.KeyB) {
		logger.Debug("Diagonal movement input detected", "key", "b", "dx", -1, "dy", 1)
		return -1, 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyN) && shouldRepeat(ebiten.KeyN) {
		logger.Debug("Diagonal movement input detected", "key", "n", "dx", 1, "dy", 1)
		return 1, 1
	}

	return 0, 0
}

// IsQuitRequested ゲーム終了の判定
func IsQuitRequested() bool {
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		logger.Info("Quit requested by user", "key", "q")
		return true
	}
	return false
}
