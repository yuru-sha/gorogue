package identification

import (
	"fmt"
	"math/rand"
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
}

// ScrollTitles are random titles for unidentified scrolls
var ScrollTitles = []string{
	"ZELGO MER", "JUYED AWK YACC", "NR 9", "XIXAXA XOXAXA XUXAXA",
	"PRATYAVAYAH", "DAIYEN FOOELS", "LEP GEX VEN ZEA", "PRIRUTSENIE",
	"ELBIB YLOH", "VERR YED HORRE", "VENZAR BORGAVVE", "THARR",
	"YUM YUM", "KERNOD WEL", "ELAM EBOW", "DUAM XNAHT", "ANDOVA BEGARIN",
	"KIRJE", "VE FORBRYDERNE", "CHATCHE", "VELOX NEB", "FOOBIE BLETCH",
	"TEMOV", "GARVEN DEH",
}

// PotionColors are random colors for unidentified potions
var PotionColors = []string{
	"red", "blue", "green", "yellow", "black", "brown", "orange", "pink",
	"purple", "white", "clear", "grey", "dark", "light blue", "magenta",
	"amber", "bubbly", "cloudy", "dark green", "dark blue", "emerald",
	"fizzy", "glowing", "golden", "icy", "luminescent", "metallic",
	"milky", "murky", "oily", "puce", "ruby", "silver", "smoky",
	"swirling", "viscous", "ecru", "ochre",
}

// RingMaterials are random materials for unidentified rings
var RingMaterials = []string{
	"wooden", "granite", "opal", "clay", "coral", "black onyx", "moonstone",
	"tiger eye", "jade", "bronze", "agate", "topaz", "sapphire", "ruby",
	"diamond", "pearl", "iron", "brass", "copper", "twisted", "steel",
	"silver", "gold", "ivory", "emerald", "wire", "engagement", "shining",
	"fluorite", "obsidian", "agate", "plastic",
}

// WandMaterials are random materials for unidentified wands
var WandMaterials = []string{
	"glass", "balsa", "crystal", "maple", "pine", "oak", "ebony", "marble",
	"silver", "runed", "long", "short", "bent", "curvy", "twisted", "forked",
	"spiked", "jeweled", "black", "octagonal", "mahogany", "walnut",
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
	}

	// Initialize random appearances
	mgr.initializeAppearances(rng)

	return mgr
}

// initializeAppearances sets up random appearances for items
func (im *IdentificationManager) initializeAppearances(rng *rand.Rand) {
	// Assign random scroll titles
	scrollNames := []string{
		"identify", "teleportation", "sleep", "enchant armor", "enchant weapon",
		"create monster", "remove curse", "aggravate monster", "magic mapping",
		"hold monster", "confuse monster", "scare monster", "blank paper",
		"genocide", "light", "food detection", "gold detection", "potion detection",
		"magic detection", "monster detection", "trap detection", "strength",
		"hit point maximum increase", "monster confusion", "destroy armor",
		"fire", "ice", "charging", "polymorph", "fake",
	}

	shuffledTitles := make([]string, len(ScrollTitles))
	copy(shuffledTitles, ScrollTitles)
	rng.Shuffle(len(shuffledTitles), func(i, j int) {
		shuffledTitles[i], shuffledTitles[j] = shuffledTitles[j], shuffledTitles[i]
	})

	for i, name := range scrollNames {
		if i < len(shuffledTitles) {
			im.scrollTitles[name] = shuffledTitles[i]
		}
	}

	// Assign random potion colors
	potionNames := []string{
		"healing", "extra healing", "haste self", "restore strength", "blindness",
		"paralysis", "confusion", "hallucination", "poison", "gain strength",
		"see invisible", "gain experience", "thirst quenching", "magic detection",
		"monster detection", "object detection", "raise level", "gain dexterity",
		"gain constitution", "gain intelligence", "gain wisdom", "gain charisma",
		"cure disease", "speed", "levitation", "invisibility",
	}

	shuffledColors := make([]string, len(PotionColors))
	copy(shuffledColors, PotionColors)
	rng.Shuffle(len(shuffledColors), func(i, j int) {
		shuffledColors[i], shuffledColors[j] = shuffledColors[j], shuffledColors[i]
	})

	for i, name := range potionNames {
		if i < len(shuffledColors) {
			im.potionColors[name] = shuffledColors[i]
		}
	}

	// Assign random ring materials
	ringNames := []string{
		"protection", "add strength", "sustain strength", "searching", "see invisible",
		"adornment", "teleportation", "stealth", "regeneration", "slow digestion",
		"dexterity", "increase damage", "protection from magic", "hunger",
		"aggravate monster", "maintain armor", "teleport control",
	}

	shuffledMaterials := make([]string, len(RingMaterials))
	copy(shuffledMaterials, RingMaterials)
	rng.Shuffle(len(shuffledMaterials), func(i, j int) {
		shuffledMaterials[i], shuffledMaterials[j] = shuffledMaterials[j], shuffledMaterials[i]
	})

	for i, name := range ringNames {
		if i < len(shuffledMaterials) {
			im.ringMaterials[name] = shuffledMaterials[i]
		}
	}

	// Assign random wand materials
	wandNames := []string{
		"light", "lightning", "fire", "cold", "polymorph", "magic missile",
		"haste monster", "slow monster", "invisibility", "teleportation", "sleep", "drain life",
	}
	shuffledWandMaterials := make([]string, len(WandMaterials))
	copy(shuffledWandMaterials, WandMaterials)
	rng.Shuffle(len(shuffledWandMaterials), func(i, j int) {
		shuffledWandMaterials[i], shuffledWandMaterials[j] = shuffledWandMaterials[j], shuffledWandMaterials[i]
	})

	for i, name := range wandNames {
		if i < len(shuffledWandMaterials) {
			im.wandMaterials[name] = shuffledWandMaterials[i]
		}
	}

	logger.Debug("Initialized item appearances for identification system")
}

