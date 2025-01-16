package main

import (
	"context"
	"os"

	"github.com/anaseto/gruid"
	tcell "github.com/anaseto/gruid-tcell"
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func main() {
	// ロガーの初期化
	if err := logger.Setup(); err != nil {
		panic(err)
	}
	defer logger.Cleanup()

	// ゲームエンジンの初期化
	engine := core.NewEngine()
	if engine == nil {
		logger.Fatal("Failed to initialize game engine")
		os.Exit(1)
	}

	// ドライバーの設定
	driver := tcell.NewDriver(tcell.Config{})

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
