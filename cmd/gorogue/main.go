package main

import (
	"context"
	"os"

	"github.com/anaseto/gruid"
	gtcell "github.com/anaseto/gruid-tcell"
	"github.com/gdamore/tcell/v2"
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// DefaultStyleManager implements tcell.StyleManager interface
type DefaultStyleManager struct{}

// GetStyle implements tcell.StyleManager.GetStyle
func (sm DefaultStyleManager) GetStyle(st gruid.Style) tcell.Style {
	style := tcell.StyleDefault
	if st.Fg != 0 {
		style = style.Foreground(tcell.Color(st.Fg))
	}
	if st.Bg != 0 {
		style = style.Background(tcell.Color(st.Bg))
	}
	return style
}

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
	driver := gtcell.NewDriver(gtcell.Config{
		StyleManager: DefaultStyleManager{},
	})

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