// GetDisplayName returns the display name for an item (identified or unidentified)
func (im *IdentificationManager) GetDisplayName(itm *item.Item) string {
	switch itm.Type {
	case item.ItemScroll:
		if im.IsIdentified(itm) {
			return fmt.Sprintf("scroll of %s", itm.Name)
		}
		if title, exists := im.scrollTitles[itm.Name]; exists {
			return fmt.Sprintf("scroll titled %q", title)
		}
		return "scroll titled \"UNKNOWN\""

	case item.ItemPotion:
		if im.IsIdentified(itm) {
			return fmt.Sprintf("potion of %s", itm.Name)
		}
		if color, exists := im.potionColors[itm.Name]; exists {
			return fmt.Sprintf("%s potion", color)
		}
		return "unknown potion"

	case item.ItemRing:
		if im.IsIdentified(itm) {
			return fmt.Sprintf("ring of %s", itm.Name)
		}
		if material, exists := im.ringMaterials[itm.Name]; exists {
			return fmt.Sprintf("%s ring", material)
		}
		return "unknown ring"

	case item.ItemWand:
		if im.IsIdentified(itm) {
			return itm.Name
		}
		if material, exists := im.wandMaterials[wandKey(itm.Name)]; exists {
			return fmt.Sprintf("%s wand", material)
		}
		return "unknown wand"

	case item.ItemWeapon:
		// Weapons are usually identified
		return itm.Name

	case item.ItemArmor:
		// Armor is usually identified
		return itm.Name

	case item.ItemFood:
		// Food is usually identified
		return itm.Name

	case item.ItemGold:
		// Gold is always identified
		return fmt.Sprintf("%d gold pieces", itm.Value)

	case item.ItemAmulet:
		// The Amulet of Yendor is always identified
		return itm.Name

	default:
		return itm.Name
	}
}

// IsIdentified checks if an item type is identified
func (im *IdentificationManager) IsIdentified(itm *item.Item) bool {
	switch itm.Type {
	case item.ItemScroll:
		return im.identifiedScrolls[itm.Name]
	case item.ItemPotion:
		return im.identifiedPotions[itm.Name]
	case item.ItemRing:
		return im.identifiedRings[itm.Name]
	case item.ItemWand:
		return im.identifiedWands[wandKey(itm.Name)]
	case item.ItemWeapon, item.ItemArmor, item.ItemFood, item.ItemGold, item.ItemAmulet:
		// These are always identified
		return true
	default:
		return true
	}
}

// IdentifyItem identifies an item type globally
func (im *IdentificationManager) IdentifyItem(itm *item.Item) {
	switch itm.Type {
	case item.ItemScroll:
		im.identifiedScrolls[itm.Name] = true
		logger.Debug("Identified scroll", "name", itm.Name)
	case item.ItemPotion:
		im.identifiedPotions[itm.Name] = true
		logger.Debug("Identified potion", "name", itm.Name)
	case item.ItemRing:
		im.identifiedRings[itm.Name] = true
		logger.Debug("Identified ring", "name", itm.Name)
	case item.ItemWand:
		im.identifiedWands[wandKey(itm.Name)] = true
		logger.Debug("Identified wand", "name", itm.Name)
	}
}

func wandKey(name string) string {
	return strings.TrimPrefix(strings.ToLower(name), "wand of ")
}

// IdentifyByUse identifies an item when used
func (im *IdentificationManager) IdentifyByUse(itm *item.Item) {
	if !im.IsIdentified(itm) {
		im.IdentifyItem(itm)
		logger.Info("Item identified by use", "item", itm.Name, "type", itm.Type)
	}
}

// SaveState returns the identified item types in a stable, category-qualified form.
func (im *IdentificationManager) SaveState() map[string]bool {
	state := make(map[string]bool)
	for name, identified := range im.identifiedScrolls {
		if identified {
			state["scroll:"+name] = true
		}
	}
	for name, identified := range im.identifiedPotions {
		if identified {
			state["potion:"+name] = true
		}
	}
	for name, identified := range im.identifiedRings {
		if identified {
			state["ring:"+name] = true
		}
	}
	for name, identified := range im.identifiedWands {
		if identified {
			state["wand:"+name] = true
		}
	}
	return state
}

// SaveAppearanceState returns unidentified item appearances in a stable, category-qualified form.
func (im *IdentificationManager) SaveAppearanceState() map[string]string {
	state := make(map[string]string)
	for name, appearance := range im.scrollTitles {
		state["scroll:"+name] = appearance
	}
	for name, appearance := range im.potionColors {
		state["potion:"+name] = appearance
	}
	for name, appearance := range im.ringMaterials {
		state["ring:"+name] = appearance
	}
	for name, appearance := range im.wandMaterials {
		state["wand:"+name] = appearance
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
		case "scroll":
			im.identifiedScrolls[name] = true
		case "potion":
			im.identifiedPotions[name] = true
		case "ring":
			im.identifiedRings[name] = true
		case "wand":
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
		case "scroll":
			im.scrollTitles[name] = appearance
		case "potion":
			im.potionColors[name] = appearance
		case "ring":
			im.ringMaterials[name] = appearance
		case "wand":
			im.wandMaterials[name] = appearance
		}
	}
}

// GetIdentificationScroll creates a scroll of identify
func (im *IdentificationManager) GetIdentificationScroll() *item.Item {
	return item.NewItem(0, 0, item.ItemScroll, "identify", 100)
}
