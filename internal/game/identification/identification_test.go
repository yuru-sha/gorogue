package identification

import (
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestIdentificationStateRoundTrip(t *testing.T) {
	logger.Setup()
	manager := NewIdentificationManagerWithRand(nil)
	manager.IdentifyItem(item.NewItem(0, 0, item.ItemWand, "Wand of Light", 120))
	manager.IdentifyItem(item.NewItem(0, 0, item.ItemPotion, "healing", 25))

	restored := NewIdentificationManagerWithRand(nil)
	restored.LoadState(manager.SaveState())
	if !restored.IsIdentified(item.NewItem(0, 0, item.ItemWand, "Wand of Light", 120)) {
		t.Fatal("wand identification was not restored")
	}
	if !restored.IsIdentified(item.NewItem(0, 0, item.ItemPotion, "healing", 25)) {
		t.Fatal("potion identification was not restored")
	}
}
