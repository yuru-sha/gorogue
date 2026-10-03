// Package screen ゲーム画面の描画と入力処理を提供
// Gruidライブラリを使用したローグライクゲームのUI管理
package screen

import (
	"strings"

	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/wizard"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// InputMode represents the current input mode
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeEquip
	ModeUnequip
	ModeDrop
	ModeUse
	ModeQuaff
	ModeRead
	ModeReadTarget
	ModeEat
	ModeCLI
	ModeDirection
	ModeThrow
	ModeZap
	ModeCall
)

// GameScreen handles the main game display
type GameScreen struct {
	width, height     int
	player            *actor.Player
	level             *dungeon.Level
	displayCells      []displayCell
	dungeonManager    *dungeon.DungeonManager
	messages          []string
	lastStats         map[string]any
	wizardMode        *wizard.WizardMode
	cliMode           *cli.CLIMode
	saveIntegration   *save.SaveGameIntegration
	engineLoad        func() error
	inputMode         InputMode
	equippableItems   []*gameitem.Item
	cliBuffer         string
	cliHistory        []string
	cmdParser         *command.Parser
	commandSession    *command.Session
	directionCallback func(dx, dy int)
	pendingCommand    command.Type
	pendingDirection  command.Direction
	equipAction       command.Type
	unequipAction     command.Type
	callItemLetter    rune
	pendingReadScroll rune
	callNameBuffer    string
	presentation      presentationMode
	helpStage         helpStage
	helpPage          int
	sequencePages     [][]string
	sequenceStage     int
}

type presentationMode uint8

const (
	presentationPlay presentationMode = iota
	presentationHelp
	presentationDeath
	presentationVictory
)

type helpStage uint8

const (
	helpAwaitingKey helpStage = iota
	helpShowingList
)

// NewGameScreen creates a new game screen
func NewGameScreen(width, height int, player *actor.Player) *GameScreen {
	screen := &GameScreen{
		width:           width,
		height:          height,
		player:          player,
		messages:        make([]string, 0, 7), // 7行分のメッセージを保持
		lastStats:       make(map[string]any),
		inputMode:       ModeNormal,
		equippableItems: make([]*gameitem.Item, 0),
		cliBuffer:       "",
		cliHistory:      make([]string, 0),
		cmdParser:       command.NewParser(),
		commandSession:  command.NewSession(),
	}

	logger.Debug("Created game screen",
		"width", width,
		"height", height,
	)
	return screen
}

// SetLevel sets the dungeon level for the game screen
func (s *GameScreen) SetLevel(level *dungeon.Level) {
	s.level = level
	level.UpdateVisibility(s.player.Position.X, s.player.Position.Y)
	s.wizardMode = wizard.NewWizardMode(level, s.player)
	if s.dungeonManager != nil {
		s.cliMode = cli.NewCLIModeWithDungeonManager(s.dungeonManager, s.player)
	} else {
		s.cliMode = cli.NewCLIMode(level, s.player)
	}
	s.cliMode.SetSaveIntegration(s.saveIntegration)
	s.cliMode.SetCommandSession(s.commandSession)
	logger.Debug("Set dungeon level for game screen",
		"width", level.Width,
		"height", level.Height,
	)
}

// SetDungeonManager sets the dungeon manager for the game screen
func (s *GameScreen) SetDungeonManager(dm *dungeon.DungeonManager) {
	s.dungeonManager = dm
	if dm != nil {
		s.level = dm.GetCurrentLevel()
		s.level.UpdateVisibility(s.player.Position.X, s.player.Position.Y)
		if s.wizardMode != nil {
			s.wizardMode.SetLevel(s.level)
		}
		if s.cliMode == nil {
			s.cliMode = cli.NewCLIModeWithDungeonManager(dm, s.player)
		} else {
			s.cliMode.Player = s.player
			s.cliMode.Dungeon = dm
			s.cliMode.SetLevel(s.level)
		}
		if s.cliMode != nil {
			s.cliMode.SetSaveIntegration(s.saveIntegration)
			s.cliMode.SetCommandSession(s.commandSession)
		}
	}
	logger.Debug("Set dungeon manager for game screen")
}

// SetSaveIntegration binds save/load commands to the shared save state.
func (s *GameScreen) SetSaveIntegration(integration *save.SaveGameIntegration) {
	s.saveIntegration = integration
	if s.cliMode != nil {
		s.cliMode.SetSaveIntegration(integration)
	}
}

// SetEngineLoad binds the engine-side loader that the GUI invokes when the
// player requests a load from normal input. The callback is responsible for
// the consume-on-success semantics and mirroring the new state into both
// the engine and the game screen.
func (s *GameScreen) SetEngineLoad(loader func() error) {
	s.engineLoad = loader
}

// ApplyLoadedState mirrors a state returned from the save integration so the
// game screen, command session, and wizard mode track the restored player
// and dungeon after a successful load.
func (s *GameScreen) ApplyLoadedState(player *actor.Player, dm *dungeon.DungeonManager) {
	if player != nil {
		s.player = player
	}
	if dm != nil {
		s.dungeonManager = dm
		s.level = dm.GetCurrentLevel()
	}
	if s.wizardMode != nil && s.level != nil {
		s.wizardMode.Player = s.player
		s.wizardMode.SetLevel(s.level)
	}
	if s.cliMode != nil {
		s.cliMode.Player = s.player
		s.cliMode.Dungeon = s.dungeonManager
		if s.level != nil {
			s.cliMode.SetLevel(s.level)
		}
		s.cliMode.SetSaveIntegration(s.saveIntegration)
	}
	if s.level != nil && s.player != nil {
		s.level.UpdateVisibility(s.player.Position.X, s.player.Position.Y)
	}
}

func (s *GameScreen) executeCommand(cmd command.Command, args ...string) command.Result {
	result := s.commandSession.Execute(&command.Context{
		Player:  s.player,
		Level:   s.level,
		Dungeon: s.dungeonManager,
		Save:    s.saveIntegration,
	}, cmd, args...)
	if result.Player != nil {
		s.player = result.Player
	}
	if result.Dungeon != nil {
		s.dungeonManager = result.Dungeon
	}
	if result.Level != nil {
		s.level = result.Level
	}
	if s.wizardMode != nil && s.level != nil {
		s.wizardMode.Player = s.player
		s.wizardMode.SetLevel(s.level)
	}
	if s.cliMode != nil {
		s.cliMode.Player = s.player
		s.cliMode.Dungeon = s.dungeonManager
		s.cliMode.SetLevel(s.level)
		s.cliMode.SetSaveIntegration(s.saveIntegration)
	}
	return result
}

func (s *GameScreen) addCommandResult(result command.Result) {
	if result.Message != "" {
		s.AddMessage(result.Message)
	}
}

// AddMessage adds a message to the message log
func (s *GameScreen) AddMessage(msg string) {
	for line := range strings.SplitSeq(msg, "\n") {
		s.messages = append(s.messages, line)
		if len(s.messages) > 7 {
			s.messages = s.messages[len(s.messages)-7:]
		}
	}
	logger.Debug("Added message to log",
		"message", msg,
		"messages_count", len(s.messages),
	)
}
