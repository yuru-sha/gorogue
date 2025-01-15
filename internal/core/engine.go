package core

import (
	"image/color"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	uiscreen "github.com/yuru-sha/gorogue/internal/ui/screen"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	screenWidth  = 80
	screenHeight = 50
	tileSize     = 24
)

// Engine represents the game engine and implements ebiten.Game interface
type Engine struct {
	stateManager *state.StateManager
	level        *dungeon.Level
	player       *actor.Player
	gameScreen   *uiscreen.GameScreen
	menuScreen   *uiscreen.MenuScreen
}

// NewEngine creates and initializes a new game engine
func NewEngine() *Engine {
	// macOSでのEbitenの初期化問題を回避
	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
	}

	// フォントの初期化
	if err := uiscreen.InitFont(); err != nil {
		logger.Error("Failed to initialize font", "error", err.Error())
		return nil
	}

	// ダンジョンレベルの生成
	level := dungeon.NewLevel(screenWidth, screenHeight, 1)
	logger.Debug("Created new dungeon level",
		"width", level.Width,
		"height", level.Height,
	)

	// プレイヤーの生成
	player := actor.NewPlayer(screenWidth/2, screenHeight/2)
	logger.Debug("Created player",
		"x", player.Position.X,
		"y", player.Position.Y,
	)

	// 画面の生成
	gameScreen := uiscreen.NewGameScreen(screenWidth, screenHeight, player)
	menuScreen := uiscreen.NewMenuScreen(screenWidth, screenHeight)
	logger.Debug("Created screens")

	// ステートマネージャーの初期化
	stateManager := state.NewStateManager()
	stateManager.RegisterState(state.StateMenu, menuScreen)
	stateManager.RegisterState(state.StateGame, gameScreen)

	engine := &Engine{
		stateManager: stateManager,
		level:        level,
		player:       player,
		gameScreen:   gameScreen,
		menuScreen:   menuScreen,
	}

	return engine
}

// Update handles the game logic updates
func (e *Engine) Update() error {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logger.Error("Panic recovered in Update",
				"panic", r,
				"stack", string(stack),
				"player_x", e.player.Position.X,
				"player_y", e.player.Position.Y,
				"level", e.level.FloorNumber,
			)
		}
	}()

	// 現在の状態を更新
	e.stateManager.Update()

	// ゲームオーバー状態の場合は終了
	if e.stateManager.GetCurrentState() == state.StateGameOver {
		logger.Info("Game over")
		os.Exit(0)
	}

	return nil
}

// Draw renders the game state
func (e *Engine) Draw(screen *ebiten.Image) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logger.Error("Panic recovered in Draw",
				"panic", r,
				"stack", string(stack),
			)
		}
	}()

	// 背景を黒で塗りつぶす
	screen.Fill(color.Black)

	// 現在の状態を描画
	switch e.stateManager.GetCurrentState() {
	case state.StateMenu:
		e.menuScreen.Draw(screen)
	case state.StateGame:
		// ダンジョンの描画
		for y := 0; y < e.level.Height; y++ {
			for x := 0; x < e.level.Width; x++ {
				tile := e.level.GetTile(x, y)
				if tile != nil {
					text.Draw(screen, string(tile.Symbol), uiscreen.GetFont(), x*tileSize, (y+2)*tileSize, color.RGBA{
						R: tile.Color[0],
						G: tile.Color[1],
						B: tile.Color[2],
						A: 255,
					})
				}
			}
		}

		// プレイヤーの描画
		text.Draw(screen, string(e.player.Symbol), uiscreen.GetFont(),
			e.player.Position.X*tileSize,
			(e.player.Position.Y+2)*tileSize,
			color.White)

		// ゲーム画面の描画
		e.gameScreen.Draw(screen)
	}
}

// Layout returns the game's screen dimensions
func (e *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth * tileSize, screenHeight * tileSize
}
