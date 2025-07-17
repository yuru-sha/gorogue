// Package save オートセーブ機能
// 定期的な自動セーブ、条件付きセーブ、セーブ間隔管理を提供
package save

import (
	"fmt"
	"time"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// AutoSaveManager manages automatic save functionality
type AutoSaveManager struct {
	saveManager   *SaveManager
	lastSaveTurn  int
	lastSaveTime  time.Time
	saveInterval  int // Turns between auto-saves
	enabled       bool
	saveOnFloor   bool // Auto-save when changing floors
	saveOnDeath   bool // Auto-save when player dies
	saveOnVictory bool // Auto-save when player wins
	saveOnExit    bool // Auto-save when exiting game
}

// NewAutoSaveManager creates a new auto-save manager
func NewAutoSaveManager() *AutoSaveManager {
	return &AutoSaveManager{
		lastSaveTurn:  0,
		lastSaveTime:  time.Now(),
		saveInterval:  100, // Default: every 100 turns
		enabled:       true,
		saveOnFloor:   true,
		saveOnDeath:   true,
		saveOnVictory: true,
		saveOnExit:    true,
	}
}

// Initialize initializes the auto-save manager
func (asm *AutoSaveManager) Initialize(saveManager *SaveManager) error {
	if saveManager == nil {
		return fmt.Errorf("save manager cannot be nil")
	}

	asm.saveManager = saveManager
	asm.lastSaveTime = time.Now()

	logger.Info("Auto-save manager initialized",
		"enabled", asm.enabled,
		"interval", asm.saveInterval,
		"save_on_floor", asm.saveOnFloor,
		"save_on_death", asm.saveOnDeath,
		"save_on_victory", asm.saveOnVictory,
		"save_on_exit", asm.saveOnExit,
	)

	return nil
}

// AutoSave performs an automatic save
func (asm *AutoSaveManager) AutoSave(saveData *SaveData) error {
	if !asm.enabled {
		return nil
	}

	if asm.saveManager == nil {
		return fmt.Errorf("save manager not initialized")
	}

	// Update auto-save tracking
	asm.lastSaveTurn = saveData.GameInfo.TurnCount
	asm.lastSaveTime = time.Now()

	// Perform auto-save
	if err := asm.saveManager.AutoSave(saveData); err != nil {
		return fmt.Errorf("auto-save failed: %w", err)
	}

	logger.Info("Auto-save completed",
		"turn", saveData.GameInfo.TurnCount,
		"char_name", saveData.GameInfo.CharName,
		"level", saveData.PlayerData.Level,
		"floor", saveData.DungeonData.CurrentFloor,
	)

	return nil
}

// ShouldAutoSave checks if auto-save should be performed based on turn count
func (asm *AutoSaveManager) ShouldAutoSave(currentTurn int) bool {
	if !asm.enabled {
		return false
	}

	return currentTurn-asm.lastSaveTurn >= asm.saveInterval
}

// ShouldAutoSaveTime checks if auto-save should be performed based on time
func (asm *AutoSaveManager) ShouldAutoSaveTime(timeInterval time.Duration) bool {
	if !asm.enabled {
		return false
	}

	return time.Since(asm.lastSaveTime) >= timeInterval
}

// SaveOnFloorChange performs auto-save when changing floors
func (asm *AutoSaveManager) SaveOnFloorChange(saveData *SaveData) error {
	if !asm.enabled || !asm.saveOnFloor {
		return nil
	}

	logger.Debug("Auto-save triggered by floor change",
		"new_floor", saveData.DungeonData.CurrentFloor,
	)

	return asm.AutoSave(saveData)
}

// SaveOnDeath performs auto-save when player dies
func (asm *AutoSaveManager) SaveOnDeath(saveData *SaveData, reason string) error {
	if !asm.enabled || !asm.saveOnDeath {
		return nil
	}

	logger.Debug("Auto-save triggered by player death",
		"reason", reason,
		"floor", saveData.DungeonData.CurrentFloor,
	)

	return asm.AutoSave(saveData)
}

// SaveOnVictory performs auto-save when player wins
func (asm *AutoSaveManager) SaveOnVictory(saveData *SaveData) error {
	if !asm.enabled || !asm.saveOnVictory {
		return nil
	}

	logger.Debug("Auto-save triggered by victory")

	return asm.AutoSave(saveData)
}

// SaveOnExit performs auto-save when exiting game
func (asm *AutoSaveManager) SaveOnExit(saveData *SaveData) error {
	if !asm.enabled || !asm.saveOnExit {
		return nil
	}

	logger.Debug("Auto-save triggered by game exit")

	return asm.AutoSave(saveData)
}

// HasAutoSave checks if an auto-save exists
func (asm *AutoSaveManager) HasAutoSave() bool {
	if asm.saveManager == nil {
		return false
	}

	return asm.saveManager.HasAutoSave()
}

// LoadAutoSave loads the auto-save
func (asm *AutoSaveManager) LoadAutoSave() (*SaveData, error) {
	if asm.saveManager == nil {
		return nil, fmt.Errorf("save manager not initialized")
	}

	return asm.saveManager.LoadAutoSave()
}

// DeleteAutoSave deletes the auto-save
func (asm *AutoSaveManager) DeleteAutoSave() error {
	if asm.saveManager == nil {
		return fmt.Errorf("save manager not initialized")
	}

	return asm.saveManager.DeleteSave(AutoSaveSlot)
}

// GetAutoSaveInfo returns information about the auto-save
func (asm *AutoSaveManager) GetAutoSaveInfo() (string, error) {
	if asm.saveManager == nil {
		return "", fmt.Errorf("save manager not initialized")
	}

	return asm.saveManager.GetSaveSlotInfo(AutoSaveSlot)
}

// SetEnabled enables or disables auto-save
func (asm *AutoSaveManager) SetEnabled(enabled bool) {
	asm.enabled = enabled
	logger.Info("Auto-save enabled state changed", "enabled", enabled)
}

// IsEnabled returns whether auto-save is enabled
func (asm *AutoSaveManager) IsEnabled() bool {
	return asm.enabled
}

// SetSaveInterval sets the save interval in turns
func (asm *AutoSaveManager) SetSaveInterval(interval int) {
	if interval < 1 {
		interval = 1
	}
	asm.saveInterval = interval
	logger.Info("Auto-save interval changed", "interval", interval)
}

// GetSaveInterval returns the current save interval
func (asm *AutoSaveManager) GetSaveInterval() int {
	return asm.saveInterval
}

// SetSaveOnFloor sets whether to save when changing floors
func (asm *AutoSaveManager) SetSaveOnFloor(enabled bool) {
	asm.saveOnFloor = enabled
	logger.Info("Auto-save on floor change changed", "enabled", enabled)
}

// GetSaveOnFloor returns whether auto-save on floor change is enabled
func (asm *AutoSaveManager) GetSaveOnFloor() bool {
	return asm.saveOnFloor
}

// SetSaveOnDeath sets whether to save when player dies
func (asm *AutoSaveManager) SetSaveOnDeath(enabled bool) {
	asm.saveOnDeath = enabled
	logger.Info("Auto-save on death changed", "enabled", enabled)
}

// GetSaveOnDeath returns whether auto-save on death is enabled
func (asm *AutoSaveManager) GetSaveOnDeath() bool {
	return asm.saveOnDeath
}

// SetSaveOnVictory sets whether to save when player wins
func (asm *AutoSaveManager) SetSaveOnVictory(enabled bool) {
	asm.saveOnVictory = enabled
	logger.Info("Auto-save on victory changed", "enabled", enabled)
}

// GetSaveOnVictory returns whether auto-save on victory is enabled
func (asm *AutoSaveManager) GetSaveOnVictory() bool {
	return asm.saveOnVictory
}

// SetSaveOnExit sets whether to save when exiting game
func (asm *AutoSaveManager) SetSaveOnExit(enabled bool) {
	asm.saveOnExit = enabled
	logger.Info("Auto-save on exit changed", "enabled", enabled)
}

// GetSaveOnExit returns whether auto-save on exit is enabled
func (asm *AutoSaveManager) GetSaveOnExit() bool {
	return asm.saveOnExit
}

// GetLastSaveTurn returns the turn when the last auto-save was performed
func (asm *AutoSaveManager) GetLastSaveTurn() int {
	return asm.lastSaveTurn
}

// GetLastSaveTime returns the time when the last auto-save was performed
func (asm *AutoSaveManager) GetLastSaveTime() time.Time {
	return asm.lastSaveTime
}

// GetTimeSinceLastSave returns the time since the last auto-save
func (asm *AutoSaveManager) GetTimeSinceLastSave() time.Duration {
	return time.Since(asm.lastSaveTime)
}

// GetTurnsSinceLastSave returns the number of turns since the last auto-save
func (asm *AutoSaveManager) GetTurnsSinceLastSave(currentTurn int) int {
	return currentTurn - asm.lastSaveTurn
}

// GetNextAutoSaveTurn returns the turn when the next auto-save will occur
func (asm *AutoSaveManager) GetNextAutoSaveTurn() int {
	return asm.lastSaveTurn + asm.saveInterval
}

// GetAutoSaveProgress returns the progress towards the next auto-save (0.0 to 1.0)
func (asm *AutoSaveManager) GetAutoSaveProgress(currentTurn int) float64 {
	if asm.saveInterval == 0 {
		return 1.0
	}

	turnsSinceLastSave := currentTurn - asm.lastSaveTurn
	progress := float64(turnsSinceLastSave) / float64(asm.saveInterval)

	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}

// GetSettings returns the current auto-save settings
func (asm *AutoSaveManager) GetSettings() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         asm.enabled,
		"save_interval":   asm.saveInterval,
		"save_on_floor":   asm.saveOnFloor,
		"save_on_death":   asm.saveOnDeath,
		"save_on_victory": asm.saveOnVictory,
		"save_on_exit":    asm.saveOnExit,
		"last_save_turn":  asm.lastSaveTurn,
		"last_save_time":  asm.lastSaveTime,
	}
}

