package screen

import (
	"sort"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// SymbolScreen represents the symbol explanation screen
type SymbolScreen struct {
	width  int
	height int
	grid   gruid.Grid
}

type symbolExplanation struct {
	symbol string
	name   string
}

// シンボル説明データ。モンスター名はactor.MonsterTypesを正規ソースとする。
func symbolExplanations() []symbolExplanation {
	explanations := []symbolExplanation{
		// 地形
		{".", "floor"},
		{"#", "wall"},
		{"+", "closed door"},
		{"-", "horizontal open door"},
		{"|", "vertical open door"},
		{"^", "trap"},
		{"%", "stairs"},

		// プレイヤー
		{"@", "you"},
	}

	// モンスター
	monsters := make([]actor.MonsterType, 0, len(actor.MonsterTypes))
	for _, monsterType := range actor.MonsterTypes {
		if monsterType.Symbol >= 'A' && monsterType.Symbol <= 'Z' {
			monsters = append(monsters, monsterType)
		}
	}
	sort.Slice(monsters, func(i, j int) bool {
		return monsters[i].Symbol < monsters[j].Symbol
	})
	for _, monsterType := range monsters {
		explanations = append(explanations, symbolExplanation{string(monsterType.Symbol), monsterType.Name})
	}

	// アイテム（PyRogue準拠）
	explanations = append(explanations,
		symbolExplanation{")", "weapon"},
		symbolExplanation{"]", "armor"},
		symbolExplanation{"!", "potion"},
		symbolExplanation{"?", "scroll"},
		symbolExplanation{"/", "wand or staff"},
		symbolExplanation{"=", "ring"},
		symbolExplanation{",", "amulet"},
		symbolExplanation{":", "food"},
		symbolExplanation{"*", "gold"},
	)
	return explanations
}

// NewSymbolScreen creates a new symbol explanation screen
func NewSymbolScreen(width, height int) *SymbolScreen {
	return &SymbolScreen{
		width:  width,
		height: height,
		grid:   gruid.NewGrid(width, height),
	}
}

// HandleInput handles input events
func (s *SymbolScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		logger.Debug("SymbolScreen key pressed", "key", msg.Key)
		switch msg.Key {
		case gruid.KeyEscape, gruid.KeyEnter, gruid.KeySpace:
			// 任意のキーで前の画面に戻る
			return state.StateGame
		default:
			// 文字キーでも戻れるように
			keyStr := string(msg.Key)
			if len(keyStr) == 1 {
				return state.StateGame
			}
		}
	}

	return state.StateSymbol
}

// Draw draws the symbol explanation screen (PyRogue準拠)
func (s *SymbolScreen) Draw(grid *gruid.Grid) {
	// グリッドをクリア
	grid.Fill(gruid.Cell{Rune: ' ', Style: gruid.Style{Bg: 0x000000, Fg: 0xFFFFFF}})

	// PyRogue風のタイトル
	title := "Character Descriptions"
	titleX := (s.width - len(title)) / 2
	s.drawText(grid, titleX, 2, title, gruid.Style{Fg: 0xFFFF00})

	// 2カラムレイアウト（PyRogue準拠）
	leftCol := 10
	rightCol := 45
	startY := 5
	explanations := symbolExplanations()

	// 左カラム
	y := startY
	for i := 0; i < len(explanations)/2 && y < s.height-3; i++ {
		sym := explanations[i]
		s.drawSymbol(grid, leftCol, y, sym.symbol, sym.name)

		y++
	}

	// 右カラム
	y = startY
	for i := len(explanations) / 2; i < len(explanations) && y < s.height-3; i++ {
		sym := explanations[i]
		s.drawSymbol(grid, rightCol, y, sym.symbol, sym.name)

		y++
	}

	// 下部の説明（PyRogue準拠）
	helpText := "--Press any key to continue--"
	helpX := (s.width - len(helpText)) / 2
	s.drawText(grid, helpX, s.height-2, helpText, gruid.Style{Fg: 0x808080})

	logger.Trace("Symbol screen drawn")
}

func (s *SymbolScreen) drawSymbol(grid *gruid.Grid, x, y int, symbol, name string) {
	symbolStyle := gruid.Style{Fg: 0xFFFFFF}
	switch symbol[0] {
	case '@':
		symbolStyle.Fg = 0x00FF00 // プレイヤーは緑
	case ')', ']', '!', '?', '/', '=', ',', ':', '*':
		symbolStyle.Fg = 0xFFFF00 // アイテムは黄色
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
		'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		symbolStyle.Fg = 0xFF0000 // モンスターは赤
	case '%':
		symbolStyle.Fg = 0x00FFFF // 階段はシアン
	case '^':
		symbolStyle.Fg = 0xFF00FF // トラップはマゼンタ
	}

	text := symbol + " " + name
	s.drawText(grid, x, y, text[:1], symbolStyle)
	s.drawText(grid, x+1, y, text[1:], gruid.Style{Fg: 0xCCCCCC})
}

// drawText draws text at the specified position with the given style
func (s *SymbolScreen) drawText(grid *gruid.Grid, x, y int, text string, style gruid.Style) {
	for i, r := range []rune(text) {
		pos := gruid.Point{X: x + i, Y: y}
		if pos.X >= 0 && pos.X < grid.Size().X && pos.Y >= 0 && pos.Y < grid.Size().Y {
			grid.Set(pos, gruid.Cell{Rune: r, Style: style})
		}
	}
}
