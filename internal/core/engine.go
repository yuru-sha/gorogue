package core

import (
	"image/color"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/ui/input"
	"github.com/yuru-sha/gorogue/internal/ui/screen"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 80
	screenHeight = 50
	tileSize     = 24
)

// Engine represents the game engine and implements ebiten.Game interface
type Engine struct {
	level      *dungeon.Level
	player     *actor.Player
	gameScreen *screen.GameScreen
	gameFont   font.Face
}

// NewEngine creates and initializes a new game engine
func NewEngine() *Engine {
	// macOSでのEbitenの初期化問題を回避
	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
	}

	// フォントの初期化
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		logger.Error("Failed to parse font", "error", err.Error())
		return nil
	}

	gameFont, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		logger.Error("Failed to create font face", "error", err.Error())
		return nil
	}

	// ダンジョンレベルの生成
	level := dungeon.NewLevel(screenWidth, screenHeight, 1)
	logger.Debug("Created new dungeon level",
		"width", level.Width,
		"height", level.Height,
	)

	// プレイヤーの生成
	player := actor.NewPlayer(screenWidth/2, screenHeight/2)
	logger.Debug("Created player",
		"x", player.Position.X,
		"y", player.Position.Y,
	)

	// ゲーム画面の生成
	gameScreen := screen.NewGameScreen(screenWidth, screenHeight, player)
	logger.Debug("Created game screen")

	engine := &Engine{
		level:      level,
		player:     player,
		gameScreen: gameScreen,
		gameFont:   gameFont,
	}

	return engine
}

// Update handles the game logic updates
func (e *Engine) Update() error {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logger.Error("Panic recovered in Update",
				"panic", r,
				"stack", string(stack),
				"player_x", e.player.Position.X,
				"player_y", e.player.Position.Y,
				"level", e.level.FloorNumber,
			)
		}
	}()

	// 移動処理
	dx, dy := input.GetMovementDirection()
	if dx != 0 || dy != 0 {
		newX := e.player.Position.X + dx
		newY := e.player.Position.Y + dy
		if e.isValidMove(newX, newY) {
			oldX, oldY := e.player.Position.X, e.player.Position.Y
			e.player.Position.Move(dx, dy)
			logger.Debug("Player moved",
				"from_x", oldX,
				"from_y", oldY,
				"to_x", newX,
				"to_y", newY,
				"tile", e.level.GetTile(newX, newY).Type.String(),
			)
		} else {
			logger.Debug("Movement blocked",
				"attempted_x", newX,
				"attempted_y", newY,
				"reason", e.getMovementBlockReason(newX, newY),
			)
		}
	}

	// 終了判定
	if input.IsQuitRequested() {
		logger.Info("Game quit requested", nil)
		os.Exit(0)
	}

	return nil
}

// getMovementBlockReason returns the reason why movement was blocked
func (e *Engine) getMovementBlockReason(x, y int) string {
	if x < 0 || x >= e.level.Width || y < 0 || y >= e.level.Height {
		return "out_of_bounds"
	}
	tile := e.level.GetTile(x, y)
	if tile == nil {
		return "no_tile"
	}
	if !tile.Walkable {
		return "unwalkable_" + tile.Type.String()
	}
	return "unknown"
}

// isValidMove checks if the given position is valid for movement
func (e *Engine) isValidMove(x, y int) bool {
	// 画面外チェック
	if x < 0 || x >= e.level.Width || y < 0 || y >= e.level.Height {
		logger.Debug("Movement out of bounds",
			"attempted_x", x,
			"attempted_y", y,
			"bounds_width", e.level.Width,
			"bounds_height", e.level.Height,
		)
		return false
	}

	// タイルの歩行可能判定
	tile := e.level.GetTile(x, y)
	if tile == nil {
		logger.Debug("No tile at position",
			"position_x", x,
			"position_y", y,
		)
		return false
	}
	if !tile.Walkable {
		logger.Debug("Tile not walkable",
			"position_x", x,
			"position_y", y,
			"tile", tile.Type.String(),
		)
	}
	return tile.Walkable
}

// Draw renders the game state
func (e *Engine) Draw(screen *ebiten.Image) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logger.Error("Panic recovered in Draw",
				"panic", r,
				"stack", string(stack),
			)
		}
	}()

	// 背景を黒で塗りつぶす
	screen.Fill(color.Black)

	// ダンジョンの描画
	for y := 0; y < e.level.Height; y++ {
		for x := 0; x < e.level.Width; x++ {
			tile := e.level.GetTile(x, y)
			if tile != nil {
				// タイルの文字を描画
				text.Draw(screen, string(tile.Symbol), e.gameFont, x*tileSize, (y+2)*tileSize, color.RGBA{
					R: tile.Color[0],
					G: tile.Color[1],
					B: tile.Color[2],
					A: 255,
				})
			}
		}
	}

	// プレイヤーの描画
	text.Draw(screen, string(e.player.Symbol), e.gameFont,
		e.player.Position.X*tileSize,
		(e.player.Position.Y+2)*tileSize,
		color.White)

	// ゲーム画面（ステータスバーなど）の描画
	e.gameScreen.Draw(screen)
}

// Layout returns the game's screen dimensions
func (e *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth * tileSize, screenHeight * tileSize
}