// SetSettings applies auto-save settings
func (asm *AutoSaveManager) SetSettings(settings map[string]interface{}) {
	if enabled, ok := settings["enabled"].(bool); ok {
		asm.enabled = enabled
	}

	if interval, ok := settings["save_interval"].(int); ok {
		asm.SetSaveInterval(interval)
	}

	if saveOnFloor, ok := settings["save_on_floor"].(bool); ok {
		asm.saveOnFloor = saveOnFloor
	}

	if saveOnDeath, ok := settings["save_on_death"].(bool); ok {
		asm.saveOnDeath = saveOnDeath
	}

	if saveOnVictory, ok := settings["save_on_victory"].(bool); ok {
		asm.saveOnVictory = saveOnVictory
	}

	if saveOnExit, ok := settings["save_on_exit"].(bool); ok {
		asm.saveOnExit = saveOnExit
	}

	logger.Info("Auto-save settings updated", "settings", asm.GetSettings())
}

// GetStatus returns the current auto-save status
func (asm *AutoSaveManager) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":              asm.enabled,
		"initialized":          asm.saveManager != nil,
		"has_auto_save":        asm.HasAutoSave(),
		"last_save_turn":       asm.lastSaveTurn,
		"last_save_time":       asm.lastSaveTime,
		"time_since_last_save": asm.GetTimeSinceLastSave(),
		"settings":             asm.GetSettings(),
	}

	if asm.HasAutoSave() {
		if info, err := asm.GetAutoSaveInfo(); err == nil {
			status["auto_save_info"] = info
		}
	}

	return status
}

