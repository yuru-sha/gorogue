package identification

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// IdentificationManager manages item identification state
type IdentificationManager struct {
	// Global identification state for each item type
	identifiedScrolls map[string]bool
	identifiedPotions map[string]bool
	identifiedRings   map[string]bool
	identifiedWands   map[string]bool

	// Random appearances for unidentified items
	scrollTitles  map[string]string
	potionColors  map[string]string
	ringMaterials map[string]string
	wandMaterials map[string]string

	// Player-supplied names for unidentified source item types.
	calledScrolls map[string]string
	calledPotions map[string]string
	calledRings   map[string]string
	calledWands   map[string]string
}

// DiscoveredItem identifies a source item known by the player.
type DiscoveredItem struct {
	Type item.ItemType
	Name string
}

const (
	appearanceGold   = "gold"
	appearanceSilver = "silver"

	scrollCategory = "scroll"
	potionCategory = "potion"
	ringCategory   = "ring"
	wandCategory   = "wand"
)

// ScrollTitles are random titles for unidentified scrolls
var ScrollTitles = []string{
	"ZELGO MER", "JUYED AWK YACC", "NR 9", "XIXAXA XOXAXA XUXAXA",
	"PRATYAVAYAH", "DAIYEN FOOELS", "LEP GEX VEN ZEA", "PRIRUTSENIE",
	"ELBIB YLOH", "VERR YED HORRE", "VENZAR BORGAVVE", "THARR",
	"YUM YUM", "KERNOD WEL", "ELAM EBOW", "DUAM XNAHT", "ANDOVA BEGARIN",
	"KIRJE", "VE FORBRYDERNE", "CHATCHE", "VELOX NEB", "FOOBIE BLETCH",
	"TEMOV", "GARVEN DEH",
}

// PotionColors contains Rogue 5.4.4's potion-color appearances.
var PotionColors = []string{
	"amber", "aquamarine", "black", "blue", "brown", "clear", "crimson", "cyan",
	"ecru", appearanceGold, "green", "grey", "magenta", "orange", "pink", "plaid",
	"purple", "red", appearanceSilver, "tan", "tangerine", "topaz", "turquoise",
	"vermilion", "violet", "white", "yellow",
}

// RingMaterials contains the source stone and gem appearances.
var RingMaterials = []string{
	"agate", "alexandrite", "amethyst", "carnelian", "diamond", "emerald",
	"germanium", "granite", "garnet", "jade", "kryptonite", "lapis lazuli",
	"moonstone", "obsidian", "onyx", "opal", "pearl", "peridot", "ruby",
	"sapphire", "stibotantalite", "tiger eye", "topaz", "turquoise",
	"taaffeite", "zircon",
}

var wandWoodMaterials = []string{
	"avocado wood", "balsa", "bamboo", "banyan", "birch", "cedar", "cherry",
	"cinnibar", "cypress", "dogwood", "driftwood", "ebony", "elm", "eucalyptus",
	"fall", "hemlock", "holly", "ironwood", "kukui wood", "mahogany", "manzanita",
	"maple", "oaken", "persimmon wood", "pecan", "pine", "poplar", "redwood",
	"rosewood", "spruce", "teak", "walnut", "zebrawood",
}

var wandMetalMaterials = []string{
	"aluminum", "beryllium", "bone", "brass", "bronze", "copper", "electrum",
	appearanceGold, "iron", "lead", "magnesium", "mercury", "nickel", "pewter",
	"platinum", "steel", appearanceSilver, "silicon", "tin", "titanium", "tungsten", "zinc",
}

