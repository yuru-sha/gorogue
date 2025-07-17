package actor

import (
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func init() {
	// テスト用のログ初期化
	logger.Setup()
}

func TestNewActor(t *testing.T) {
	tests := []struct {
		name            string
		x, y            int
		symbol          rune
		color           gruid.Color
		hp, attack, def int
		wantX, wantY    int
		wantHP          int
		wantMaxHP       int
		wantAttack      int
		wantDefense     int
	}{
		{
			name:        "基本的なアクター作成",
			x:           5,
			y:           10,
			symbol:      '@',
			color:       0xFFFFFF,
			hp:          20,
			attack:      8,
			def:         3,
			wantX:       5,
			wantY:       10,
			wantHP:      20,
			wantMaxHP:   20,
			wantAttack:  8,
			wantDefense: 3,
		},
		{
			name:        "最小値でのアクター作成",
			x:           0,
			y:           0,
			symbol:      'T',
			color:       0x00FF00,
			hp:          1,
			attack:      1,
			def:         0,
			wantX:       0,
			wantY:       0,
			wantHP:      1,
			wantMaxHP:   1,
			wantAttack:  1,
			wantDefense: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor := NewActor(tt.x, tt.y, tt.symbol, tt.color, tt.hp, tt.attack, tt.def)

			if actor.Position.X != tt.wantX {
				t.Errorf("NewActor() X = %v, want %v", actor.Position.X, tt.wantX)
			}
			if actor.Position.Y != tt.wantY {
				t.Errorf("NewActor() Y = %v, want %v", actor.Position.Y, tt.wantY)
			}
			if actor.Symbol != tt.symbol {
				t.Errorf("NewActor() Symbol = %v, want %v", actor.Symbol, tt.symbol)
			}
			if actor.Color != tt.color {
				t.Errorf("NewActor() Color = %v, want %v", actor.Color, tt.color)
			}
			if actor.HP != tt.wantHP {
				t.Errorf("NewActor() HP = %v, want %v", actor.HP, tt.wantHP)
			}
			if actor.MaxHP != tt.wantMaxHP {
				t.Errorf("NewActor() MaxHP = %v, want %v", actor.MaxHP, tt.wantMaxHP)
			}
			if actor.Attack != tt.wantAttack {
				t.Errorf("NewActor() Attack = %v, want %v", actor.Attack, tt.wantAttack)
			}
			if actor.Defense != tt.wantDefense {
				t.Errorf("NewActor() Defense = %v, want %v", actor.Defense, tt.wantDefense)
			}
		})
	}
}

func TestActorIsAlive(t *testing.T) {
	tests := []struct {
		name string
		hp   int
		want bool
	}{
		{"生きているアクター", 10, true},
		{"最小HP", 1, true},
		{"死んでいるアクター", 0, false},
		{"負のHP", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor := NewActor(0, 0, '@', 0xFFFFFF, 20, 5, 2)
			actor.HP = tt.hp

			if got := actor.IsAlive(); got != tt.want {
				t.Errorf("Actor.IsAlive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActorTakeDamage(t *testing.T) {
	tests := []struct {
		name      string
		initialHP int
		damage    int
		wantHP    int
	}{
		{"通常のダメージ", 20, 5, 15},
		{"最大ダメージ", 20, 20, 0},
		{"オーバーキル", 20, 25, 0},
		{"ゼロダメージ", 20, 0, 20},
		{"最小ダメージ", 1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor := NewActor(0, 0, '@', 0xFFFFFF, tt.initialHP, 5, 2)
			actor.TakeDamage(tt.damage)

			if actor.HP != tt.wantHP {
				t.Errorf("Actor.TakeDamage() HP = %v, want %v", actor.HP, tt.wantHP)
			}
		})
	}
}

func TestActorHeal(t *testing.T) {
	tests := []struct {
		name       string
		initialHP  int
		maxHP      int
		healAmount int
		wantHP     int
	}{
		{"通常の回復", 10, 20, 5, 15},
		{"最大値まで回復", 10, 20, 10, 20},
		{"過剰回復", 10, 20, 15, 20},
		{"ゼロ回復", 10, 20, 0, 10},
		{"満タンから回復", 20, 20, 5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor := NewActor(0, 0, '@', 0xFFFFFF, tt.maxHP, 5, 2)
			actor.HP = tt.initialHP
			actor.Heal(tt.healAmount)

			if actor.HP != tt.wantHP {
				t.Errorf("Actor.Heal() HP = %v, want %v", actor.HP, tt.wantHP)
			}
		})
	}
}

func TestActorCalculateDamage(t *testing.T) {
	tests := []struct {
		name          string
		attack        int
		targetDefense int
		wantDamage    int
	}{
		{"通常の攻撃", 10, 3, 7},
		{"防御力が高い", 5, 4, 1},
		{"防御力が同じ", 5, 5, 1},
		{"防御力が上回る", 5, 10, 1},
		{"最大攻撃力", 20, 5, 15},
		{"最小ダメージ保証", 1, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor := NewActor(0, 0, '@', 0xFFFFFF, 20, tt.attack, 2)
			damage := actor.CalculateDamage(tt.targetDefense)

			if damage != tt.wantDamage {
				t.Errorf("Actor.CalculateDamage() = %v, want %v", damage, tt.wantDamage)
			}
		})
	}
}

// 統合テスト: 戦闘シナリオ
func TestActorCombatScenario(t *testing.T) {
	// プレイヤーアクター（攻撃力10、防御力3、HP20）
	player := NewActor(0, 0, '@', 0xFFFFFF, 20, 10, 3)
	// モンスターアクター（攻撃力8、防御力2、HP15）
	monster := NewActor(5, 5, 'M', 0xFF0000, 15, 8, 2)

	// プレイヤーがモンスターを攻撃
	damage := player.CalculateDamage(monster.Defense)
	expectedDamage := 10 - 2 // 8ダメージ
	if damage != expectedDamage {
		t.Errorf("Player damage = %v, want %v", damage, expectedDamage)
	}

	monster.TakeDamage(damage)
	expectedMonsterHP := 15 - 8 // 7HP
	if monster.HP != expectedMonsterHP {
		t.Errorf("Monster HP = %v, want %v", monster.HP, expectedMonsterHP)
	}

	// モンスターがプレイヤーを攻撃
	counterDamage := monster.CalculateDamage(player.Defense)
	expectedCounterDamage := 8 - 3 // 5ダメージ
	if counterDamage != expectedCounterDamage {
		t.Errorf("Monster damage = %v, want %v", counterDamage, expectedCounterDamage)
	}

	player.TakeDamage(counterDamage)
	expectedPlayerHP := 20 - 5 // 15HP
	if player.HP != expectedPlayerHP {
		t.Errorf("Player HP = %v, want %v", player.HP, expectedPlayerHP)
	}

	// 両方とも生きているか確認
	if !player.IsAlive() {
		t.Error("Player should be alive")
	}
	if !monster.IsAlive() {
		t.Error("Monster should be alive")
	}

	// プレイヤーが回復
	player.Heal(3)
	expectedPlayerHPAfterHeal := 15 + 3 // 18HP
	if player.HP != expectedPlayerHPAfterHeal {
		t.Errorf("Player HP after heal = %v, want %v", player.HP, expectedPlayerHPAfterHeal)
	}
}
