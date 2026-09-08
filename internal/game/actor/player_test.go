package actor

import "testing"

func TestPlayerEatFoodCapsHunger(t *testing.T) {
	player := NewPlayer(0, 0)
	player.Hunger = 90

	player.EatFood(20)

	if player.Hunger != 100 {
		t.Fatalf("Hunger = %d, want 100", player.Hunger)
	}
}

func TestPlayerAddExpLevelsUp(t *testing.T) {
	player := NewPlayer(0, 0)
	player.TakeDamage(5)

	player.AddExp(10)

	if player.Level != 2 {
		t.Fatalf("Level = %d, want 2", player.Level)
	}
	if player.HP != player.MaxHP {
		t.Fatalf("HP = %d, want max HP %d", player.HP, player.MaxHP)
	}
}
