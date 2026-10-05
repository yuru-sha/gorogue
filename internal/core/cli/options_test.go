package cli

import (
	"strings"
	"testing"

	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestFruitOptionChangesDisplayedFoodName(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 1)
	player.Inventory.AddItem(item.NewItem(0, 0, item.ItemFood, "slime-mold", 0))
	manager := dungeon.NewDungeonManagerWithSeed(player, 1)
	mode := NewCLIModeWithDungeonManager(manager, player)
	mode.IsActive = true
	mode.SetOptions(&config.Options{Fruit: "pear"})
	if result := mode.ExecuteCommand("inventory"); !strings.Contains(result, "pear") || strings.Contains(result, "slime-mold") {
		t.Fatalf("inventory output = %q, want pear instead of slime-mold", result)
	}
}
