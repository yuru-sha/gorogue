package cli

import (
	"strings"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func init() {
	// テスト用のログ初期化
	logger.Setup()
}

func TestNewCLIMode(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{
		Width:  20,
		Height: 20,
	}

	cli := NewCLIMode(level, player)

	if cli == nil {
		t.Fatal("NewCLIMode() returned nil")
	}

	if cli.Level != level {
		t.Error("CLIMode level not set correctly")
	}

	if cli.Player != player {
		t.Error("CLIMode player not set correctly")
	}

	if cli.IsActive {
		t.Error("CLIMode should start inactive")
	}

	if len(cli.Commands) == 0 {
		t.Error("CLIMode should have commands registered")
	}
}

func TestCLIModeLevelCommandUsesDungeonManager(t *testing.T) {
	player := actor.NewPlayerWithSeed(0, 0, 12345)
	manager := dungeon.NewDungeonManagerWithSeed(player, 12345)
	cli := NewCLIModeWithDungeonManager(manager, player)
	cli.IsActive = true

	result := cli.ExecuteCommand("level 2")

	if manager.GetCurrentFloor() != 2 || cli.Level != manager.GetCurrentLevel() {
		t.Fatalf("CLI did not switch to floor 2: floor=%d level_matches=%t", manager.GetCurrentFloor(), cli.Level == manager.GetCurrentLevel())
	}
	if !strings.Contains(result, "Moved to floor 2") {
		t.Fatalf("unexpected level command result: %s", result)
	}
}

func TestCLILookRespectsVisibility(t *testing.T) {
	level := &dungeon.Level{
		Width:  5,
		Height: 5,
		Tiles:  make([][]*dungeon.Tile, 5),
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*dungeon.Tile, level.Width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = dungeon.NewTile(dungeon.TileFloor)
		}
	}

	visible := level.GetTile(2, 1)
	visible.Visible = true
	level.Monsters = []*actor.Monster{actor.NewMonster(2, 1, 'B')}
	level.Items = []*item.Item{item.NewItem(2, 1, item.ItemFood, "visible ration", 1)}

	explored := level.GetTile(3, 3)
	explored.Explored = true
	level.Monsters = append(level.Monsters, actor.NewMonster(3, 3, 'O'))
	level.Items = append(level.Items, item.NewItem(3, 3, item.ItemFood, "remembered ration", 1))

	level.Monsters = append(level.Monsters, actor.NewMonster(4, 4, 'Z'))
	level.Items = append(level.Items, item.NewItem(4, 4, item.ItemFood, "hidden ration", 1))

	cli := NewCLIMode(level, actor.NewPlayerWithSeed(1, 1, 42))
	cli.IsActive = true

	visibleResult := cli.ExecuteCommand("look 2 1")
	if !strings.Contains(visibleResult, "Terrain: floor") ||
		!strings.Contains(visibleResult, "Monster:") || !strings.Contains(visibleResult, "Items:") {
		t.Fatalf("visible look result = %q", visibleResult)
	}

	exploredResult := cli.ExecuteCommand("look 3 3")
	if !strings.Contains(exploredResult, "Terrain: floor") ||
		strings.Contains(exploredResult, "Monster:") || strings.Contains(exploredResult, "Items:") {
		t.Fatalf("explored look result = %q", exploredResult)
	}

	unexploredResult := cli.ExecuteCommand("look 4 4")
	if !strings.Contains(unexploredResult, "You cannot see that location.") ||
		strings.Contains(unexploredResult, "Terrain:") ||
		strings.Contains(unexploredResult, "Monster:") || strings.Contains(unexploredResult, "Items:") {
		t.Fatalf("unexplored look result = %q", unexploredResult)
	}
}

func TestCLIModeToggle(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)

	// 初期状態は非アクティブ
	if cli.IsActive {
		t.Error("CLIMode should start inactive")
	}

	// アクティブ化
	cli.Toggle()
	if !cli.IsActive {
		t.Error("CLIMode should be active after toggle")
	}

	// 非アクティブ化
	cli.Toggle()
	if cli.IsActive {
		t.Error("CLIMode should be inactive after second toggle")
	}
}