// WandMaterials contains Rogue's wood and metal appearance vocabulary.
var WandMaterials = []string{
	"avocado wood", "balsa", "bamboo", "banyan", "birch", "cedar", "cherry",
	"cinnibar", "cypress", "dogwood", "driftwood", "ebony", "elm", "eucalyptus",
	"fall", "hemlock", "holly", "ironwood", "kukui wood", "mahogany", "manzanita",
	"maple", "oaken", "persimmon wood", "pecan", "pine", "poplar", "redwood",
	"rosewood", "spruce", "teak", "walnut", "zebrawood", "aluminum", "beryllium",
	"bone", "brass", "bronze", "copper", "electrum", appearanceGold, "iron", "lead",
	"magnesium", "mercury", "nickel", "pewter", "platinum", "steel", appearanceSilver,
	"silicon", "tin", "titanium", "tungsten", "zinc",
}

var scrollSyllables = []string{
	"a", "ab", "ag", "aks", "ala", "an", "app", "arg", "arze", "ash",
	"bek", "bie", "bit", "bjor", "blu", "bot", "bu", "byt", "comp", "con",
	"cos", "cre", "dalf", "dan", "den", "do", "e", "eep", "el", "eng",
	"er", "ere", "erk", "esh", "evs", "fa", "fid", "fri", "fu", "gan",
	"gar", "glen", "gop", "gre", "ha", "hyd", "i", "ing", "ip", "ish",
	"it", "ite", "iv", "jo", "kho", "kli", "klis", "la", "lech", "mar",
	"me", "mi", "mic", "mik", "mon", "mung", "mur", "nej", "nelg", "nep",
	"ner", "nes", "nes", "nih", "nin", "o", "od", "ood", "org", "orn",
	"ox", "oxy", "pay", "ple", "plu", "po", "pot", "prok", "re", "rea",
	"rhov", "ri", "ro", "rog", "rok", "rol", "sa", "san", "sat", "sef",
	"seh", "shu", "ski", "sna", "sne", "snik", "sno", "so", "sol", "sri",
	"sta", "sun", "ta", "tab", "tem",
	//nolint:misspell // "ther" is a canonical Rogue scroll-name syllable.
	"ther", "ti", "tox", "trol", "tue",
	"turs", "u", "ulk", "um", "un", "uni", "ur", "val", "viv", "vly",
	"vom", "wah", "wed", "werg", "wex", "whon", "wun", "xo", "y", "yot",
	"yu", "zant", "zeb", "zim", "zok", "zon", "zum",
}

// NewIdentificationManager creates a new identification manager
func NewIdentificationManager() *IdentificationManager {
	return NewIdentificationManagerWithRand(rand.New(rand.NewSource(time.Now().UnixNano())))
}

// NewIdentificationManagerWithRand creates deterministic unidentified appearances.
func NewIdentificationManagerWithRand(rng *rand.Rand) *IdentificationManager {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	mgr := &IdentificationManager{
		identifiedScrolls: make(map[string]bool),
		identifiedPotions: make(map[string]bool),
		identifiedRings:   make(map[string]bool),
		identifiedWands:   make(map[string]bool),
		scrollTitles:      make(map[string]string),
		potionColors:      make(map[string]string),
		ringMaterials:     make(map[string]string),
		wandMaterials:     make(map[string]string),
		calledScrolls:     make(map[string]string),
		calledPotions:     make(map[string]string),
		calledRings:       make(map[string]string),
		calledWands:       make(map[string]string),
	}

	// Initialize random appearances
	mgr.initializeAppearances(rng)

	return mgr
}

