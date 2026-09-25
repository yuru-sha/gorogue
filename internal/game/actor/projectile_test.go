package actor

import (
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestPlayerProjectileAttackUsesThrownDiceAndMatchingLauncher(t *testing.T) {
	tests := []struct {
		name       string
		projectile *item.Item
		launcher   *item.Item
		rolls      []int64
		wantDamage int
	}{
		{
			name:       "shuriken uses two thrown dice",
			projectile: &item.Item{ItemID: 109, Enchantment: 0},
			rolls:      []int64{9, 0, 0},
			wantDamage: 3,
		},
		{
			name:       "arrow uses matching bow and both enchantments",
			projectile: &item.Item{ItemID: 107, Enchantment: 1},
			launcher:   &item.Item{ItemID: 104, Enchantment: 1},
			rolls:      []int64{9, 0, 0},
			wantDamage: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := playerWithRolls(tt.rolls...)
			projectile := *tt.projectile
			projectile.Name = map[int]string{107: "arrow", 109: "shuriken"}[projectile.ItemID]
			if tt.launcher != nil {
				launcher := *tt.launcher
				launcher.Name = "short bow"
				tt.launcher = &launcher
			}

			got := player.ProjectileAttackRoll(10, true, &projectile, tt.launcher)

			if !got.Hit || got.Damage != tt.wantDamage {
				t.Fatalf("projectile attack = %+v, want hit for %d", got, tt.wantDamage)
			}
		})
	}
}
