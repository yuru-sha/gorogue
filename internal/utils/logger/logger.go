package logger

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	maxGenerations = 5 // Number of log generations to keep
)

var (
	// ログファイルのパス
	logDir    = "logs"
	gameLog   = filepath.Join(logDir, "game.log")
	errorLog  = filepath.Join(logDir, "error.log")
	gameFile  *os.File
	errorFile *os.File

	// slogロガーのインスタンス
	gameLogger  *slog.Logger
	errorLogger *slog.Logger
)

// rotateLogFile rotates log files, keeping maxGenerations
func rotateLogFile(basePath string) error {
	// Remove the oldest log file
	oldestLog := fmt.Sprintf("%s.%d", basePath, maxGenerations)
	os.Remove(oldestLog)

	// Rotate existing log files
	for i := maxGenerations - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", basePath, i)
		newPath := fmt.Sprintf("%s.%d", basePath, i+1)
		// ファイルローテーションのエラーは致命的でないため、ログ出力のみ
		if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Failed to rotate log file %s: %v", oldPath, err)
		}
	}

	// Rotate current log file
	if _, err := os.Stat(basePath); err == nil {
		newPath := fmt.Sprintf("%s.1", basePath)
		if err := os.Rename(basePath, newPath); err != nil {
			return fmt.Errorf("failed to rotate log file: %v", err)
		}
	}

	return nil
}

// Setup initializes the logger
func Setup() error {
	// ログディレクトリの作成
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Rotate log files
	if err := rotateLogFile(gameLog); err != nil {
		return err
	}
	if err := rotateLogFile(errorLog); err != nil {
		return err
	}

	// ゲームログファイルの作成
	var err error
	gameFile, err = os.OpenFile(gameLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create game log file: %v", err)
	}

	// エラーログファイルの作成
	errorFile, err = os.OpenFile(errorLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create error log file: %v", err)
	}

	// JSONハンドラーの作成
	gameOpts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	errorOpts := &slog.HandlerOptions{
		Level:     slog.LevelError,
		AddSource: true,
	}

	gameLogger = slog.New(slog.NewJSONHandler(gameFile, gameOpts))
	errorLogger = slog.New(slog.NewJSONHandler(errorFile, errorOpts))

	// 開始ログの出力
	gameLogger.Info("Game Started",
		"start_time", time.Now().Format(time.RFC3339),
		"go_version", runtime.Version(),
		"platform", runtime.GOOS,
		"architecture", runtime.GOARCH,
		"num_cpu", runtime.NumCPU(),
		"max_procs", runtime.GOMAXPROCS(0),
		"process_id", os.Getpid(),
	)

	return nil
}

// Cleanup closes the log files
func Cleanup() {
	if gameFile != nil {
		gameFile.Close()
	}
	if errorFile != nil {
		errorFile.Close()
	}
}

// addCallerInfo adds caller information to the log attributes
func addCallerInfo(attrs []any) []any {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return attrs
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return attrs
	}

	return append(attrs,
		"module", filepath.Base(file),
		"function", fn.Name(),
		"line", line,
	)
}

// Trace logs a trace message
func Trace(msg string, attrs ...any) {
	gameLogger.Debug(msg, addCallerInfo(attrs)...)
}

// Debug logs a debug message
func Debug(msg string, attrs ...any) {
	gameLogger.Debug(msg, addCallerInfo(attrs)...)
}

// Info logs an info message
func Info(msg string, attrs ...any) {
	gameLogger.Info(msg, addCallerInfo(attrs)...)
}

// Warn logs a warning message
func Warn(msg string, attrs ...any) {
	gameLogger.Warn(msg, addCallerInfo(attrs)...)
}

// Error logs an error message
func Error(msg string, attrs ...any) {
	errorLogger.Error(msg, addCallerInfo(attrs)...)
}

// Fatal logs a fatal error message and exits
func Fatal(msg string, attrs ...any) {
	errorLogger.Error(msg, addCallerInfo(attrs)...)
	os.Exit(1)
}