// initializeAppearances sets up source-native randomized item appearances.
func (im *IdentificationManager) initializeAppearances(rng *rand.Rand) {
	for _, definition := range item.ScrollTypes {
		im.scrollTitles[definition.Name] = randomScrollTitle(rng)
	}

	usedColors := make([]bool, len(PotionColors))
	for _, definition := range item.PotionTypes {
		index := rng.Intn(len(PotionColors))
		for usedColors[index] {
			index = rng.Intn(len(PotionColors))
		}
		usedColors[index] = true
		im.potionColors[definition.Name] = PotionColors[index]
	}

	usedStones := make([]bool, len(RingMaterials))
	for _, definition := range item.RingTypes {
		index := rng.Intn(len(RingMaterials))
		for usedStones[index] {
			index = rng.Intn(len(RingMaterials))
		}
		usedStones[index] = true
		im.ringMaterials[definition.Name] = RingMaterials[index]
	}

	usedWood := make([]bool, len(wandWoodMaterials))
	usedMetal := make([]bool, len(wandMetalMaterials))
	for _, definition := range item.WandTypes {
		for {
			if rng.Intn(2) == 0 {
				index := rng.Intn(len(wandMetalMaterials))
				if !usedMetal[index] {
					usedMetal[index] = true
					im.wandMaterials[definition.Name] = "wand:" + wandMetalMaterials[index]
					break
				}
			} else {
				index := rng.Intn(len(wandWoodMaterials))
				if !usedWood[index] {
					usedWood[index] = true
					im.wandMaterials[definition.Name] = "staff:" + wandWoodMaterials[index]
					break
				}
			}
		}
	}

	logger.Debug("Initialized item appearances for identification system")
}

func randomScrollTitle(rng *rand.Rand) string {
	wordCount := rng.Intn(3) + 2
	var title strings.Builder
	for word := range wordCount {
		syllableCount := rng.Intn(3) + 1
		for range syllableCount {
			title.WriteString(scrollSyllables[rng.Intn(len(scrollSyllables))])
		}
		if word+1 < wordCount {
			title.WriteByte(' ')
		}
	}
	return title.String()
}

// GetDisplayName returns the display name for an item (identified or unidentified)
//
//nolint:gocyclo // Item categories have distinct display-name rules.
func (im *IdentificationManager) GetDisplayName(itm *item.Item) string {
	name := itemName(itm)
	switch itm.Type {
	case item.ItemScroll:
		if im.IsIdentified(itm) {
			return fmt.Sprintf("scroll of %s", name)
		}
		if called := im.calledScrolls[name]; called != "" {
			return fmt.Sprintf("scroll called %s", called)
		}
		if title, exists := im.scrollTitles[name]; exists {
			return fmt.Sprintf("scroll titled %q", title)
		}
		return "scroll titled \"UNKNOWN\""

	case item.ItemPotion:
		if im.IsIdentified(itm) {
			return fmt.Sprintf("potion of %s (%s)", name, im.potionColors[name])
		}
		if color, exists := im.potionColors[name]; exists {
			if called := im.calledPotions[name]; called != "" {
				return fmt.Sprintf("potion called %s (%s)", called, color)
			}
			return fmt.Sprintf("%s potion", color)
		}
		return "unknown potion"

	case item.ItemRing:
		if im.IsIdentified(itm) {
			return fmt.Sprintf("ring of %s (%s)", name, im.ringMaterials[name])
		}
		if material, exists := im.ringMaterials[name]; exists {
			if called := im.calledRings[name]; called != "" {
				return fmt.Sprintf("ring called %s (%s)", called, material)
			}
			return fmt.Sprintf("%s ring", material)
		}
		return "unknown ring"

	case item.ItemWand:
		kind, material, exists := strings.Cut(im.wandMaterials[name], ":")
		if !exists {
			return "unknown wand"
		}
		if im.IsIdentified(itm) {
			return fmt.Sprintf("%s of %s (%s)", kind, name, material)
		}
		if called := im.calledWands[name]; called != "" {
			return fmt.Sprintf("%s called %s (%s)", kind, called, material)
		}
		return fmt.Sprintf("%s %s", material, kind)

	case item.ItemWeapon, item.ItemArmor, item.ItemFood:
		return itm.Name
	case item.ItemGold:
		return fmt.Sprintf("%d Gold pieces", itm.Value)
	case item.ItemAmulet:
		return "The Amulet of Yendor"
	default:
		return itm.Name
	}
}

