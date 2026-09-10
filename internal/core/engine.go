package core

import (
	"time"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/game/score"
	uiscreen "github.com/yuru-sha/gorogue/internal/ui/screen"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	screenWidth  = 80
	screenHeight = 50
)

// Engine represents the game engine and implements gruid.Model interface
type Engine struct {
	grid            gruid.Grid
	stateManager    *state.StateManager
	dungeonManager  *dungeon.DungeonManager
	player          *actor.Player
	gameScreen      *uiscreen.GameScreen
	menuScreen      *uiscreen.MenuScreen
	helpScreen      *uiscreen.HelpScreen
	gameOverScreen  *uiscreen.GameOverScreen
	victoryScreen   *uiscreen.VictoryScreen
	symbolScreen    *uiscreen.SymbolScreen
	saveIntegration *save.SaveGameIntegration
	msgs            []gruid.Msg
}

// NewEngine creates and initializes a new game engine
func NewEngine() *Engine {
	return NewEngineWithSeed(time.Now().UnixNano())
}

// NewEngineWithSeed creates an engine with reproducible game randomness.
func NewEngineWithSeed(seed int64) *Engine {
	// グリッドの初期化
	grid := gruid.NewGrid(screenWidth, screenHeight)

	// プレイヤーの生成（仮位置、後でダンジョンマネージャーが適切な位置に配置）
	player := actor.NewPlayerWithSeed(0, 0, seed)
	logger.Debug("Created player",
		"x", player.Position.X,
		"y", player.Position.Y,
	)

	// ダンジョンマネージャーの生成
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)
	saveIntegration := save.NewSaveGameIntegration()
	if err := saveIntegration.Initialize(); err != nil {
		logger.Warn("Failed to initialize save integration", "error", err)
	}
	saveIntegration.SetGameState(player, dungeonManager)

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
	gameScreen.SetSaveIntegration(saveIntegration)
	menuScreen := uiscreen.NewMenuScreen(screenWidth, screenHeight)
	helpScreen := uiscreen.NewHelpScreen(screenWidth, screenHeight)
	symbolScreen := uiscreen.NewSymbolScreen(screenWidth, screenHeight)
	saveLoadScreen := uiscreen.NewSaveLoadScreenWithIntegration(screenWidth, screenHeight, saveIntegration)

	// ゲームオーバー・勝利画面（初期は空のスコアエントリーで作成）
	gameOverScreen := uiscreen.NewGameOverScreen(screenWidth, screenHeight, nil)
	victoryScreen := uiscreen.NewVictoryScreen(screenWidth, screenHeight, nil)

	logger.Debug("Created screens")

	// ステートマネージャーの初期化
	stateManager := state.NewStateManager()
	stateManager.RegisterState(state.StateMenu, menuScreen)
	stateManager.RegisterState(state.StateGame, gameScreen)
	stateManager.RegisterState(state.StateHelp, helpScreen)
	stateManager.RegisterState(state.StateGameOver, gameOverScreen)
	stateManager.RegisterState(state.StateVictory, victoryScreen)
	stateManager.RegisterState(state.StateSymbol, symbolScreen)
	stateManager.RegisterState(state.StateSaveLoad, saveLoadScreen)

	// メニュー状態で開始
	stateManager.SetState(state.StateMenu)

	engine := &Engine{
		grid:            grid,
		stateManager:    stateManager,
		dungeonManager:  dungeonManager,
		player:          player,
		gameScreen:      gameScreen,
		menuScreen:      menuScreen,
		helpScreen:      helpScreen,
		gameOverScreen:  gameOverScreen,
		victoryScreen:   victoryScreen,
		symbolScreen:    symbolScreen,
		saveIntegration: saveIntegration,
		msgs:            make([]gruid.Msg, 0),
	}
	saveLoadScreen.SetOnLoad(engine.restoreLoadedGame)

	return engine
}

func (e *Engine) restoreLoadedGame(player *actor.Player, dungeonManager *dungeon.DungeonManager) {
	e.player = player
	e.dungeonManager = dungeonManager
	e.gameScreen = uiscreen.NewGameScreen(screenWidth, screenHeight, player)
	e.gameScreen.SetLevel(dungeonManager.GetCurrentLevel())
	e.gameScreen.SetDungeonManager(dungeonManager)
	e.gameScreen.SetSaveIntegration(e.saveIntegration)
	e.stateManager.RegisterState(state.StateGame, e.gameScreen)
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

	// 現在の状態を描画 - state managerを使用
	e.stateManager.Draw(&e.grid)

	return e.grid
}

// Model returns the game's model configuration
func (e *Engine) Model() gruid.Model {
	return e
}

// ShowGameOver transitions to the game over screen with the given score entry
func (e *Engine) ShowGameOver(scoreEntry *score.ScoreEntry) {
	e.gameOverScreen = uiscreen.NewGameOverScreen(screenWidth, screenHeight, scoreEntry)
	e.stateManager.RegisterState(state.StateGameOver, e.gameOverScreen)
	e.stateManager.SetState(state.StateGameOver)
}

// ShowVictory transitions to the victory screen with the given score entry
func (e *Engine) ShowVictory(scoreEntry *score.ScoreEntry) {
	e.victoryScreen = uiscreen.NewVictoryScreen(screenWidth, screenHeight, scoreEntry)
	e.stateManager.RegisterState(state.StateVictory, e.victoryScreen)
	e.stateManager.SetState(state.StateVictory)
}
