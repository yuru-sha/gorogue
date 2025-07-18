// Package score スコア計算システム
// ゲームの各要素からスコアを計算する
package score

import (
	"math"
	"time"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/save"
)

// ScoreCalculator はスコア計算を行う
type ScoreCalculator struct {
	// 基本スコア倍率
	baseScoreMultiplier float64
	
	// 時間ペナルティ設定
	timepenaltyEnabled bool
	timePenaltyFactor  float64
	
	// 勝利ボーナス
	victoryBonus int
	
	// 階層ボーナス
	floorBonus int
	
	// モンスター討伐ボーナス
	monsterKillBonus int
	
	// ゴールドボーナス
	goldBonus int
	
	// レベルボーナス
	levelBonus int
	
	// 生存ボーナス
	survivalBonus int
}

// NewScoreCalculator は新しいスコア計算機を作成する
func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{
		baseScoreMultiplier: 1.0,
		timepenaltyEnabled:  true,
		timePenaltyFactor:   0.1,    // 10秒ごとに1ポイント減点
		victoryBonus:        10000,  // 勝利ボーナス
		floorBonus:          100,    // 階層ごとのボーナス
		monsterKillBonus:    10,     // モンスター討伐ボーナス
		goldBonus:           1,      // ゴールドボーナス
		levelBonus:          500,    // レベルボーナス
		survivalBonus:       1000,   // 生存ボーナス
	}
}

// CalculateScore はゲーム統計からスコアを計算する
func (sc *ScoreCalculator) CalculateScore(player *actor.Player, stats *save.Stats, playTime int64, isVictory bool) int {
	score := 0.0
	
	// 基本スコア計算
	score += sc.calculateBaseScore(player, stats)
	
	// 勝利ボーナス
	if isVictory {
		score += float64(sc.victoryBonus)
	}
	
	// 階層ボーナス
	score += float64(stats.DeepestFloor * sc.floorBonus)
	
	// モンスター討伐ボーナス
	score += float64(stats.MonstersKilled * sc.monsterKillBonus)
	
	// ゴールドボーナス
	score += float64(stats.GoldCollected * sc.goldBonus)
	
	// レベルボーナス
	score += float64(player.Level * sc.levelBonus)
	
	// 生存ボーナス
	if stats.DeathCount == 0 {
		score += float64(sc.survivalBonus)
	}
	
	// 効率ボーナス
	score += sc.calculateEfficiencyBonus(stats, playTime)
	
	// 時間ペナルティ
	if sc.timepenaltyEnabled {
		score -= sc.calculateTimePenalty(playTime)
	}
	
	// 基本倍率を適用
	score *= sc.baseScoreMultiplier
	
	// 負の値にならないようにする
	if score < 0 {
		score = 0
	}
	
	return int(score)
}

// calculateBaseScore は基本スコアを計算する
func (sc *ScoreCalculator) calculateBaseScore(player *actor.Player, stats *save.Stats) float64 {
	baseScore := 0.0
	
	// プレイヤーレベルによる基本スコア
	baseScore += float64(player.Level * 100)
	
	// 経験値による基本スコア
	baseScore += float64(player.Exp)
	
	// 装備品による基本スコア
	if player.Equipment.Weapon != nil {
		baseScore += float64(player.Equipment.Weapon.Value)
	}
	if player.Equipment.Armor != nil {
		baseScore += float64(player.Equipment.Armor.Value)
	}
	if player.Equipment.RingLeft != nil {
		baseScore += float64(player.Equipment.RingLeft.Value)
	}
	if player.Equipment.RingRight != nil {
		baseScore += float64(player.Equipment.RingRight.Value)
	}
	
	// インベントリの価値
	for _, item := range player.Inventory.Items {
		if item != nil {
			baseScore += float64(item.Value * item.Quantity)
		}
	}
	
	return baseScore
}

// calculateEfficiencyBonus は効率ボーナスを計算する
func (sc *ScoreCalculator) calculateEfficiencyBonus(stats *save.Stats, playTime int64) float64 {
	bonus := 0.0
	
	// ターン効率ボーナス
	if stats.TurnCount > 0 {
		monstersPerTurn := float64(stats.MonstersKilled) / float64(stats.TurnCount)
		bonus += monstersPerTurn * 1000 // 1ターンあたりのモンスター討伐数
		
		goldPerTurn := float64(stats.GoldCollected) / float64(stats.TurnCount)
		bonus += goldPerTurn * 100 // 1ターンあたりのゴールド収集数
	}
	
	// 時間効率ボーナス
	if playTime > 0 {
		turnsPerSecond := float64(stats.TurnCount) / float64(playTime)
		bonus += turnsPerSecond * 500 // 1秒あたりのターン数
	}
	
	// 探索効率ボーナス
	if stats.FloorsVisited > 0 {
		roomsPerFloor := float64(stats.RoomsEntered) / float64(stats.FloorsVisited)
		bonus += roomsPerFloor * 50 // 1階あたりの部屋入室数
	}
	
	// 戦闘効率ボーナス
	if stats.DamageTaken > 0 {
		damageRatio := float64(stats.DamageDealt) / float64(stats.DamageTaken)
		bonus += damageRatio * 100 // ダメージ効率
	}
	
	return bonus
}

// calculateTimePenalty は時間ペナルティを計算する
func (sc *ScoreCalculator) calculateTimePenalty(playTime int64) float64 {
	// 10秒ごとに1ポイント減点
	penalty := float64(playTime) * sc.timePenaltyFactor
	
	// 最大ペナルティを設定（全スコアの50%まで）
	maxPenalty := 5000.0
	if penalty > maxPenalty {
		penalty = maxPenalty
	}
	
	return penalty
}

