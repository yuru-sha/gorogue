package core

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	uiscreen "github.com/yuru-sha/gorogue/internal/ui/screen"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	screenWidth  = 80
	screenHeight = 50
)

// Engine represents the game engine and implements gruid.Model interface
type Engine struct {
	grid           gruid.Grid
	stateManager   *state.StateManager
	dungeonManager *dungeon.DungeonManager
	player         *actor.Player
	gameScreen     *uiscreen.GameScreen
	menuScreen     *uiscreen.MenuScreen
	msgs           []gruid.Msg
}

// NewEngine creates and initializes a new game engine
func NewEngine() *Engine {
	// グリッドの初期化
	grid := gruid.NewGrid(screenWidth, screenHeight)

	// プレイヤーの生成（仮位置、後でダンジョンマネージャーが適切な位置に配置）
	player := actor.NewPlayer(0, 0)
	logger.Debug("Created player",
		"x", player.Position.X,
		"y", player.Position.Y,
	)

	// ダンジョンマネージャーの生成
	dungeonManager := dungeon.NewDungeonManager(player)

	// プレイヤーを最初の部屋の中央に配置
	level := dungeonManager.GetCurrentLevel()
	if len(level.Rooms) > 0 {
		firstRoom := level.Rooms[0]
		player.Position.X = firstRoom.X + firstRoom.Width/2
		player.Position.Y = firstRoom.Y + firstRoom.Height/2
		logger.Debug("Placed player in first room",
			"x", player.Position.X,
			"y", player.Position.Y,
			"room_x", firstRoom.X,
			"room_y", firstRoom.Y,
		)
	}

	// 画面の生成
	gameScreen := uiscreen.NewGameScreen(screenWidth, screenHeight, player)
	gameScreen.SetLevel(level)                   // ダンジョンレベルを設定
	gameScreen.SetDungeonManager(dungeonManager) // ダンジョンマネージャーを設定
	menuScreen := uiscreen.NewMenuScreen(screenWidth, screenHeight)
	logger.Debug("Created screens")

	// ステートマネージャーの初期化
	stateManager := state.NewStateManager()
	stateManager.RegisterState(state.StateMenu, menuScreen)
	stateManager.RegisterState(state.StateGame, gameScreen)

	// ゲーム状態で開始
	stateManager.SetState(state.StateGame)

	engine := &Engine{
		grid:           grid,
		stateManager:   stateManager,
		dungeonManager: dungeonManager,
		player:         player,
		gameScreen:     gameScreen,
		menuScreen:     menuScreen,
		msgs:           make([]gruid.Msg, 0),
	}

	return engine
}

// Update implements gruid.Model.Update
func (e *Engine) Update(msg gruid.Msg) gruid.Effect {
	e.msgs = append(e.msgs, msg)

	switch msg := msg.(type) {
	case gruid.MsgInit:
		// 初期化時の処理
		return nil
	case gruid.MsgKeyDown:
		// キー入力の処理
		return e.stateManager.HandleInput(msg)
	case gruid.MsgQuit:
		// 終了処理
		return gruid.End()
	}

	return nil
}

// Draw implements gruid.Model.Draw
func (e *Engine) Draw() gruid.Grid {
	// グリッドをクリア
	e.grid.Fill(gruid.Cell{Rune: ' '})

	// 現在の状態を描画
	switch e.stateManager.GetCurrentState() {
	case state.StateMenu:
		e.menuScreen.Draw(&e.grid)
	case state.StateGame:
		e.gameScreen.Draw(&e.grid)
	}

	return e.grid
}

// Model returns the game's model configuration
func (e *Engine) Model() gruid.Model {
	return e
}