// IsIdentified reports whether this instance or its source type is known.
func (im *IdentificationManager) IsIdentified(itm *item.Item) bool {
	if itm.IsIdentified {
		return true
	}
	name := itemName(itm)
	switch itm.Type {
	case item.ItemScroll:
		return im.identifiedScrolls[name]
	case item.ItemPotion:
		return im.identifiedPotions[name]
	case item.ItemRing:
		return im.identifiedRings[name]
	case item.ItemWand:
		return im.identifiedWands[name]
	case item.ItemWeapon, item.ItemArmor, item.ItemFood, item.ItemGold, item.ItemAmulet:
		return true
	default:
		return true
	}
}

// IdentifyItem marks the selected instance and all items of its source type.
func (im *IdentificationManager) IdentifyItem(itm *item.Item) {
	itm.IsIdentified = true
	name := itemName(itm)
	switch itm.Type {
	case item.ItemScroll:
		im.identifiedScrolls[name] = true
		logger.Debug("Identified scroll", "name", itm.Name)
	case item.ItemPotion:
		im.identifiedPotions[name] = true
		logger.Debug("Identified potion", "name", itm.Name)
	case item.ItemRing:
		im.identifiedRings[name] = true
		logger.Debug("Identified ring", "name", itm.Name)
	case item.ItemWand:
		im.identifiedWands[name] = true
		logger.Debug("Identified wand", "name", itm.Name)
	}
}

func itemName(itm *item.Item) string {
	name := strings.ToLower(strings.TrimSpace(itm.Name))
	if itm.Type == item.ItemWand {
		return wandKey(name)
	}
	return name
}

// SetCall names an unidentified source item type without identifying it.
func (im *IdentificationManager) SetCall(itm *item.Item, callName string) bool {
	if im == nil || itm == nil || im.IsIdentified(itm) {
		return false
	}
	callName = strings.TrimSpace(callName)
	if callName == "" {
		return false
	}
	name := itemName(itm)
	switch itm.Type {
	case item.ItemScroll:
		im.calledScrolls[name] = callName
	case item.ItemPotion:
		im.calledPotions[name] = callName
	case item.ItemRing:
		im.calledRings[name] = callName
	case item.ItemWand:
		im.calledWands[name] = callName
	default:
		return false
	}
	return true
}

// SaveCallState returns stable category-qualified player names.
func (im *IdentificationManager) SaveCallState() map[string]string {
	state := make(map[string]string)
	for name, call := range im.calledScrolls {
		state[scrollCategory+":"+name] = call
	}
	for name, call := range im.calledPotions {
		state[potionCategory+":"+name] = call
	}
	for name, call := range im.calledRings {
		state[ringCategory+":"+name] = call
	}
	for name, call := range im.calledWands {
		state[wandCategory+":"+name] = call
	}
	return state
}

// LoadCallState restores names saved by SaveCallState.
func (im *IdentificationManager) LoadCallState(state map[string]string) {
	for key, call := range state {
		category, name, ok := strings.Cut(key, ":")
		if !ok || name == "" || call == "" {
			continue
		}
		switch category {
		case scrollCategory:
			im.calledScrolls[name] = call
		case potionCategory:
			im.calledPotions[name] = call
		case ringCategory:
			im.calledRings[name] = call
		case wandCategory:
			im.calledWands[name] = call
		}
	}
}

func wandKey(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimPrefix(name, "wand of ")
	return strings.TrimPrefix(name, "staff of ")
}

// IdentifyByUse identifies an item when its effect reveals its source type.
func (im *IdentificationManager) IdentifyByUse(itm *item.Item) {
	if !im.IsIdentified(itm) {
		im.IdentifyItem(itm)
		logger.Info("Item identified by use", "item", itm.Name, "type", itm.Type)
	}
}