// CalculateScoreWithBreakdown はスコアの詳細内訳を計算する
func (sc *ScoreCalculator) CalculateScoreWithBreakdown(player *actor.Player, stats *save.Stats, playTime int64, isVictory bool) *ScoreBreakdown {
	breakdown := &ScoreBreakdown{
		BaseScore:        int(sc.calculateBaseScore(player, stats)),
		FloorBonus:       stats.DeepestFloor * sc.floorBonus,
		MonsterKillBonus: stats.MonstersKilled * sc.monsterKillBonus,
		GoldBonus:        stats.GoldCollected * sc.goldBonus,
		LevelBonus:       player.Level * sc.levelBonus,
		EfficiencyBonus:  int(sc.calculateEfficiencyBonus(stats, playTime)),
		TimePenalty:      int(sc.calculateTimePenalty(playTime)),
		PlayTime:         playTime,
		IsVictory:        isVictory,
	}
	
	if isVictory {
		breakdown.VictoryBonus = sc.victoryBonus
	}
	
	if stats.DeathCount == 0 {
		breakdown.SurvivalBonus = sc.survivalBonus
	}
	
	// 総合スコア計算
	totalScore := breakdown.BaseScore + breakdown.VictoryBonus + breakdown.FloorBonus + 
		breakdown.MonsterKillBonus + breakdown.GoldBonus + breakdown.LevelBonus + 
		breakdown.SurvivalBonus + breakdown.EfficiencyBonus - breakdown.TimePenalty
	
	if totalScore < 0 {
		totalScore = 0
	}
	
	breakdown.TotalScore = totalScore
	
	return breakdown
}

// ScoreBreakdown はスコアの詳細内訳
type ScoreBreakdown struct {
	BaseScore        int   `json:"base_score"`
	VictoryBonus     int   `json:"victory_bonus"`
	FloorBonus       int   `json:"floor_bonus"`
	MonsterKillBonus int   `json:"monster_kill_bonus"`
	GoldBonus        int   `json:"gold_bonus"`
	LevelBonus       int   `json:"level_bonus"`
	SurvivalBonus    int   `json:"survival_bonus"`
	EfficiencyBonus  int   `json:"efficiency_bonus"`
	TimePenalty      int   `json:"time_penalty"`
	TotalScore       int   `json:"total_score"`
	PlayTime         int64 `json:"play_time"`
	IsVictory        bool  `json:"is_victory"`
}

// GetScoreGrade はスコアに基づいてグレードを返す
func (sc *ScoreCalculator) GetScoreGrade(score int) string {
	switch {
	case score >= 50000:
		return "S+"
	case score >= 40000:
		return "S"
	case score >= 30000:
		return "A+"
	case score >= 25000:
		return "A"
	case score >= 20000:
		return "B+"
	case score >= 15000:
		return "B"
	case score >= 10000:
		return "C+"
	case score >= 5000:
		return "C"
	case score >= 1000:
		return "D"
	default:
		return "F"
	}
}

// GetScoreRank はスコアの順位を文字列で返す
func (sc *ScoreCalculator) GetScoreRank(rank int) string {
	switch rank {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", rank)
	}
}

// CreateScoreEntry はゲーム終了時にスコアエントリを作成する
func (sc *ScoreCalculator) CreateScoreEntry(playerName string, player *actor.Player, stats *save.Stats, gameInfo *save.GameInfo, isVictory bool, deathReason string) ScoreEntry {
	playTime := gameInfo.PlayTime
	score := sc.CalculateScore(player, stats, playTime, isVictory)
	
	return ScoreEntry{
		PlayerName:     playerName,
		Score:          score,
		Level:          player.Level,
		DeepestFloor:   stats.DeepestFloor,
		PlayTime:       playTime,
		TurnCount:      stats.TurnCount,
		MonstersKilled: stats.MonstersKilled,
		GoldCollected:  stats.GoldCollected,
		IsVictory:      isVictory,
		DeathReason:    deathReason,
		GameSeed:       gameInfo.Seed,
		Timestamp:      time.Now(),
		Version:        ScoreVersion,
	}
}

// SetScoreMultiplier はスコア倍率を設定する
func (sc *ScoreCalculator) SetScoreMultiplier(multiplier float64) {
	sc.baseScoreMultiplier = multiplier
}

// SetTimePenaltyEnabled は時間ペナルティの有効/無効を設定する
func (sc *ScoreCalculator) SetTimePenaltyEnabled(enabled bool) {
	sc.timepenaltyEnabled = enabled
}

// GetScoreMultiplier はスコア倍率を取得する
func (sc *ScoreCalculator) GetScoreMultiplier() float64 {
	return sc.baseScoreMultiplier
}

// IsTimePenaltyEnabled は時間ペナルティが有効かどうかを返す
func (sc *ScoreCalculator) IsTimePenaltyEnabled() bool {
	return sc.timepenaltyEnabled
}

// Helper functions

// FormatPlayTime はプレイ時間を見やすい形式でフォーマットする
func FormatPlayTime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		secs := seconds % 60
		return fmt.Sprintf("%d分%d秒", minutes, secs)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		secs := seconds % 60
		return fmt.Sprintf("%d時間%d分%d秒", hours, minutes, secs)
	}
}

// FormatScore はスコアを見やすい形式でフォーマットする
func FormatScore(score int) string {
	if score >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(score)/1000000)
	} else if score >= 1000 {
		return fmt.Sprintf("%.1fK", float64(score)/1000)
	}
	return fmt.Sprintf("%d", score)
}