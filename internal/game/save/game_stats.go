// Package save ゲーム統計管理機能
// プレイヤーの行動、成果、進行状況を追跡し、統計情報を提供
package save

import (
	"fmt"
	"time"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// GameStats manages game statistics and metrics
type GameStats struct {
	startTime time.Time
	stats     Stats
}

// NewGameStats creates a new game statistics manager
func NewGameStats() *GameStats {
	return &GameStats{
		startTime: time.Now(),
		stats:     Stats{},
	}
}

// Reset resets all statistics
func (gs *GameStats) Reset() {
	gs.startTime = time.Now()
	gs.stats = Stats{}
	logger.Debug("Game statistics reset")
}

// GetStats returns the current statistics
func (gs *GameStats) GetStats() Stats {
	return gs.stats
}

// LoadStats loads statistics from save data
func (gs *GameStats) LoadStats(stats Stats) {
	gs.stats = stats
	logger.Debug("Game statistics loaded",
		"monsters_killed", stats.MonstersKilled,
		"items_found", stats.ItemsFound,
		"gold_collected", stats.GoldCollected,
	)
}

// GetPlayTime returns the total play time in seconds
func (gs *GameStats) GetPlayTime() int64 {
	return int64(time.Since(gs.startTime).Seconds())
}

// GetTurnCount returns the current turn count
func (gs *GameStats) GetTurnCount() int {
	return gs.stats.TurnCount
}

// OnTurnEnd handles end of turn processing
func (gs *GameStats) OnTurnEnd() {
	gs.stats.TurnCount++
	gs.stats.StepsTaken++
}

// OnMonsterKilled handles monster death
func (gs *GameStats) OnMonsterKilled(monster *actor.Monster) {
	gs.stats.MonstersKilled++

	// Add to damage dealt (assuming monster's max HP as damage)
	gs.stats.DamageDealt += monster.MaxHP

	logger.Debug("Monster killed",
		"monster", monster.Type.Name,
		"total_killed", gs.stats.MonstersKilled,
	)
}

// OnItemFound handles item discovery
func (gs *GameStats) OnItemFound(itemName string) {
	gs.stats.ItemsFound++

	logger.Debug("Item found",
		"item", itemName,
		"total_found", gs.stats.ItemsFound,
	)
}

// OnItemUsed handles item usage
func (gs *GameStats) OnItemUsed(itemName string) {
	gs.stats.ItemsUsed++

	logger.Debug("Item used",
		"item", itemName,
		"total_used", gs.stats.ItemsUsed,
	)
}

// OnItemIdentified handles item identification
func (gs *GameStats) OnItemIdentified(itemName string) {
	gs.stats.ItemsIdentified++

	logger.Debug("Item identified",
		"item", itemName,
		"total_identified", gs.stats.ItemsIdentified,
	)
}

// OnDamageDealt handles damage dealt
func (gs *GameStats) OnDamageDealt(damage int) {
	gs.stats.DamageDealt += damage
}

// OnDamageTaken handles damage taken
func (gs *GameStats) OnDamageTaken(damage int) {
	gs.stats.DamageTaken += damage
}

// OnHealing handles healing
func (gs *GameStats) OnHealing(amount int) {
	gs.stats.TimesHealed++
}

// OnGoldCollected handles gold collection
func (gs *GameStats) OnGoldCollected(amount int) {
	gs.stats.GoldCollected += amount
}

// OnFloorChange handles floor change
func (gs *GameStats) OnFloorChange(newFloor int) {
	gs.stats.FloorsVisited++

	if newFloor > gs.stats.DeepestFloor {
		gs.stats.DeepestFloor = newFloor
	}

	logger.Debug("Floor changed",
		"floor", newFloor,
		"deepest", gs.stats.DeepestFloor,
	)
}

// OnRoomEntered handles room entry
func (gs *GameStats) OnRoomEntered() {
	gs.stats.RoomsEntered++
}

// OnTileExplored handles tile exploration
func (gs *GameStats) OnTileExplored() {
	gs.stats.TilesExplored++
}

// OnSecretFound handles secret discovery
func (gs *GameStats) OnSecretFound() {
	gs.stats.SecretsFound++
}

// OnTrapTriggered handles trap triggering
func (gs *GameStats) OnTrapTriggered() {
	gs.stats.TrapsTriggered++
}

// OnAmuletFound handles amulet discovery
func (gs *GameStats) OnAmuletFound() {
	gs.stats.AmuletFound = true
	logger.Info("Amulet of Yendor found!")
}

// OnPlayerVictory handles player victory
func (gs *GameStats) OnPlayerVictory() {
	gs.stats.EscapedWithAmulet = true
	logger.Info("Player achieved victory!")
}

// OnPlayerDeath handles player death
func (gs *GameStats) OnPlayerDeath(reason string, floor int) {
	gs.stats.DeathCount++
	gs.stats.LastDeathReason = reason
	gs.stats.LastDeathFloor = floor

	logger.Info("Player died",
		"reason", reason,
		"floor", floor,
		"death_count", gs.stats.DeathCount,
	)
}

// OnLevelUp handles level up
func (gs *GameStats) OnLevelUp(newLevel int) {
	if newLevel > gs.stats.HighestLevel {
		gs.stats.HighestLevel = newLevel
	}

	logger.Info("Player leveled up",
		"level", newLevel,
		"highest", gs.stats.HighestLevel,
	)
}

// GetSummary returns a formatted summary of statistics
func (gs *GameStats) GetSummary() map[string]interface{} {
	playTime := gs.GetPlayTime()

	return map[string]interface{}{
		"play_time_seconds":   playTime,
		"play_time_formatted": formatPlayTime(playTime),
		"turns_taken":         gs.stats.TurnCount,
		"steps_taken":         gs.stats.StepsTaken,
		"monsters_killed":     gs.stats.MonstersKilled,
		"damage_dealt":        gs.stats.DamageDealt,
		"damage_taken":        gs.stats.DamageTaken,
		"times_healed":        gs.stats.TimesHealed,
		"items_found":         gs.stats.ItemsFound,
		"items_used":          gs.stats.ItemsUsed,
		"items_identified":    gs.stats.ItemsIdentified,
		"gold_collected":      gs.stats.GoldCollected,
		"floors_visited":      gs.stats.FloorsVisited,
		"rooms_entered":       gs.stats.RoomsEntered,
		"tiles_explored":      gs.stats.TilesExplored,
		"secrets_found":       gs.stats.SecretsFound,
		"traps_triggered":     gs.stats.TrapsTriggered,
		"deepest_floor":       gs.stats.DeepestFloor,
		"highest_level":       gs.stats.HighestLevel,
		"amulet_found":        gs.stats.AmuletFound,
		"escaped_with_amulet": gs.stats.EscapedWithAmulet,
		"death_count":         gs.stats.DeathCount,
		"last_death_reason":   gs.stats.LastDeathReason,
		"last_death_floor":    gs.stats.LastDeathFloor,
	}
}

// GetEfficiencyMetrics returns efficiency metrics
func (gs *GameStats) GetEfficiencyMetrics() map[string]float64 {
	metrics := make(map[string]float64)

	if gs.stats.TurnCount > 0 {
		metrics["monsters_per_turn"] = float64(gs.stats.MonstersKilled) / float64(gs.stats.TurnCount)
		metrics["damage_per_turn"] = float64(gs.stats.DamageDealt) / float64(gs.stats.TurnCount)
		metrics["items_per_turn"] = float64(gs.stats.ItemsFound) / float64(gs.stats.TurnCount)
		metrics["gold_per_turn"] = float64(gs.stats.GoldCollected) / float64(gs.stats.TurnCount)
	}

	if gs.stats.MonstersKilled > 0 {
		metrics["avg_damage_per_monster"] = float64(gs.stats.DamageDealt) / float64(gs.stats.MonstersKilled)
	}

	if gs.stats.DamageDealt > 0 {
		metrics["damage_efficiency"] = float64(gs.stats.DamageDealt) / float64(gs.stats.DamageTaken)
	}

	if gs.stats.ItemsFound > 0 {
		metrics["identification_rate"] = float64(gs.stats.ItemsIdentified) / float64(gs.stats.ItemsFound)
		metrics["usage_rate"] = float64(gs.stats.ItemsUsed) / float64(gs.stats.ItemsFound)
	}

	playTime := gs.GetPlayTime()
	if playTime > 0 {
		metrics["turns_per_second"] = float64(gs.stats.TurnCount) / float64(playTime)
	}

	return metrics
}

// GetProgressMetrics returns progress metrics
func (gs *GameStats) GetProgressMetrics() map[string]float64 {
	metrics := make(map[string]float64)

	// Floor progress (0-100%)
	metrics["floor_progress"] = float64(gs.stats.DeepestFloor) / 26.0 * 100.0
	if metrics["floor_progress"] > 100.0 {
		metrics["floor_progress"] = 100.0
	}

	// Exploration efficiency
	if gs.stats.FloorsVisited > 0 {
		metrics["avg_rooms_per_floor"] = float64(gs.stats.RoomsEntered) / float64(gs.stats.FloorsVisited)
	}

	// Combat effectiveness
	if gs.stats.TurnCount > 0 {
		metrics["survival_rate"] = float64(gs.stats.TurnCount-gs.stats.DeathCount) / float64(gs.stats.TurnCount) * 100.0
	}

	// Completion metrics
	if gs.stats.AmuletFound {
		metrics["amulet_progress"] = 100.0
	} else {
		metrics["amulet_progress"] = float64(gs.stats.DeepestFloor) / 26.0 * 50.0 // 50% for reaching bottom
	}

	return metrics
}

// GetAchievements returns achieved milestones
func (gs *GameStats) GetAchievements() []string {
	achievements := make([]string, 0)

	// Combat achievements
	if gs.stats.MonstersKilled >= 100 {
		achievements = append(achievements, "Monster Slayer (100+ kills)")
	}
	if gs.stats.MonstersKilled >= 500 {
		achievements = append(achievements, "Monster Destroyer (500+ kills)")
	}

	// Exploration achievements
	if gs.stats.DeepestFloor >= 13 {
		achievements = append(achievements, "Deep Explorer (Floor 13+)")
	}
	if gs.stats.DeepestFloor >= 26 {
		achievements = append(achievements, "Dungeon Master (Floor 26)")
	}

	// Item achievements
	if gs.stats.ItemsFound >= 50 {
		achievements = append(achievements, "Treasure Hunter (50+ items)")
	}
	if gs.stats.ItemsIdentified >= 25 {
		achievements = append(achievements, "Sage (25+ identifications)")
	}

	// Gold achievements
	if gs.stats.GoldCollected >= 1000 {
		achievements = append(achievements, "Rich Adventurer (1000+ gold)")
	}
	if gs.stats.GoldCollected >= 5000 {
		achievements = append(achievements, "Wealthy Noble (5000+ gold)")
	}

	// Special achievements
	if gs.stats.AmuletFound {
		achievements = append(achievements, "Amulet Finder")
	}
	if gs.stats.EscapedWithAmulet {
		achievements = append(achievements, "Rogue Victor")
	}

	// Survival achievements
	if gs.stats.DeathCount == 0 && gs.stats.TurnCount > 1000 {
		achievements = append(achievements, "Survivor (1000+ turns, no deaths)")
	}

	// Speed achievements
	playTime := gs.GetPlayTime()
	if gs.stats.EscapedWithAmulet && playTime < 3600 { // 1 hour
		achievements = append(achievements, "Speed Runner (Victory in <1 hour)")
	}

	return achievements
}

// formatPlayTime formats play time in a human-readable format
func formatPlayTime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		seconds = seconds % 60
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		seconds = seconds % 60
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
}

