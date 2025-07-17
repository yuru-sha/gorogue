// Package config 環境変数を管理し、設定値を提供する
// .envファイルから環境変数を読み込み、型安全なアクセスを提供する
package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// デフォルト値の定数
const (
	DefaultDebugMode       = false
	DefaultLogLevel        = "INFO"
	DefaultSaveDirectory   = "saves"
	DefaultAutoSaveEnabled = true
	DefaultSaveInterval    = 100 // ターン数
	DefaultWindowWidth     = 800
	DefaultWindowHeight    = 600
	DefaultFontSize        = 16
	DefaultShowTips        = true
	DefaultConfirmQuit     = true
	DefaultAutoPickup      = true
	DefaultWizardMode      = false
)

// 環境変数のキー名
const (
	EnvDebugMode       = "DEBUG"
	EnvLogLevel        = "LOG_LEVEL"
	EnvSaveDirectory   = "SAVE_DIRECTORY"
	EnvAutoSaveEnabled = "AUTO_SAVE_ENABLED"
	EnvSaveInterval    = "SAVE_INTERVAL"
	EnvWindowWidth     = "WINDOW_WIDTH"
	EnvWindowHeight    = "WINDOW_HEIGHT"
	EnvFontSize        = "FONT_SIZE"
	EnvShowTips        = "SHOW_TIPS"
	EnvConfirmQuit     = "CONFIRM_QUIT"
	EnvAutoPickup      = "AUTO_PICKUP"
	EnvWizardMode      = "WIZARD_MODE"
)

// 初期化時に.envファイルを読み込む
func init() {
	LoadEnv()
}

// LoadEnv は .env ファイルを読み込む
func LoadEnv() error {
	// .env ファイルが存在しない場合はエラーにしない
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: failed to load .env file: %v", err)
		}
	}
	return nil
}

// GetBool は環境変数を真偽値として読み込む
func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// true/false, 1/0, on/off, yes/no をサポート
	switch strings.ToLower(value) {
	case "true", "1", "on", "yes":
		return true
	case "false", "0", "off", "no":
		return false
	default:
		return defaultValue
	}
}

// GetString は環境変数を文字列として読み込む
func GetString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt は環境変数を整数として読み込む
func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

// GetFloat64 は環境変数を浮動小数点数として読み込む
func GetFloat64(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return defaultValue
}

// 設定値へのアクセサー関数

// GetDebugMode はデバッグモードの設定を取得する
func GetDebugMode() bool {
	return GetBool(EnvDebugMode, DefaultDebugMode)
}

// GetLogLevel はログレベルの設定を取得する
func GetLogLevel() string {
	return GetString(EnvLogLevel, DefaultLogLevel)
}

// GetSaveDirectory はセーブディレクトリの設定を取得する
func GetSaveDirectory() string {
	return GetString(EnvSaveDirectory, DefaultSaveDirectory)
}

// GetAutoSaveEnabled はオートセーブの設定を取得する
func GetAutoSaveEnabled() bool {
	return GetBool(EnvAutoSaveEnabled, DefaultAutoSaveEnabled)
}

// GetSaveInterval はセーブ間隔の設定を取得する
func GetSaveInterval() int {
	return GetInt(EnvSaveInterval, DefaultSaveInterval)
}

// GetWindowWidth はウィンドウ幅の設定を取得する
func GetWindowWidth() int {
	return GetInt(EnvWindowWidth, DefaultWindowWidth)
}

// GetWindowHeight はウィンドウ高さの設定を取得する
func GetWindowHeight() int {
	return GetInt(EnvWindowHeight, DefaultWindowHeight)
}

// GetFontSize はフォントサイズの設定を取得する
func GetFontSize() int {
	return GetInt(EnvFontSize, DefaultFontSize)
}

// GetShowTips はヒント表示の設定を取得する
func GetShowTips() bool {
	return GetBool(EnvShowTips, DefaultShowTips)
}

// GetConfirmQuit は終了確認の設定を取得する
func GetConfirmQuit() bool {
	return GetBool(EnvConfirmQuit, DefaultConfirmQuit)
}

// GetAutoPickup は自動拾得の設定を取得する
func GetAutoPickup() bool {
	return GetBool(EnvAutoPickup, DefaultAutoPickup)
}

// GetWizardMode はウィザードモードの設定を取得する
func GetWizardMode() bool {
	return GetBool(EnvWizardMode, DefaultWizardMode)
}

// Config 設定値を構造体として提供する
type Config struct {
	DebugMode       bool   `json:"debug_mode"`
	LogLevel        string `json:"log_level"`
	SaveDirectory   string `json:"save_directory"`
	AutoSaveEnabled bool   `json:"auto_save_enabled"`
	SaveInterval    int    `json:"save_interval"`
	WindowWidth     int    `json:"window_width"`
	WindowHeight    int    `json:"window_height"`
	FontSize        int    `json:"font_size"`
	ShowTips        bool   `json:"show_tips"`
	ConfirmQuit     bool   `json:"confirm_quit"`
	AutoPickup      bool   `json:"auto_pickup"`
	WizardMode      bool   `json:"wizard_mode"`
}

// GetConfig は現在の設定を構造体として取得する
func GetConfig() *Config {
	return &Config{
		DebugMode:       GetDebugMode(),
		LogLevel:        GetLogLevel(),
		SaveDirectory:   GetSaveDirectory(),
		AutoSaveEnabled: GetAutoSaveEnabled(),
		SaveInterval:    GetSaveInterval(),
		WindowWidth:     GetWindowWidth(),
		WindowHeight:    GetWindowHeight(),
		FontSize:        GetFontSize(),
		ShowTips:        GetShowTips(),
		ConfirmQuit:     GetConfirmQuit(),
		AutoPickup:      GetAutoPickup(),
		WizardMode:      GetWizardMode(),
	}
}

// SetEnv は環境変数を設定する（テスト用）
func SetEnv(key, value string) {
	os.Setenv(key, value)
}

// UnsetEnv は環境変数を削除する（テスト用）
func UnsetEnv(key string) {
	os.Unsetenv(key)
}

// ReloadEnv は .env ファイルを再読み込みする
func ReloadEnv() error {
	return LoadEnv()
}

// PrintConfig は現在の設定を出力する（デバッグ用）
func PrintConfig() {
	config := GetConfig()
	log.Printf("Current configuration:")
	log.Printf("  DebugMode: %v", config.DebugMode)
	log.Printf("  LogLevel: %s", config.LogLevel)
	log.Printf("  SaveDirectory: %s", config.SaveDirectory)
	log.Printf("  AutoSaveEnabled: %v", config.AutoSaveEnabled)
	log.Printf("  SaveInterval: %d", config.SaveInterval)
	log.Printf("  WindowWidth: %d", config.WindowWidth)
	log.Printf("  WindowHeight: %d", config.WindowHeight)
	log.Printf("  FontSize: %d", config.FontSize)
	log.Printf("  ShowTips: %v", config.ShowTips)
	log.Printf("  ConfirmQuit: %v", config.ConfirmQuit)
	log.Printf("  AutoPickup: %v", config.AutoPickup)
	log.Printf("  WizardMode: %v", config.WizardMode)
}