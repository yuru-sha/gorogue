// Package score スコアファイル管理システム
// ~/.gorogue/scores.json にハイスコア情報を保存・管理する
package score

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	// ScoreFileName はスコアファイル名
	ScoreFileName = "scores.json"
	
	// MaxScoreEntries は保存する最大スコア数
	MaxScoreEntries = 100
	
	// ScoreVersion はスコアファイルのバージョン
	ScoreVersion = "1.0.0"
)

// ScoreEntry はスコア情報を表す構造体
type ScoreEntry struct {
	PlayerName     string    `json:"player_name"`
	Score          int       `json:"score"`
	Level          int       `json:"level"`
	DeepestFloor   int       `json:"deepest_floor"`
	PlayTime       int64     `json:"play_time"`        // 秒数
	TurnCount      int       `json:"turn_count"`
	MonstersKilled int       `json:"monsters_killed"`
	GoldCollected  int       `json:"gold_collected"`
	IsVictory      bool      `json:"is_victory"`
	DeathReason    string    `json:"death_reason,omitempty"`
	GameSeed       int64     `json:"game_seed"`
	Timestamp      time.Time `json:"timestamp"`
	Version        string    `json:"version"`
}

// ScoreFile はスコアファイルの構造体
type ScoreFile struct {
	Version string       `json:"version"`
	Updated time.Time    `json:"updated"`
	Entries []ScoreEntry `json:"entries"`
}

// ScoreManager はスコアファイルの管理を行う
type ScoreManager struct {
	scoreFilePath string
}

// NewScoreManager は新しいスコアマネージャーを作成する
func NewScoreManager() *ScoreManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	
	scoreDir := filepath.Join(homeDir, ".gorogue")
	scoreFilePath := filepath.Join(scoreDir, ScoreFileName)
	
	return &ScoreManager{
		scoreFilePath: scoreFilePath,
	}
}

// Initialize はスコアマネージャーを初期化する
func (sm *ScoreManager) Initialize() error {
	// スコアディレクトリの作成
	scoreDir := filepath.Dir(sm.scoreFilePath)
	if err := os.MkdirAll(scoreDir, 0755); err != nil {
		return fmt.Errorf("failed to create score directory: %w", err)
	}
	
	// スコアファイルが存在しない場合は作成
	if _, err := os.Stat(sm.scoreFilePath); os.IsNotExist(err) {
		if err := sm.createEmptyScoreFile(); err != nil {
			return fmt.Errorf("failed to create empty score file: %w", err)
		}
	}
	
	logger.Info("Score manager initialized", 
		"score_file", sm.scoreFilePath,
	)
	
	return nil
}

// createEmptyScoreFile は空のスコアファイルを作成する
func (sm *ScoreManager) createEmptyScoreFile() error {
	scoreFile := ScoreFile{
		Version: ScoreVersion,
		Updated: time.Now(),
		Entries: make([]ScoreEntry, 0),
	}
	
	return sm.writeScoreFile(&scoreFile)
}

// AddScore はスコアを追加する
func (sm *ScoreManager) AddScore(entry ScoreEntry) error {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return fmt.Errorf("failed to read score file: %w", err)
	}
	
	// タイムスタンプとバージョンを設定
	entry.Timestamp = time.Now()
	entry.Version = ScoreVersion
	
	// スコアを追加
	scoreFile.Entries = append(scoreFile.Entries, entry)
	
	// スコア順でソート（高い順）
	sort.Slice(scoreFile.Entries, func(i, j int) bool {
		return scoreFile.Entries[i].Score > scoreFile.Entries[j].Score
	})
	
	// 最大エントリ数を超えた場合は削除
	if len(scoreFile.Entries) > MaxScoreEntries {
		scoreFile.Entries = scoreFile.Entries[:MaxScoreEntries]
	}
	
	// ファイルの更新日時を設定
	scoreFile.Updated = time.Now()
	
	// ファイルに保存
	if err := sm.writeScoreFile(scoreFile); err != nil {
		return fmt.Errorf("failed to write score file: %w", err)
	}
	
	logger.Info("Score added successfully",
		"player", entry.PlayerName,
		"score", entry.Score,
		"level", entry.Level,
		"victory", entry.IsVictory,
	)
	
	return nil
}

// GetHighScores は上位スコアを取得する
func (sm *ScoreManager) GetHighScores(limit int) ([]ScoreEntry, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read score file: %w", err)
	}
	
	if limit <= 0 || limit > len(scoreFile.Entries) {
		limit = len(scoreFile.Entries)
	}
	
	return scoreFile.Entries[:limit], nil
}

// GetAllScores は全スコアを取得する
func (sm *ScoreManager) GetAllScores() ([]ScoreEntry, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read score file: %w", err)
	}
	
	return scoreFile.Entries, nil
}

// GetBestScore は最高スコアを取得する
func (sm *ScoreManager) GetBestScore() (*ScoreEntry, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read score file: %w", err)
	}
	
	if len(scoreFile.Entries) == 0 {
		return nil, nil
	}
	
	return &scoreFile.Entries[0], nil
}

// GetVictoryScores は勝利スコアのみを取得する
func (sm *ScoreManager) GetVictoryScores() ([]ScoreEntry, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read score file: %w", err)
	}
	
	var victories []ScoreEntry
	for _, entry := range scoreFile.Entries {
		if entry.IsVictory {
			victories = append(victories, entry)
		}
	}
	
	return victories, nil
}

