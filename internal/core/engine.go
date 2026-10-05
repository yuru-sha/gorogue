package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/game/score"
	uiscreen "github.com/yuru-sha/gorogue/internal/ui/screen"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

type hangupMessage struct{}

const (
	screenWidth  = 80
	screenHeight = 50
)

// Engine represents the game engine and implements gruid.Model interface
type Engine struct {
	renderer        *uiscreen.TextRenderer
	stateManager    *state.StateManager
	dungeonManager  *dungeon.DungeonManager
	player          *actor.Player
	gameScreen      *uiscreen.GameScreen
	symbolScreen    *uiscreen.SymbolScreen
	saveIntegration *save.SaveGameIntegration
	msgs            []gruid.Msg
	runActive       bool
}

// NewEngine creates and initializes a new game engine
func NewEngine() *Engine {
	return NewEngineWithSeed(time.Now().UnixNano())
}

// NewEngineWithSeed creates an engine with reproducible game randomness.
func NewEngineWithSeed(seed int64) *Engine {
	options := config.DefaultOptions()
	return NewEngineWithOptions(seed, &options)
}

// NewEngineWithOptions creates an engine with explicit Rogue options.
func NewEngineWithOptions(seed int64, options *config.Options) *Engine {
	renderer := uiscreen.NewTextRenderer(screenWidth, screenHeight)

	// プレイヤーの生成（仮位置、後でダンジョンマネージャーが適切な位置に配置）
	player := actor.NewPlayerWithSeed(0, 0, seed)
	logger.Debug("Created player",
		"x", player.Position.X,
		"y", player.Position.Y,
	)

	// ダンジョンマネージャーの生成
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)
	saveIntegration := save.NewSaveGameIntegration()
	saveIntegration.SetOptions(options)
	if err := saveIntegration.Initialize(); err != nil {
		logger.Warn("Failed to initialize save integration", "error", err)
	}
	saveIntegration.SetGameState(player, dungeonManager)
	saveIntegration.SetGameInfo(save.GameInfo{Seed: seed, CharName: options.Name})

	// プレイヤーを最初の部屋の中央に配置
	level := dungeonManager.GetCurrentLevel()
	if len(level.Rooms) > 0 {
		firstRoom := level.Rooms[0]
		player.Position.X = firstRoom.X + firstRoom.Width/2
		player.Position.Y = firstRoom.Y + firstRoom.Height/2
		level.UpdateVisibilityForPlayer(player.Position.X, player.Position.Y, player.HallucinationTurns > 0)
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
	gameScreen.SetOptions(options)
	symbolScreen := uiscreen.NewSymbolScreen(screenWidth, screenHeight)

	logger.Debug("Created screens")

	stateManager := state.NewStateManager()
	stateManager.RegisterState(state.StateGame, gameScreen)
	stateManager.RegisterState(state.StateHelp, gameScreen)
	stateManager.RegisterState(state.StateGameOver, gameScreen)
	stateManager.RegisterState(state.StateVictory, gameScreen)
	stateManager.RegisterState(state.StateSymbol, symbolScreen)
	stateManager.RegisterState(state.StateSettings, gameScreen)
	// Start directly in the already initialized dungeon.
	stateManager.SetState(state.StateGame)

	engine := &Engine{
		renderer:        renderer,
		stateManager:    stateManager,
		dungeonManager:  dungeonManager,
		player:          player,
		gameScreen:      gameScreen,
		symbolScreen:    symbolScreen,
		saveIntegration: saveIntegration,
		runActive:       true,
	}
	engine.bindScreenLoader()

	return engine
}

// bindScreenLoader exposes SyncLoadedState to the game screen so a GUI load
// action can mirror the state produced by the shared command executor into
// the engine and dependent subsystems.
func (e *Engine) bindScreenLoader() {
	if e.gameScreen == nil {
		return
	}
	e.gameScreen.SetEngineLoad(e.SyncLoadedState)
}

// SyncLoadedState mirrors the state held by the save integration into the
// engine and game screen after the shared command executor has validated,
// converted, and consumed the save. It performs no save I/O itself.
func (e *Engine) SyncLoadedState() error {
	if e.saveIntegration == nil {
		return fmt.Errorf("save integration is unavailable")
	}
	player, dm := e.saveIntegration.GetGameState()
	if player == nil || dm == nil {
		return fmt.Errorf("loaded game state is unavailable")
	}
	e.player = player
	e.dungeonManager = dm
	if e.gameScreen != nil {
		e.gameScreen.ApplyLoadedState(player, dm)
	}
	e.runActive = true
	logger.Info("Engine synced loaded state",
		"floor", dm.GetCurrentFloor(),
	)
	return nil
}

