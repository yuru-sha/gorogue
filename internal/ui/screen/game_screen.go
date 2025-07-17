package screen

import (
	"fmt"
	"reflect"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/core/wizard"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
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
	wizardMode    *wizard.WizardMode     // ウィザードモード
}

// NewGameScreen creates a new game screen
func NewGameScreen(width, height int, player *actor.Player) *GameScreen {
	screen := &GameScreen{
		width:     width,
		height:    height,
		player:    player,
		messages:  make([]string, 0, 7), // 7行分のメッセージを保持
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
		case "Left", "h", gruid.KeyArrowLeft:
			s.tryMovePlayer(-1, 0)
		case "Right", "l", gruid.KeyArrowRight:
			s.tryMovePlayer(1, 0)
		case "Up", "k", gruid.KeyArrowUp:
			s.tryMovePlayer(0, -1)
		case "Down", "j", gruid.KeyArrowDown:
			s.tryMovePlayer(0, 1)
		case "y":
			s.tryMovePlayer(-1, -1)
		case "u":
			s.tryMovePlayer(1, -1)
		case "b":
			s.tryMovePlayer(-1, 1)
		case "n":
			s.tryMovePlayer(1, 1)
		case "q":
			logger.Info("Quit requested")
			return state.StateMenu
		case "^W", "W": // Ctrl+W or Shift+W to toggle wizard mode
			s.wizardMode.Toggle()
			status := "OFF"
			if s.wizardMode.IsActive {
				status = "ON"
			}
			s.AddMessage(fmt.Sprintf("ウィザードモード: %s", status))
		default:
			// Check if it's a wizard command
			if s.wizardMode.IsActive && len(msg.Key) == 1 {
				result := s.wizardMode.ExecuteCommand(rune(msg.Key[0]))
				if result != "" {
					s.AddMessage(result)
				}
			}
		}
	}

	return state.StateGame
}

// tryMovePlayer attempts to move the player in the given direction
func (s *GameScreen) tryMovePlayer(dx, dy int) {
	newX := s.player.Position.X + dx
	newY := s.player.Position.Y + dy
	
	// 境界チェック
	if newX < 0 || newX >= s.level.Width || newY < 0 || newY >= s.level.Height {
		logger.Debug("Player movement blocked by bounds",
			"current_x", s.player.Position.X,
			"current_y", s.player.Position.Y,
			"new_x", newX,
			"new_y", newY,
		)
		return
	}
	
	// 壁の衝突判定
	tile := s.level.GetTile(newX, newY)
	if !tile.Walkable() {
		logger.Debug("Player movement blocked by wall",
			"current_x", s.player.Position.X,
			"current_y", s.player.Position.Y,
			"new_x", newX,
			"new_y", newY,
			"tile_type", tile.Type,
		)
		return
	}
	
	// モンスターとの戦闘判定
	monster := s.level.GetMonsterAt(newX, newY)
	if monster != nil {
		s.playerAttackMonster(monster)
		return
	}
	
	// 移動実行
	s.player.Position.Move(dx, dy)
	logger.Debug("Player moved",
		"new_x", s.player.Position.X,
		"new_y", s.player.Position.Y,
	)
	
	// アイテムを拾う処理
	s.pickupItem(newX, newY)
	
	// モンスターのターンを実行
	s.level.UpdateMonsters(s.player)
}

// playerAttackMonster handles player attacking a monster
func (s *GameScreen) playerAttackMonster(monster *actor.Monster) {
	damage := s.player.CalculateDamage(monster.Defense)
	monster.TakeDamage(damage)
	
	message := fmt.Sprintf("%sに%dのダメージを与えた！", monster.Type.Name, damage)
	s.AddMessage(message)
	
	if !monster.IsAlive() {
		deathMessage := fmt.Sprintf("%sを倒した！", monster.Type.Name)
		s.AddMessage(deathMessage)
		
		// 経験値とゴールドを取得
		exp := monster.MaxHP + monster.Attack
		gold := monster.MaxHP / 2
		
		s.player.GainExp(exp)
		s.player.AddGold(gold)
		
		rewardMessage := fmt.Sprintf("%d経験値、%dゴールドを得た", exp, gold)
		s.AddMessage(rewardMessage)
	} else {
		// モンスターのターンを実行
		s.level.UpdateMonsters(s.player)
	}
}

// pickupItem handles picking up an item at the given position
func (s *GameScreen) pickupItem(x, y int) {
	item := s.level.GetItemAt(x, y)
	if item == nil {
		return
	}
	
	// アイテムタイプに応じた処理
	switch item.Type {
	case gameitem.ItemGold:
		s.player.AddGold(item.Value)
		s.AddMessage(fmt.Sprintf("%dゴールドを拾った", item.Value))
	case gameitem.ItemFood:
		s.player.Hunger = min(s.player.Hunger+20, 100)
		s.AddMessage(fmt.Sprintf("%sを食べた", item.Name))
	case gameitem.ItemPotion:
		s.player.HP = min(s.player.HP+item.Value, s.player.MaxHP)
		s.AddMessage(fmt.Sprintf("%sを飲んだ", item.Name))
	default:
		// その他のアイテムは単純に拾う
		s.AddMessage(fmt.Sprintf("%sを拾った", item.Name))
	}
	
	// アイテムをレベルから削除
	s.level.RemoveItem(item)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	
	// ウィザードモードの表示を追加
	if s.wizardMode != nil && s.wizardMode.IsActive {
		statusLine2 += "  [WIZARD MODE]"
	}
	
	s.drawText(grid, 0, 1, statusLine2, gruid.Style{})

	// ダンジョンの描画
	for y := 0; y < s.level.Height; y++ {
		for x := 0; x < s.level.Width; x++ {
			tile := s.level.GetTile(x, y)
			grid.Set(gruid.Point{X: x, Y: y + 2}, tile.Cell)
		}
	}

	// アイテムの描画（最初に描画）
	for _, item := range s.level.Items {
		color := item.GetColor()
		rgb := (uint32(color[0]) << 16) | (uint32(color[1]) << 8) | uint32(color[2])
		grid.Set(gruid.Point{X: item.Position.X, Y: item.Position.Y + 2}, gruid.Cell{
			Rune:  item.Symbol,
			Style: gruid.Style{Fg: gruid.Color(rgb)}, // アイテムの色で表示
		})
	}

	// モンスターの描画（アイテムの上に描画）
	for _, monster := range s.level.Monsters {
		if monster.IsAlive() {
			grid.Set(gruid.Point{X: monster.Position.X, Y: monster.Position.Y + 2}, gruid.Cell{
				Rune:  monster.Type.Symbol,
				Style: gruid.Style{Fg: 0xFF0000}, // Red (モンスターを赤色で表示)
			})
		}
	}
	
	// プレイヤーの描画（最上位に描画）
	grid.Set(gruid.Point{X: s.player.Position.X, Y: s.player.Position.Y + 2}, gruid.Cell{
		Rune:  '@',
		Style: gruid.Style{Fg: 0x00FF00}, // Green (プレイヤーを緑色で表示)
	})

	// メッセージログの描画（下部7行）
	for i, msg := range s.messages {
		s.drawText(grid, 1, s.height-7+i, fmt.Sprintf(" %s", msg), gruid.Style{})
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
	s.wizardMode = wizard.NewWizardMode(level, s.player)
	logger.Debug("Set dungeon level for game screen",
		"width", level.Width,
		"height", level.Height,
	)
}
