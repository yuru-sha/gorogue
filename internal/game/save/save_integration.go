// Package save ゲームエンジンとの統合機能
// セーブ/ロード機能をゲームエンジンに統合し、UI連携を提供
package save

import (
	"fmt"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// SaveGameIntegration handles integration between save system and game engine
type SaveGameIntegration struct {
	saveManager   *SaveManager
	saveConverter *SaveConverter
	gameStats     *GameStats
	autoSave      *AutoSaveManager

	// Game state
	player         *actor.Player
	dungeonManager *dungeon.DungeonManager
	gameInfo       GameInfo
	settings       Settings
}

// NewSaveGameIntegration creates a new save game integration
func NewSaveGameIntegration() *SaveGameIntegration {
	return &SaveGameIntegration{
		saveManager:   NewSaveManager(),
		saveConverter: NewSaveConverter(),
		gameStats:     NewGameStats(),
		autoSave:      NewAutoSaveManager(),
		settings:      GetDefaultSettings(),
	}
}

// Initialize initializes the save game integration
func (sgi *SaveGameIntegration) Initialize() error {
	if err := sgi.saveManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize save manager: %w", err)
	}

	if err := sgi.autoSave.Initialize(sgi.saveManager); err != nil {
		return fmt.Errorf("failed to initialize auto-save: %w", err)
	}

	logger.Info("Save game integration initialized")
	return nil
}

// SaveGame saves the current game state (PyRogue style - single save)
func (sgi *SaveGameIntegration) SaveGame() error {
	if sgi.player == nil || sgi.dungeonManager == nil {
		return fmt.Errorf("game state not set")
	}

	// Update game info
	sgi.gameInfo.PlayTime = sgi.gameStats.GetPlayTime()
	sgi.gameInfo.TurnCount = sgi.gameStats.GetTurnCount()

	// Create save data
	saveData := ToSaveData(
		sgi.player,
		sgi.dungeonManager,
		sgi.gameInfo,
		sgi.gameStats.GetStats(),
		sgi.settings,
	)

	// Save to file
	if err := sgi.saveManager.SaveGame(saveData); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	logger.Info("Game saved successfully",
		"char_name", sgi.gameInfo.CharName,
		"level", sgi.player.Level,
		"floor", sgi.dungeonManager.GetCurrentFloor(),
	)

	return nil
}

// LoadGame loads the game state (PyRogue style - single save)
func (sgi *SaveGameIntegration) LoadGame() error {
	// Load save data
	saveData, err := sgi.saveManager.LoadGame()
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}

	// Convert save data to game objects
	player, dungeonManager, err := sgi.saveConverter.FromSaveData(saveData)
	if err != nil {
		return fmt.Errorf("failed to convert save data: %w", err)
	}

	// Set game state
	sgi.player = player
	sgi.dungeonManager = dungeonManager
	sgi.gameInfo = saveData.GameInfo
	sgi.settings = saveData.Settings

	// Update game stats
	sgi.gameStats.LoadStats(saveData.GameStats)

	logger.Info("Game loaded successfully",
		"char_name", sgi.gameInfo.CharName,
		"level", sgi.player.Level,
		"floor", sgi.dungeonManager.GetCurrentFloor(),
		"version", saveData.Version,
	)

	return nil
}

// QuickSave performs a quick save (PyRogue style - same as normal save)
func (sgi *SaveGameIntegration) QuickSave() error {
	return sgi.SaveGame()
}

// QuickLoad performs a quick load (PyRogue style - same as normal load)
func (sgi *SaveGameIntegration) QuickLoad() error {
	return sgi.LoadGame()
}

// AutoSave performs an automatic save
func (sgi *SaveGameIntegration) AutoSave() error {
	if !sgi.settings.AutoSave {
		return nil
	}

	if sgi.player == nil || sgi.dungeonManager == nil {
		return fmt.Errorf("game state not set")
	}

	// Update game info
	sgi.gameInfo.PlayTime = sgi.gameStats.GetPlayTime()
	sgi.gameInfo.TurnCount = sgi.gameStats.GetTurnCount()

	// Create save data
	saveData := ToSaveData(
		sgi.player,
		sgi.dungeonManager,
		sgi.gameInfo,
		sgi.gameStats.GetStats(),
		sgi.settings,
	)

	return sgi.autoSave.AutoSave(saveData)
}

// HasAutoSave checks if an auto-save exists
func (sgi *SaveGameIntegration) HasAutoSave() bool {
	return sgi.autoSave.HasAutoSave()
}

// LoadAutoSave loads the auto-save
func (sgi *SaveGameIntegration) LoadAutoSave() error {
	// Load auto-save data
	saveData, err := sgi.autoSave.LoadAutoSave()
	if err != nil {
		return fmt.Errorf("failed to load auto-save: %w", err)
	}

	// Convert save data to game objects
	player, dungeonManager, err := sgi.saveConverter.FromSaveData(saveData)
	if err != nil {
		return fmt.Errorf("failed to convert auto-save data: %w", err)
	}

	// Set game state
	sgi.player = player
	sgi.dungeonManager = dungeonManager
	sgi.gameInfo = saveData.GameInfo
	sgi.settings = saveData.Settings

	// Update game stats
	sgi.gameStats.LoadStats(saveData.GameStats)

	logger.Info("Auto-save loaded successfully",
		"char_name", sgi.gameInfo.CharName,
		"level", sgi.player.Level,
		"floor", sgi.dungeonManager.GetCurrentFloor(),
	)

	return nil
}