// GetDetailedReport returns a detailed statistics report
func (gs *GameStats) GetDetailedReport() map[string]interface{} {
	return map[string]interface{}{
		"summary":      gs.GetSummary(),
		"efficiency":   gs.GetEfficiencyMetrics(),
		"progress":     gs.GetProgressMetrics(),
		"achievements": gs.GetAchievements(),
	}
}

// IsHighScore checks if current stats represent a high score
func (gs *GameStats) IsHighScore() bool {
	// Simple high score criteria
	if gs.stats.EscapedWithAmulet {
		return true
	}

	if gs.stats.DeepestFloor >= 20 {
		return true
	}

	if gs.stats.MonstersKilled >= 200 {
		return true
	}

	if gs.stats.GoldCollected >= 3000 {
		return true
	}

	return false
}

// GetScoreValue returns a numerical score value
func (gs *GameStats) GetScoreValue() int {
	score := 0

	// Base score from progress
	score += gs.stats.DeepestFloor * 100

	// Bonus for completing
	if gs.stats.EscapedWithAmulet {
		score += 10000
	} else if gs.stats.AmuletFound {
		score += 5000
	}

	// Combat score
	score += gs.stats.MonstersKilled * 10
	score += gs.stats.DamageDealt

	// Exploration score
	score += gs.stats.ItemsFound * 5
	score += gs.stats.GoldCollected

	// Efficiency bonuses
	if gs.stats.DeathCount == 0 {
		score += 1000
	}

	// Time penalty (longer time = lower score)
	playTime := gs.GetPlayTime()
	if playTime > 0 {
		timePenalty := int(playTime / 10) // 1 point per 10 seconds
		score -= timePenalty
	}

	if score < 0 {
		score = 0
	}

	return score
}

// Export returns statistics in exportable format
func (gs *GameStats) Export() map[string]interface{} {
	return map[string]interface{}{
		"version":      "1.0",
		"timestamp":    time.Now().Unix(),
		"play_time":    gs.GetPlayTime(),
		"stats":        gs.stats,
		"summary":      gs.GetSummary(),
		"efficiency":   gs.GetEfficiencyMetrics(),
		"progress":     gs.GetProgressMetrics(),
		"achievements": gs.GetAchievements(),
		"score":        gs.GetScoreValue(),
	}
}