// DiscoveredItems returns known source item names in stable Rogue category order.
func (im *IdentificationManager) DiscoveredItems() []DiscoveredItem {
	discovered := make([]DiscoveredItem, 0)
	for _, definition := range item.PotionTypes {
		if im.identifiedPotions[definition.Name] {
			discovered = append(discovered, DiscoveredItem{Type: item.ItemPotion, Name: definition.Name})
		}
	}
	for _, definition := range item.ScrollTypes {
		if im.identifiedScrolls[definition.Name] {
			discovered = append(discovered, DiscoveredItem{Type: item.ItemScroll, Name: definition.Name})
		}
	}
	for _, definition := range item.RingTypes {
		if im.identifiedRings[definition.Name] {
			discovered = append(discovered, DiscoveredItem{Type: item.ItemRing, Name: definition.Name})
		}
	}
	for _, definition := range item.WandTypes {
		if im.identifiedWands[definition.Name] {
			discovered = append(discovered, DiscoveredItem{Type: item.ItemWand, Name: definition.Name})
		}
	}
	sort.Slice(discovered, func(i, j int) bool {
		if discovered[i].Type != discovered[j].Type {
			return discoveryTypeOrder(discovered[i].Type) < discoveryTypeOrder(discovered[j].Type)
		}
		return discovered[i].Name < discovered[j].Name
	})
	return discovered
}

func discoveryTypeOrder(category item.ItemType) int {
	switch category {
	case item.ItemPotion:
		return 0
	case item.ItemScroll:
		return 1
	case item.ItemRing:
		return 2
	default:
		return 3
	}
}

// SaveState returns the identified item types in a stable, category-qualified form.
func (im *IdentificationManager) SaveState() map[string]bool {
	state := make(map[string]bool)
	for name, identified := range im.identifiedScrolls {
		if identified {
			state[scrollCategory+":"+name] = true
		}
	}
	for name, identified := range im.identifiedPotions {
		if identified {
			state[potionCategory+":"+name] = true
		}
	}
	for name, identified := range im.identifiedRings {
		if identified {
			state[ringCategory+":"+name] = true
		}
	}
	for name, identified := range im.identifiedWands {
		if identified {
			state[wandCategory+":"+name] = true
		}
	}
	return state
}

// SaveAppearanceState returns unidentified item appearances in a stable, category-qualified form.
func (im *IdentificationManager) SaveAppearanceState() map[string]string {
	state := make(map[string]string)
	for name, appearance := range im.scrollTitles {
		state[scrollCategory+":"+name] = appearance
	}
	for name, appearance := range im.potionColors {
		state[potionCategory+":"+name] = appearance
	}
	for name, appearance := range im.ringMaterials {
		state[ringCategory+":"+name] = appearance
	}
	for name, appearance := range im.wandMaterials {
		state[wandCategory+":"+name] = appearance
	}
	return state
}

// LoadState restores identified item types saved by SaveState.
func (im *IdentificationManager) LoadState(state map[string]bool) {
	for key, identified := range state {
		if !identified {
			continue
		}
		category, name, ok := strings.Cut(key, ":")
		if !ok {
			continue
		}
		switch category {
		case scrollCategory:
			im.identifiedScrolls[name] = true
		case potionCategory:
			im.identifiedPotions[name] = true
		case ringCategory:
			im.identifiedRings[name] = true
		case wandCategory:
			im.identifiedWands[name] = true
		}
	}
}

// LoadAppearanceState restores unidentified item appearances saved by SaveAppearanceState.
func (im *IdentificationManager) LoadAppearanceState(state map[string]string) {
	for key, appearance := range state {
		category, name, ok := strings.Cut(key, ":")
		if !ok || name == "" || appearance == "" {
			continue
		}
		switch category {
		case scrollCategory:
			im.scrollTitles[name] = appearance
		case potionCategory:
			im.potionColors[name] = appearance
		case ringCategory:
			im.ringMaterials[name] = appearance
		case wandCategory:
			im.wandMaterials[name] = appearance
		}
	}
}

// GetIdentificationScroll creates a scroll of identify
func (im *IdentificationManager) GetIdentificationScroll() *item.Item {
	return item.NewItem(0, 0, item.ItemScroll, "identify", 100)
}