func TestCLIModeExecuteCommand(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	tests := []struct {
		name        string
		command     string
		expectError bool
		contains    string
	}{
		{
			name:        "ヘルプコマンド",
			command:     "help",
			expectError: false,
			contains:    "Available commands:",
		},
		{
			name:        "ステータスコマンド",
			command:     "status",
			expectError: false,
			contains:    "Player Status",
		},
		{
			name:        "存在しないコマンド",
			command:     "nonexistent",
			expectError: true,
			contains:    "Unknown command",
		},
		{
			name:        "空のコマンド",
			command:     "",
			expectError: true,
			contains:    "No command entered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.ExecuteCommand(tt.command)

			if tt.expectError {
				if !strings.Contains(result, "Unknown command") && !strings.Contains(result, "No command entered") {
					t.Errorf("Expected error message, got: %s", result)
				}
			} else {
				if strings.Contains(result, "Unknown command") || strings.Contains(result, "No command entered") {
					t.Errorf("Expected successful execution, got error: %s", result)
				}
			}

			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}

func TestCLIModeHelpCommand(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	result := cli.ExecuteCommand("help")

	// ヘルプメッセージの基本的な内容をチェック
	expectedContent := []string{
		"Available commands:",
		"help",
		"status",
		"heal",
		"position",
	}

	for _, content := range expectedContent {
		if !strings.Contains(result, content) {
			t.Errorf("Help output should contain %q, got: %s", content, result)
		}
	}
}

func TestCLIModeStatusCommand(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	result := cli.ExecuteCommand("status")

	// ステータス表示の基本的な内容をチェック
	expectedContent := []string{
		"Player Status",
		"Level:",
		"HP:",
		"Attack:",
		"Defense:",
		"Position:",
	}

	for _, content := range expectedContent {
		if !strings.Contains(result, content) {
			t.Errorf("Status output should contain %q, got: %s", content, result)
		}
	}
}

func TestCLIModePositionCommand(t *testing.T) {
	player := actor.NewPlayer(10, 15)
	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	result := cli.ExecuteCommand("status")

	// 位置情報の確認（statusコマンドの出力に位置情報が含まれる）
	if !strings.Contains(result, "10") || !strings.Contains(result, "15") {
		t.Errorf("Position output should contain player coordinates (10, 15), got: %s", result)
	}
}

func TestCLIModeHealCommand(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	// プレイヤーにダメージを与えてテスト
	player.TakeDamage(10)
	originalHP := player.HP

	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	result := cli.ExecuteCommand("heal 5")

	// HPが回復したかチェック
	expectedHP := originalHP + 5
	if player.HP != expectedHP {
		t.Errorf("Player HP should be %d after heal, got %d", expectedHP, player.HP)
	}

	// 結果メッセージのチェック
	if !strings.Contains(result, "Healed") {
		t.Errorf("Heal result should contain 'Healed', got: %s", result)
	}
}

func TestCLIModeGoldCommand(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	originalGold := player.Gold

	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	result := cli.ExecuteCommand("gold 100")

	// ゴールドが追加されたかチェック
	expectedGold := originalGold + 100
	if player.Gold != expectedGold {
		t.Errorf("Player gold should be %d after adding gold, got %d", expectedGold, player.Gold)
	}

	// 結果メッセージのチェック
	if !strings.Contains(result, "gold") {
		t.Errorf("Gold result should contain 'gold', got: %s", result)
	}
}

func TestCLIModeInvalidArguments(t *testing.T) {
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{}
	cli := NewCLIMode(level, player)
	cli.IsActive = true // CLIモードをアクティブ化

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"無効な数値 - heal", "heal abc", true},
		{"無効な数値 - gold", "gold xyz", true},
		{"負の数値 - heal", "heal -5", false},
		{"負の数値 - gold", "gold -100", false},
		{"引数不足 - heal", "heal", false},
		{"引数不足 - gold", "gold", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cli.ExecuteCommand(tt.command)

			if tt.wantErr {
				if !strings.Contains(result, "Invalid") && !strings.Contains(result, "Error") {
					t.Errorf("Expected error message for %q, got: %s", tt.command, result)
				}
			} else {
				if strings.Contains(result, "Invalid") || strings.Contains(result, "Error") {
					t.Errorf("Expected success message for %q, got: %s", tt.command, result)
				}
			}
		})
	}
}