// GetPlayerScores は特定プレイヤーのスコアを取得する
func (sm *ScoreManager) GetPlayerScores(playerName string) ([]ScoreEntry, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read score file: %w", err)
	}
	
	var playerScores []ScoreEntry
	for _, entry := range scoreFile.Entries {
		if entry.PlayerName == playerName {
			playerScores = append(playerScores, entry)
		}
	}
	
	return playerScores, nil
}

// ClearScores は全スコアを削除する
func (sm *ScoreManager) ClearScores() error {
	if err := sm.createEmptyScoreFile(); err != nil {
		return fmt.Errorf("failed to clear scores: %w", err)
	}
	
	logger.Info("All scores cleared")
	return nil
}

// GetScoreStats はスコア統計情報を取得する
func (sm *ScoreManager) GetScoreStats() (*ScoreStats, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read score file: %w", err)
	}
	
	stats := &ScoreStats{
		TotalEntries: len(scoreFile.Entries),
		VictoryCount: 0,
		HighestScore: 0,
		AverageScore: 0,
		DeepestFloor: 0,
		TotalPlayTime: 0,
		LastUpdated: scoreFile.Updated,
	}
	
	if len(scoreFile.Entries) == 0 {
		return stats, nil
	}
	
	totalScore := 0
	for _, entry := range scoreFile.Entries {
		if entry.IsVictory {
			stats.VictoryCount++
		}
		if entry.Score > stats.HighestScore {
			stats.HighestScore = entry.Score
		}
		if entry.DeepestFloor > stats.DeepestFloor {
			stats.DeepestFloor = entry.DeepestFloor
		}
		totalScore += entry.Score
		stats.TotalPlayTime += entry.PlayTime
	}
	
	stats.AverageScore = totalScore / len(scoreFile.Entries)
	
	return stats, nil
}

// ScoreStats はスコア統計情報
type ScoreStats struct {
	TotalEntries  int       `json:"total_entries"`
	VictoryCount  int       `json:"victory_count"`
	HighestScore  int       `json:"highest_score"`
	AverageScore  int       `json:"average_score"`
	DeepestFloor  int       `json:"deepest_floor"`
	TotalPlayTime int64     `json:"total_play_time"`
	LastUpdated   time.Time `json:"last_updated"`
}

// IsHighScore は指定されたスコアがハイスコアかどうかを判定する
func (sm *ScoreManager) IsHighScore(score int) (bool, int, error) {
	scoreFile, err := sm.readScoreFile()
	if err != nil {
		return false, 0, fmt.Errorf("failed to read score file: %w", err)
	}
	
	// エントリが最大数未満の場合は常にハイスコア
	if len(scoreFile.Entries) < MaxScoreEntries {
		return true, len(scoreFile.Entries) + 1, nil
	}
	
	// 最低スコアより高い場合はハイスコア
	lowestScore := scoreFile.Entries[len(scoreFile.Entries)-1].Score
	if score > lowestScore {
		// 順位を計算
		rank := len(scoreFile.Entries) + 1
		for i, entry := range scoreFile.Entries {
			if score > entry.Score {
				rank = i + 1
				break
			}
		}
		return true, rank, nil
	}
	
	return false, 0, nil
}

// GetScoreFilePath はスコアファイルのパスを取得する
func (sm *ScoreManager) GetScoreFilePath() string {
	return sm.scoreFilePath
}

// BackupScores はスコアファイルをバックアップする
func (sm *ScoreManager) BackupScores() error {
	backupPath := sm.scoreFilePath + ".backup"
	
	// 元ファイルを読み込み
	data, err := os.ReadFile(sm.scoreFilePath)
	if err != nil {
		return fmt.Errorf("failed to read score file: %w", err)
	}
	
	// バックアップファイルに書き込み
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	
	logger.Info("Scores backed up successfully", "backup_path", backupPath)
	return nil
}

// RestoreScores はバックアップからスコアファイルを復元する
func (sm *ScoreManager) RestoreScores() error {
	backupPath := sm.scoreFilePath + ".backup"
	
	// バックアップファイルを読み込み
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}
	
	// 元ファイルに書き込み
	if err := os.WriteFile(sm.scoreFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write score file: %w", err)
	}
	
	logger.Info("Scores restored successfully", "backup_path", backupPath)
	return nil
}

// プライベートメソッド

// readScoreFile はスコアファイルを読み込む
func (sm *ScoreManager) readScoreFile() (*ScoreFile, error) {
	data, err := os.ReadFile(sm.scoreFilePath)
	if err != nil {
		return nil, err
	}
	
	var scoreFile ScoreFile
	if err := json.Unmarshal(data, &scoreFile); err != nil {
		return nil, err
	}
	
	return &scoreFile, nil
}

// writeScoreFile はスコアファイルを書き込む
func (sm *ScoreManager) writeScoreFile(scoreFile *ScoreFile) error {
	data, err := json.MarshalIndent(scoreFile, "", "  ")
	if err != nil {
		return err
	}
	
	// 一時ファイルに書き込み
	tempFile := sm.scoreFilePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}
	
	// アトミックに移動
	if err := os.Rename(tempFile, sm.scoreFilePath); err != nil {
		os.Remove(tempFile)
		return err
	}
	
	return nil
}