package screen

import (
	"fmt"
	"reflect"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// GameScreen handles the main game display
type GameScreen struct {
	width, height int
	player        *actor.Player
	level         *dungeon.Level
	messages      []string
	lastStats     map[string]interface{} // 前回のステータス情報
	grid          gruid.Grid             // 画面全体のグリッド
}

// NewGameScreen creates a new game screen
func NewGameScreen(width, height int, player *actor.Player) *GameScreen {
	screen := &GameScreen{
		width:     width,
		height:    height,
		player:    player,
		messages:  make([]string, 0, 3), // 3行分のメッセージを保持
		lastStats: make(map[string]interface{}),
		grid:      gruid.NewGrid(width, height),
	}
	logger.Debug("Created game screen",
		"width", width,
		"height", height,
	)
	return screen
}

// HandleInput handles input events
func (s *GameScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case gruid.KeyEscape:
			logger.Info("Returning to menu")
			return state.StateMenu
		case "Left", "h":
			s.player.Position.Move(-1, 0)
		case "Right", "l":
			s.player.Position.Move(1, 0)
		case "Up", "k":
			s.player.Position.Move(0, -1)
		case "Down", "j":
			s.player.Position.Move(0, 1)
		case "y":
			s.player.Position.Move(-1, -1)
		case "u":
			s.player.Position.Move(1, -1)
		case "b":
			s.player.Position.Move(-1, 1)
		case "n":
			s.player.Position.Move(1, 1)
		}
	}

	return state.StateGame
}

// AddMessage adds a message to the message log
func (s *GameScreen) AddMessage(msg string) {
	s.messages = append(s.messages, msg)
	if len(s.messages) > 3 {
		s.messages = s.messages[len(s.messages)-3:]
	}
	logger.Debug("Added message to log",
		"message", msg,
		"messages_count", len(s.messages),
	)
}

// Draw draws the game screen
func (s *GameScreen) Draw(grid *gruid.Grid) {
	// 現在のステータス情報を収集
	currentStats := map[string]interface{}{
		"level":   s.player.Level,
		"hp":      s.player.HP,
		"max_hp":  s.player.MaxHP,
		"attack":  s.player.Attack,
		"defense": s.player.Defense,
		"hunger":  s.player.Hunger,
		"exp":     s.player.Exp,
		"gold":    s.player.Gold,
	}

	// ステータスに変更があった場合のみログ出力
	if !reflect.DeepEqual(s.lastStats, currentStats) {
		logger.Debug("Player stats changed",
			"level", s.player.Level,
			"hp", s.player.HP,
			"max_hp", s.player.MaxHP,
			"attack", s.player.Attack,
			"defense", s.player.Defense,
			"hunger", s.player.Hunger,
			"exp", s.player.Exp,
			"gold", s.player.Gold,
		)
		s.lastStats = currentStats
	}

	// 画面描画の詳細ログはTRACEレベルで出力
	logger.Trace("Drawing game screen")

	// グリッドをクリア
	grid.Fill(gruid.Cell{Rune: ' '})

	// ステータス行の描画（上部2行）
	statusLine1 := fmt.Sprintf(
		" Lv:%2d  HP:%3d/%3d  Atk:%2d  Def:%2d  Hunger:%3d%%  Exp:%4d  Gold:%4d",
		s.player.Level,
		s.player.HP,
		s.player.MaxHP,
		s.player.Attack,
		s.player.Defense,
		s.player.Hunger,
		s.player.Exp,
		s.player.Gold,
	)
	s.drawText(grid, 0, 0, statusLine1, gruid.Style{})

	// 装備情報の描画
	statusLine2 := fmt.Sprintf(
		" Weap:%-12s  Armor:%-12s  Ring(L):%-12s  Ring(R):%-12s",
		"None",
		"None",
		"None",
		"None",
	)
	s.drawText(grid, 0, 1, statusLine2, gruid.Style{})

	// ダンジョンの描画
	for y := 0; y < s.level.Height; y++ {
		for x := 0; x < s.level.Width; x++ {
			tile := s.level.GetTile(x, y)
			grid.Set(gruid.Point{X: x, Y: y + 2}, tile.Cell)
		}
	}

	// プレイヤーの描画
	grid.Set(gruid.Point{X: s.player.Position.X, Y: s.player.Position.Y + 2}, gruid.Cell{
		Rune:  '@',
		Style: gruid.Style{},
	})

	// メッセージログの描画（下部3行）
	for i, msg := range s.messages {
		s.drawText(grid, 1, s.height-3+i, fmt.Sprintf(" %s", msg), gruid.Style{})
	}
}

// drawText draws text at the specified position with the given style
func (s *GameScreen) drawText(grid *gruid.Grid, x, y int, text string, style gruid.Style) {
	for i, r := range text {
		pos := gruid.Point{X: x + i, Y: y}
		if pos.X >= grid.Size().X {
			break
		}
		grid.Set(pos, gruid.Cell{Rune: r, Style: style})
	}
}

// SetLevel sets the dungeon level for the game screen
func (s *GameScreen) SetLevel(level *dungeon.Level) {
	s.level = level
	logger.Debug("Set dungeon level for game screen",
		"width", level.Width,
		"height", level.Height,
	)
}