// ValidateSettings validates auto-save settings
func (asm *AutoSaveManager) ValidateSettings() error {
	if asm.saveInterval < 1 {
		return fmt.Errorf("save interval must be at least 1 turn")
	}

	if asm.saveInterval > 10000 {
		return fmt.Errorf("save interval cannot exceed 10000 turns")
	}

	return nil
}

// Reset resets the auto-save manager state
func (asm *AutoSaveManager) Reset() {
	asm.lastSaveTurn = 0
	asm.lastSaveTime = time.Now()
	logger.Debug("Auto-save manager reset")
}

// ForceAutoSave forces an auto-save regardless of settings
func (asm *AutoSaveManager) ForceAutoSave(saveData *SaveData) error {
	if asm.saveManager == nil {
		return fmt.Errorf("save manager not initialized")
	}

	// Update tracking
	asm.lastSaveTurn = saveData.GameInfo.TurnCount
	asm.lastSaveTime = time.Now()

	// Force save
	if err := asm.saveManager.AutoSave(saveData); err != nil {
		return fmt.Errorf("forced auto-save failed: %w", err)
	}

	logger.Info("Forced auto-save completed")

	return nil
}

// GetRecommendedInterval returns a recommended save interval based on game state
func (asm *AutoSaveManager) GetRecommendedInterval(playerLevel int, currentFloor int) int {
	// Base interval
	interval := 100

	// Increase interval for higher level players (they're more experienced)
	if playerLevel > 10 {
		interval += (playerLevel - 10) * 10
	}

	// Decrease interval for deeper floors (more dangerous)
	if currentFloor > 15 {
		interval -= (currentFloor - 15) * 5
	}

	// Clamp to reasonable bounds
	if interval < 25 {
		interval = 25
	} else if interval > 500 {
		interval = 500
	}

	return interval
}

// GetStatistics returns auto-save statistics
func (asm *AutoSaveManager) GetStatistics() map[string]interface{} {
	timeSinceLastSave := asm.GetTimeSinceLastSave()

	return map[string]interface{}{
		"total_auto_saves":             asm.lastSaveTurn / asm.saveInterval, // Approximate
		"last_save_turn":               asm.lastSaveTurn,
		"save_interval":                asm.saveInterval,
		"time_since_last_save":         timeSinceLastSave,
		"time_since_last_save_minutes": timeSinceLastSave.Minutes(),
		"auto_save_enabled":            asm.enabled,
		"has_auto_save_file":           asm.HasAutoSave(),
	}
}

// Export exports auto-save settings for backup
func (asm *AutoSaveManager) Export() map[string]interface{} {
	return map[string]interface{}{
		"version":    "1.0",
		"timestamp":  time.Now().Unix(),
		"settings":   asm.GetSettings(),
		"statistics": asm.GetStatistics(),
		"status":     asm.GetStatus(),
	}
}

// Import imports auto-save settings from backup
func (asm *AutoSaveManager) Import(data map[string]interface{}) error {
	if settings, ok := data["settings"].(map[string]interface{}); ok {
		asm.SetSettings(settings)
	}

	// Note: We don't restore timestamps and turn counts as they're session-specific

	logger.Info("Auto-save settings imported")

	return nil
}