// SetGameState sets the current game state
func (sgi *SaveGameIntegration) SetGameState(player *actor.Player, dungeonManager *dungeon.DungeonManager) {
	sgi.player = player
	sgi.dungeonManager = dungeonManager
}

// GetGameState returns the current game state
func (sgi *SaveGameIntegration) GetGameState() (*actor.Player, *dungeon.DungeonManager) {
	return sgi.player, sgi.dungeonManager
}

// SetGameInfo sets the game information
func (sgi *SaveGameIntegration) SetGameInfo(gameInfo GameInfo) {
	sgi.gameInfo = gameInfo
}

// GetGameInfo returns the game information
func (sgi *SaveGameIntegration) GetGameInfo() GameInfo {
	return sgi.gameInfo
}

// GetSettings returns the current settings
func (sgi *SaveGameIntegration) GetSettings() Settings {
	return sgi.settings
}

// SetSettings sets the game settings
func (sgi *SaveGameIntegration) SetSettings(settings Settings) {
	sgi.settings = settings
}

// GetGameStats returns the game statistics manager
func (sgi *SaveGameIntegration) GetGameStats() *GameStats {
	return sgi.gameStats
}

// DeleteSave deletes the save file (PyRogue style - single save)
func (sgi *SaveGameIntegration) DeleteSave() error {
	return sgi.saveManager.DeleteSave()
}

// HasSave checks if a save file exists (PyRogue style - single save)
func (sgi *SaveGameIntegration) HasSave() bool {
	return sgi.saveManager.FileExists()
}

// GetSaveInfo returns information about the save file (PyRogue style - single save)
func (sgi *SaveGameIntegration) GetSaveInfo() (string, error) {
	return sgi.saveManager.GetSaveInfo()
}

// ExportSave exports the save file (PyRogue style - single save)
func (sgi *SaveGameIntegration) ExportSave(path string) error {
	return sgi.saveManager.ExportSave(path)
}

// ImportSave imports a save file (PyRogue style - single save)
func (sgi *SaveGameIntegration) ImportSave(path string) error {
	return sgi.saveManager.ImportSave(path)
}

// ShouldAutoSave checks if auto-save should be performed
func (sgi *SaveGameIntegration) ShouldAutoSave() bool {
	if !sgi.settings.AutoSave {
		return false
	}

	return sgi.autoSave.ShouldAutoSave(sgi.gameStats.GetTurnCount())
}

// OnPlayerDeath handles player death
func (sgi *SaveGameIntegration) OnPlayerDeath(reason string) {
	sgi.gameStats.OnPlayerDeath(reason, sgi.dungeonManager.GetCurrentFloor())

	// Save death state if enabled
	if sgi.settings.AutoSave {
		sgi.gameInfo.IsCompleted = true
		sgi.gameInfo.IsVictory = false

		if err := sgi.AutoSave(); err != nil {
			logger.Error("Failed to save death state", "error", err)
		}
	}
}

// OnPlayerVictory handles player victory
func (sgi *SaveGameIntegration) OnPlayerVictory() {
	sgi.gameStats.OnPlayerVictory()

	// Save victory state
	sgi.gameInfo.IsCompleted = true
	sgi.gameInfo.IsVictory = true

	if err := sgi.AutoSave(); err != nil {
		logger.Error("Failed to save victory state", "error", err)
	}
}

// OnTurnEnd handles end of turn processing
func (sgi *SaveGameIntegration) OnTurnEnd() {
	sgi.gameStats.OnTurnEnd()

	// Check for auto-save
	if sgi.ShouldAutoSave() {
		if err := sgi.AutoSave(); err != nil {
			logger.Error("Auto-save failed", "error", err)
		}
	}
}

// OnFloorChange handles floor change
func (sgi *SaveGameIntegration) OnFloorChange(newFloor int) {
	sgi.gameStats.OnFloorChange(newFloor)

	// Auto-save on floor change if enabled
	if sgi.settings.AutoSave {
		if err := sgi.AutoSave(); err != nil {
			logger.Error("Auto-save on floor change failed", "error", err)
		}
	}
}

// OnMonsterKilled handles monster death
func (sgi *SaveGameIntegration) OnMonsterKilled(monster *actor.Monster) {
	sgi.gameStats.OnMonsterKilled(monster)
}

// OnItemFound handles item discovery
func (sgi *SaveGameIntegration) OnItemFound(item string) {
	sgi.gameStats.OnItemFound(item)
}

// OnItemUsed handles item usage
func (sgi *SaveGameIntegration) OnItemUsed(item string) {
	sgi.gameStats.OnItemUsed(item)
}

