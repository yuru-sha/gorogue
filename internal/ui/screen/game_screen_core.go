// Package screen ゲーム画面の描画と入力処理を提供
// Gruidライブラリを使用したローグライクゲームのUI管理
package screen

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/wizard"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// InputMode represents the current input mode
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeEquip
	ModeUnequip
	ModeDrop
	ModeQuaff
	ModeRead
	ModeCLI
)

// GameScreen handles the main game display
type GameScreen struct {
	width, height   int
	player          *actor.Player
	level           *dungeon.Level
	dungeonManager  *dungeon.DungeonManager
	messages        []string
	lastStats       map[string]interface{} // 前回のステータス情報
	grid            gruid.Grid             // 画面全体のグリッド
	wizardMode      *wizard.WizardMode     // ウィザードモード
	cliMode         *cli.CLIMode           // CLIデバッグモード
	inputMode       InputMode              // 現在の入力モード
	equippableItems []*gameitem.Item       // 装備可能アイテムリスト
	cliBuffer       string                 // CLI入力バッファ
	cliHistory      []string               // CLIコマンド履歴
	cmdParser       *command.Parser        // Command parser
}

// NewGameScreen creates a new game screen
func NewGameScreen(width, height int, player *actor.Player) *GameScreen {
	screen := &GameScreen{
		width:           width,
		height:          height,
		player:          player,
		messages:        make([]string, 0, 7), // 7行分のメッセージを保持
		lastStats:       make(map[string]interface{}),
		grid:            gruid.NewGrid(width, height),
		inputMode:       ModeNormal,
		equippableItems: make([]*gameitem.Item, 0),
		cliBuffer:       "",
		cliHistory:      make([]string, 0),
		cmdParser:       command.NewParser(),
	}

	// PyRogue風の初期メッセージを追加
	screen.AddMessage("Welcome to PyRogue!")
	screen.AddMessage("Use vi keys (hjkl), arrow keys, or numpad (1-9) to move.")
	screen.AddMessage("You are a skilled warrior.")
	screen.AddMessage("You are equipped with a dagger and leather armor.")
	screen.AddMessage("You start with no rings, potions, scrolls, food, and a scroll.")
	screen.AddMessage("You see a lit room.")
	screen.AddMessage("You enter the dungeon. Your quest begins!")

	logger.Debug("Created game screen",
		"width", width,
		"height", height,
	)
	return screen
}

// SetLevel sets the dungeon level for the game screen
func (s *GameScreen) SetLevel(level *dungeon.Level) {
	s.level = level
	s.wizardMode = wizard.NewWizardMode(level, s.player)
	s.cliMode = cli.NewCLIMode(level, s.player)
	logger.Debug("Set dungeon level for game screen",
		"width", level.Width,
		"height", level.Height,
	)
}

// SetDungeonManager sets the dungeon manager for the game screen
func (s *GameScreen) SetDungeonManager(dm *dungeon.DungeonManager) {
	s.dungeonManager = dm
	logger.Debug("Set dungeon manager for game screen")
}

// AddMessage adds a message to the message log
func (s *GameScreen) AddMessage(msg string) {
	s.messages = append(s.messages, msg)
	if len(s.messages) > 7 {
		s.messages = s.messages[len(s.messages)-7:]
	}
	logger.Debug("Added message to log",
		"message", msg,
		"messages_count", len(s.messages),
	)
}
