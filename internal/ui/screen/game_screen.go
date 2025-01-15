package screen

import (
	"fmt"
	"image/color"
	"reflect"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/ui/input"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// GameScreen handles the main game display
type GameScreen struct {
	width, height int
	player        *actor.Player
	messages      []string
	lastStats     map[string]interface{} // 前回のステータス情報
}

// NewGameScreen creates a new game screen
func NewGameScreen(width, height int, player *actor.Player) *GameScreen {
	screen := &GameScreen{
		width:     width,
		height:    height,
		player:    player,
		messages:  make([]string, 0, 3), // 3行分のメッセージを保持
		lastStats: make(map[string]interface{}),
	}
	logger.Debug("Created game screen",
		"width", width,
		"height", height,
	)
	return screen
}

// Update updates the game screen state
func (s *GameScreen) Update() state.GameState {
	// ESCキーでメニューに戻る
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		logger.Info("Returning to menu")
		return state.StateMenu
	}

	// 移動処理
	dx, dy := input.GetMovementDirection()
	if dx != 0 || dy != 0 {
		newX := s.player.Position.X + dx
		newY := s.player.Position.Y + dy
		// TODO: 移動の有効性チェックを実装
		s.player.Position.Move(dx, dy)
		logger.Debug("Player moved",
			"from_x", newX-dx,
			"from_y", newY-dy,
			"to_x", newX,
			"to_y", newY,
		)
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
func (s *GameScreen) Draw(screen *ebiten.Image) {
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

	// ステータス行の描画（上部2行）
	statusLine1 := fmt.Sprintf(
		"Lv:%d HP:%d/%d Atk:%d Def:%d Hunger:%d%% Exp:%d Gold:%d",
		s.player.Level,
		s.player.HP,
		s.player.MaxHP,
		s.player.Attack,
		s.player.Defense,
		s.player.Hunger,
		s.player.Exp,
		s.player.Gold,
	)
	text.Draw(screen, statusLine1, GetFont(), 1, 24, color.White)

	// TODO: 装備情報の描画
	statusLine2 := fmt.Sprintf(
		"Weap:None Armor:None Ring(L):None Ring(R):None",
	)
	text.Draw(screen, statusLine2, GetFont(), 1, 48, color.White)

	// 上部区切り線
	separator := strings.Repeat("━", s.width)
	text.Draw(screen, separator, GetFont(), 0, 72, color.White)

	// メッセージログの描画（下部3行）
	messageY := s.height*24 - 72 // 3行分上
	text.Draw(screen, separator, GetFont(), 0, messageY, color.White)
	for i, msg := range s.messages {
		text.Draw(screen, msg, GetFont(), 1, messageY+24*(i+1), color.White)
	}
}
