package actor

import (
	"testing"

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
		hp, attack, def int
	}{
		{"basic actor", 5, 10, 20, 8, 3},
		{"minimum stats", 0, 0, 1, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewActor(tt.x, tt.y, tt.hp, tt.attack, tt.def)
			if got.Position.X != tt.x || got.Position.Y != tt.y ||
				got.HP != tt.hp || got.MaxHP != tt.hp ||
				got.Attack != tt.attack || got.Defense != tt.def {
				t.Fatalf("NewActor() = %+v, want position=(%d,%d) hp=%d attack=%d defense=%d",
					got, tt.x, tt.y, tt.hp, tt.attack, tt.def)
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
			actor := NewActor(0, 0, 20, 5, 2)
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
			actor := NewActor(0, 0, tt.initialHP, 5, 2)
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
			actor := NewActor(0, 0, tt.maxHP, 5, 2)
			actor.HP = tt.initialHP
			actor.Heal(tt.healAmount)

			if actor.HP != tt.wantHP {
				t.Errorf("Actor.Heal() HP = %v, want %v", actor.HP, tt.wantHP)
			}
		})
	}
}