// Update implements gruid.Model.Update
func (e *Engine) Update(msg gruid.Msg) gruid.Effect {
	e.msgs = append(e.msgs, msg)

	switch msg := msg.(type) {
	case gruid.MsgInit:
		return gruid.Sub(subscribeForHangup)
	case hangupMessage:
		return e.handleHangup()
	case gruid.MsgKeyDown:
		// キー入力の処理
		previousState := e.stateManager.GetCurrentState()
		effect := e.stateManager.HandleInput(msg)
		switch {
		case previousState == state.StateGame && e.stateManager.GetCurrentState() == state.StateGameOver:
			e.showGameOver()
		case previousState == state.StateGame && e.stateManager.GetCurrentState() == state.StateVictory:
			e.showVictory()
		}
		e.updateRunActive()
		return effect
	case gruid.MsgQuit:
		// 終了処理
		return gruid.End()
	}

	return nil
}

func subscribeForHangup(ctx context.Context, msgs chan<- gruid.Msg) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	defer signal.Stop(signals)
	for {
		select {
		case <-ctx.Done():
			return
		case <-signals:
			select {
			case <-ctx.Done():
				return
			case msgs <- hangupMessage{}:
			}
			<-ctx.Done()
			return
		}
	}
}

func (e *Engine) updateRunActive() {
	switch e.stateManager.GetCurrentState() {
	case state.StateGame:
		e.runActive = true
	case state.StateGameOver, state.StateVictory, state.StateQuit:
		e.runActive = false
	}
}

func (e *Engine) handleHangup() gruid.Effect {
	if !e.runActive {
		return gruid.End()
	}
	if e.saveIntegration == nil {
		logger.Error("SIGHUP recovery save unavailable")
		return gruid.End()
	}
	player, dungeonManager := e.saveIntegration.GetGameState()
	if player == nil || dungeonManager == nil {
		logger.Error("SIGHUP recovery save unavailable")
		return gruid.End()
	}
	e.player, e.dungeonManager = player, dungeonManager
	e.saveIntegration.SetGameState(player, dungeonManager)
	if err := e.saveIntegration.SaveGame(); err != nil {
		logger.Error("SIGHUP recovery save failed", "error", err)
		e.gameScreen.AddMessage("SIGHUP recovery save failed: " + err.Error())
	} else {
		logger.Info("SIGHUP recovery save completed")
	}
	return gruid.End()
}

func (e *Engine) showGameOver() {
	if e.saveIntegration == nil || e.player == nil {
		e.ShowGameOver(nil)
		return
	}

	statsManager := e.saveIntegration.GetGameStats()
	stats := statsManager.GetStats()
	gameInfo := e.saveIntegration.GetGameInfo()
	gameInfo.PlayTime = statsManager.GetPlayTime()
	scoreEntry := score.NewScoreCalculator().CreateScoreEntry(
		gameInfo.CharName,
		e.player,
		&stats,
		&gameInfo,
		false,
		stats.LastDeathReason,
	)
	e.ShowGameOver(&scoreEntry)
}

// Draw implements gruid.Model.Draw
func (e *Engine) Draw() gruid.Grid {
	// グリッドをクリア
	return e.renderer.Draw(e.stateManager)
}

// Model returns the game's model configuration
func (e *Engine) Model() gruid.Model {
	return e
}

// ShowGameOver transitions to the game-over sequence.
func (e *Engine) ShowGameOver(scoreEntry *score.ScoreEntry) {
	deathLines := []string{"You died."}
	scoreLines := []string{"Press any key to continue."}
	if scoreEntry != nil {
		if scoreEntry.DeathReason != "" {
			deathLines = append(deathLines, "Death: "+scoreEntry.DeathReason)
		}
		scoreLines = append([]string{fmt.Sprintf("Score: %d", scoreEntry.Score)}, scoreLines...)
	}
	e.gameScreen.StartDeathSequence(deathLines, scoreLines)
	e.stateManager.SetState(state.StateGameOver)
}

func (e *Engine) showVictory() {
	if e.saveIntegration == nil || e.player == nil {
		e.ShowVictory(nil)
		return
	}
	statsManager := e.saveIntegration.GetGameStats()
	stats := statsManager.GetStats()
	gameInfo := e.saveIntegration.GetGameInfo()
	gameInfo.PlayTime = statsManager.GetPlayTime()
	entry := score.NewScoreCalculator().CreateScoreEntry(gameInfo.CharName, e.player, &stats, &gameInfo, true, "")
	e.ShowVictory(&entry)
}

// ShowVictory transitions to the victory sequence.
func (e *Engine) ShowVictory(scoreEntry *score.ScoreEntry) {
	victoryLines := []string{"Congratulations, you have made it to the light of day!"}
	scoreLines := []string{"You have joined the elite ranks of those who escaped alive."}
	if scoreEntry != nil {
		scoreLines = append(scoreLines, fmt.Sprintf("Score: %d", scoreEntry.Score))
	}
	scoreLines = append(scoreLines, "Press any key to continue.")
	e.gameScreen.StartVictorySequence(victoryLines, scoreLines)
	e.stateManager.SetState(state.StateVictory)
}
