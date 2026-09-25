package identification

import (
	"math/rand"
	"strings"
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

func TestRogueAppearancesAndIdentification(t *testing.T) {
	logger.Setup()
	manager := NewIdentificationManagerWithRand(rand.New(rand.NewSource(744)))
	if len(PotionColors) != 27 || len(RingMaterials) != 26 || len(WandMaterials) != 55 {
		t.Fatalf("appearance pools have source lengths colors/rings/wands = %d/%d/%d, want 27/26/55", len(PotionColors), len(RingMaterials), len(WandMaterials))
	}

	seenColors := make(map[string]bool)
	for _, definition := range item.PotionTypes {
		potion := item.NewItem(0, 0, item.ItemPotion, definition.Name, definition.Value)
		display := manager.GetDisplayName(potion)
		if !strings.HasSuffix(display, " potion") {
			t.Errorf("unidentified potion %q displayed as %q", definition.Name, display)
		}
		appearance := strings.TrimSuffix(display, " potion")
		if seenColors[appearance] {
			t.Errorf("potion appearance %q was reused", appearance)
		}
		seenColors[appearance] = true
	}
	seenMaterials := make(map[string]bool)
	for _, definition := range item.RingTypes {
		display := manager.GetDisplayName(item.NewItem(0, 0, item.ItemRing, definition.Name, definition.Value))
		if !strings.HasSuffix(display, " ring") {
			t.Errorf("unidentified ring %q displayed as %q", definition.Name, display)
		}
		appearance := strings.TrimSuffix(display, " ring")
		if seenMaterials[appearance] {
			t.Errorf("ring appearance %q was reused", appearance)
		}
		seenMaterials[appearance] = true
	}
	seenMaterials = make(map[string]bool)
	for _, definition := range item.WandTypes {
		display := manager.GetDisplayName(item.NewItem(0, 0, item.ItemWand, definition.Name, definition.Value))
		if !strings.HasSuffix(display, " wand") && !strings.HasSuffix(display, " staff") {
			t.Errorf("unidentified stick %q displayed as %q", definition.Name, display)
		}
		appearance := strings.TrimSuffix(strings.TrimSuffix(display, " wand"), " staff")
		if seenMaterials[appearance] {
			t.Errorf("wand appearance %q was reused", appearance)
		}
		seenMaterials[appearance] = true
	}
	for _, definition := range item.ScrollTypes {
		display := manager.GetDisplayName(item.NewItem(0, 0, item.ItemScroll, definition.Name, definition.Value))
		if !strings.HasPrefix(display, "scroll titled ") || strings.Contains(display, "UNKNOWN") {
			t.Errorf("unidentified scroll %q displayed as %q", definition.Name, display)
		}
	}

	potion := item.NewItem(0, 0, item.ItemPotion, "poison", 5)
	manager.IdentifyByUse(potion)
	if !potion.IsIdentified || !manager.IsIdentified(potion) {
		t.Fatal("using a potion did not identify the instance and its type")
	}
	if got := manager.GetDisplayName(potion); !strings.HasPrefix(got, "potion of poison (") {
		t.Errorf("identified potion name = %q, want source name with retained appearance", got)
	}

	restored := NewIdentificationManagerWithRand(rand.New(rand.NewSource(1)))
	restored.LoadState(manager.SaveState())
	restored.LoadAppearanceState(manager.SaveAppearanceState())
	if got := restored.GetDisplayName(item.NewItem(0, 0, item.ItemPotion, "poison", 5)); got != manager.GetDisplayName(potion) {
		t.Errorf("restored identified potion display = %q, want %q", got, manager.GetDisplayName(potion))
	}
}

func TestIdentifyItemMarksSelectedItemAndItsType(t *testing.T) {
	logger.Setup()
	manager := NewIdentificationManagerWithRand(rand.New(rand.NewSource(9)))
	first := item.NewItem(0, 0, item.ItemScroll, "magic mapping", 150)
	second := item.NewItem(0, 0, item.ItemScroll, "magic mapping", 150)
	manager.IdentifyItem(first)
	if !first.IsIdentified || !manager.IsIdentified(first) || !manager.IsIdentified(second) {
		t.Fatal("explicit identification did not mark the chosen item and its source type")
	}
}

func TestDiscoveredItemsAreSourceGroupedAndStable(t *testing.T) {
	manager := NewIdentificationManagerWithRand(rand.New(rand.NewSource(3)))
	if got := manager.DiscoveredItems(); len(got) != 0 {
		t.Fatalf("new manager reports discovered items: %+v", got)
	}
	for _, discovered := range []struct {
		category item.ItemType
		name     string
	}{
		{item.ItemWand, "light"},
		{item.ItemRing, "protection"},
		{item.ItemScroll, "magic mapping"},
		{item.ItemPotion, "poison"},
		{item.ItemPotion, "healing"},
	} {
		manager.IdentifyItem(item.NewItem(0, 0, discovered.category, discovered.name, 0))
	}

	want := []DiscoveredItem{
		{Type: item.ItemPotion, Name: "healing"},
		{Type: item.ItemPotion, Name: "poison"},
		{Type: item.ItemScroll, Name: "magic mapping"},
		{Type: item.ItemRing, Name: "protection"},
		{Type: item.ItemWand, Name: "light"},
	}
	for i := range want {
		got := manager.DiscoveredItems()
		if got[i] != want[i] {
			t.Errorf("discovered item %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCalledNamesAffectUnidentifiedDisplayAndRoundTrip(t *testing.T) {
	manager := NewIdentificationManagerWithRand(rand.New(rand.NewSource(17)))
	potion := item.NewItem(0, 0, item.ItemPotion, "poison", 5)
	if manager.SetCall(potion, "sick stuff") != true {
		t.Fatal("SetCall rejected an unidentified potion")
	}
	want := "potion called sick stuff (" + manager.potionColors["poison"] + ")"
	if got := manager.GetDisplayName(potion); got != want {
		t.Fatalf("called potion display = %q, want %q", got, want)
	}
	if got := manager.GetDisplayName(item.NewItem(0, 0, item.ItemPotion, "poison", 5)); got != want {
		t.Errorf("another item of the called type displays as %q, want %q", got, want)
	}

	restored := NewIdentificationManagerWithRand(rand.New(rand.NewSource(91)))
	restored.LoadAppearanceState(manager.SaveAppearanceState())
	restored.LoadCallState(manager.SaveCallState())
	if got := restored.GetDisplayName(potion); got != want {
		t.Errorf("restored called potion display = %q, want %q", got, want)
	}
	restored.IdentifyItem(potion)
	if restored.SetCall(potion, "wrong guess") {
		t.Error("SetCall accepted an already identified potion")
	}
}
