package identification

import (
	"fmt"
	"math/rand"

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
	mgr.initializeAppearances()

	return mgr
}

// initializeAppearances sets up random appearances for items
func (im *IdentificationManager) initializeAppearances() {
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
	rand.Shuffle(len(shuffledTitles), func(i, j int) {
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
	rand.Shuffle(len(shuffledColors), func(i, j int) {
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
	rand.Shuffle(len(shuffledMaterials), func(i, j int) {
		shuffledMaterials[i], shuffledMaterials[j] = shuffledMaterials[j], shuffledMaterials[i]
	})

	for i, name := range ringNames {
		if i < len(shuffledMaterials) {
			im.ringMaterials[name] = shuffledMaterials[i]
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
	}
}

// IdentifyByUse identifies an item when used
func (im *IdentificationManager) IdentifyByUse(itm *item.Item) {
	if !im.IsIdentified(itm) {
		im.IdentifyItem(itm)
		logger.Info("Item identified by use", "item", itm.Name, "type", itm.Type)
	}
}

// GetIdentificationScroll creates a scroll of identify
func (im *IdentificationManager) GetIdentificationScroll() *item.Item {
	return item.NewItem(0, 0, item.ItemScroll, "identify", 100)
}
