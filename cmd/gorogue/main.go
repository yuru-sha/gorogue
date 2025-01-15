package main

import (
	"os"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func main() {
	// ロガーの初期化
	if err := logger.Setup(); err != nil {
		panic(err)
	}
	defer logger.Cleanup()

	// macOSでのEbitenの初期化問題を回避
	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
	}

	// ゲームエンジンの初期化
	engine := core.NewEngine()
	if engine == nil {
		logger.Fatal("Failed to initialize game engine")
		os.Exit(1)
	}

	// ゲームの実行
	if err := ebiten.RunGame(engine); err != nil {
		logger.Fatal("Game terminated with error", "error", err.Error())
		os.Exit(1)
	}
}