// OnDamageDealt handles damage dealt
func (sgi *SaveGameIntegration) OnDamageDealt(damage int) {
	sgi.gameStats.OnDamageDealt(damage)
}

// OnDamageTaken handles damage taken
func (sgi *SaveGameIntegration) OnDamageTaken(damage int) {
	sgi.gameStats.OnDamageTaken(damage)
}

// OnGoldCollected handles gold collection
func (sgi *SaveGameIntegration) OnGoldCollected(amount int) {
	sgi.gameStats.OnGoldCollected(amount)
}

// GetSaveManager returns the save manager
func (sgi *SaveGameIntegration) GetSaveManager() *SaveManager {
	return sgi.saveManager
}

// GetSaveConverter returns the save converter
func (sgi *SaveGameIntegration) GetSaveConverter() *SaveConverter {
	return sgi.saveConverter
}

// GetAutoSaveManager returns the auto-save manager
func (sgi *SaveGameIntegration) GetAutoSaveManager() *AutoSaveManager {
	return sgi.autoSave
}

// Validate validates the current save system state
func (sgi *SaveGameIntegration) Validate() error {
	if sgi.saveManager == nil {
		return fmt.Errorf("save manager not initialized")
	}

	if sgi.saveConverter == nil {
		return fmt.Errorf("save converter not initialized")
	}

	if sgi.gameStats == nil {
		return fmt.Errorf("game stats not initialized")
	}

	if sgi.autoSave == nil {
		return fmt.Errorf("auto-save manager not initialized")
	}

	return nil
}

// GetStatus returns the current save system status
func (sgi *SaveGameIntegration) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"initialized":       sgi.Validate() == nil,
		"has_game_state":    sgi.player != nil && sgi.dungeonManager != nil,
		"auto_save_enabled": sgi.settings.AutoSave,
		"has_auto_save":     sgi.HasAutoSave(),
		"save_directory":    sgi.saveManager.GetSaveDirectory(),
		"has_save_file":     sgi.saveManager.FileExists(),
	}

	if sgi.player != nil {
		status["player_level"] = sgi.player.Level
		status["player_hp"] = sgi.player.HP
		status["player_gold"] = sgi.player.Gold
	}

	if sgi.dungeonManager != nil {
		status["current_floor"] = sgi.dungeonManager.GetCurrentFloor()
	}

	if diskUsage, err := sgi.saveManager.GetDiskUsage(); err == nil {
		status["disk_usage"] = diskUsage
	}

	return status
}

// RepairSave attempts to repair a corrupted save file (PyRogue style - single save)
func (sgi *SaveGameIntegration) RepairSave() error {
	return sgi.saveManager.RepairSave()
}

// CreateNewGame creates a new game with the specified parameters
func (sgi *SaveGameIntegration) CreateNewGame(charName string, seed int64) error {
	// Create new player
	player := actor.NewPlayer(0, 0)

	// Create new dungeon manager
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)

	// Set initial position
	level := dungeonManager.GetCurrentLevel()
	if len(level.Rooms) > 0 {
		firstRoom := level.Rooms[0]
		player.Position.X = firstRoom.X + firstRoom.Width/2
		player.Position.Y = firstRoom.Y + firstRoom.Height/2
	}

	// Set game state
	sgi.player = player
	sgi.dungeonManager = dungeonManager

	// Initialize game info
	sgi.gameInfo = GameInfo{
		Seed:        seed,
		PlayTime:    0,
		TurnCount:   0,
		CharName:    charName,
		Difficulty:  "Normal",
		GameMode:    "Normal",
		IsWizard:    false,
		IsCompleted: false,
		IsVictory:   false,
	}

	// Reset game stats
	sgi.gameStats.Reset()

	logger.Info("New game created",
		"char_name", charName,
		"seed", seed,
		"player_pos", fmt.Sprintf("(%d,%d)", player.Position.X, player.Position.Y),
	)

	return nil
}

// GetDefaultSettings returns default game settings
func GetDefaultSettings() Settings {
	return Settings{
		ShowTips:     true,
		AutoPickup:   true,
		ConfirmQuit:  true,
		AutoSave:     true,
		SaveInterval: 100, // Every 100 turns
		WizardMode:   false,
		DebugMode:    false,
		KeyBindings:  GetDefaultKeyBindings(),
	}
}

// GetDefaultKeyBindings returns default key bindings
func GetDefaultKeyBindings() map[string]string {
	return map[string]string{
		"move_north":     "k",
		"move_south":     "j",
		"move_east":      "l",
		"move_west":      "h",
		"move_northeast": "u",
		"move_northwest": "y",
		"move_southeast": "n",
		"move_southwest": "b",
		"inventory":      "i",
		"equipment":      "W",
		"drop":           "d",
		"pick_up":        "g",
		"quaff":          "q",
		"read":           "r",
		"save":           "S",
		"load":           "L",
		"quick_save":     "ctrl+s",
		"quick_load":     "ctrl+l",
		"help":           "?",
		"quit":           "Q",
	}
}
