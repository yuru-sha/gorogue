package magic

import (
	"fmt"
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// EffectResult represents the result of using a magic item
type EffectResult struct {
	Message    string
	Success    bool
	Identified bool // Whether the item should be identified after use
}

// UseScroll applies the effect of a scroll
func UseScroll(scrollName string, player *actor.Player, level *dungeon.Level) *EffectResult {
	switch scrollName {
	case "identify":
		return useScrollOfIdentify(player)
	case "teleportation":
		return useScrollOfTeleportation(player, level)
	case "sleep":
		return useScrollOfSleep(level)
	case "enchant armor":
		return useScrollOfEnchantArmor(player)
	case "enchant weapon":
		return useScrollOfEnchantWeapon(player)
	case "remove curse":
		return useScrollOfRemoveCurse(player)
	case "magic mapping":
		return useScrollOfMagicMapping(level)
	case "light":
		return useScrollOfLight(level)
	case "food detection":
		return useScrollOfDetection(level, "food")
	case "gold detection":
		return useScrollOfDetection(level, "gold")
	case "potion detection":
		return useScrollOfDetection(level, "potion")
	case "monster detection":
		return useScrollOfDetection(level, "monster")
	case "blank paper":
		return &EffectResult{
			Message:    "This scroll is blank.",
			Success:    false,
			Identified: true,
		}
	default:
		return &EffectResult{
			Message:    "Nothing happens.",
			Success:    false,
			Identified: true,
		}
	}
}

// UsePotion applies the effect of a potion
func UsePotion(potionName string, player *actor.Player) *EffectResult {
	switch potionName {
	case "healing":
		return usePotionOfHealing(player, 10)
	case "extra healing":
		return usePotionOfHealing(player, 20)
	case "haste self":
		return usePotionOfHaste(player)
	case "restore strength":
		return usePotionOfRestoreStrength(player)
	case "gain strength":
		return usePotionOfGainStrength(player)
	case "gain experience":
		return usePotionOfGainExperience(player)
	case "see invisible":
		return usePotionOfSeeInvisible(player)
	case "blindness":
		return usePotionOfBlindness(player)
	case "paralysis":
		return usePotionOfParalysis(player)
	case "confusion":
		return usePotionOfConfusion(player)
	case "poison":
		return usePotionOfPoison(player)
	case "thirst quenching":
		return &EffectResult{
			Message:    "You feel refreshed.",
			Success:    true,
			Identified: true,
		}
	default:
		return &EffectResult{
			Message:    "Nothing happens.",
			Success:    false,
			Identified: true,
		}
	}
}

// useScrollOfIdentify identifies an unknown item
func useScrollOfIdentify(player *actor.Player) *EffectResult {
	// TODO: Implement item selection for identification
	return &EffectResult{
		Message:    "You feel enlightened. (TODO: Select item to identify)",
		Success:    true,
		Identified: true,
	}
}

// useScrollOfTeleportation teleports the player to a random location
func useScrollOfTeleportation(player *actor.Player, level *dungeon.Level) *EffectResult {
	// Find a random walkable tile
	for attempts := 0; attempts < 100; attempts++ {
		x := rand.Intn(level.Width)
		y := rand.Intn(level.Height)

		tile := level.GetTile(x, y)
		if tile.Walkable() {
			player.Position.X = x
			player.Position.Y = y
			logger.Debug("Player teleported", "x", x, "y", y)
			return &EffectResult{
				Message:    "You suddenly find yourself somewhere else!",
				Success:    true,
				Identified: true,
			}
		}
	}

	return &EffectResult{
		Message:    "The scroll crumbles to dust.",
		Success:    false,
		Identified: true,
	}
}

// useScrollOfSleep puts nearby monsters to sleep
func useScrollOfSleep(level *dungeon.Level) *EffectResult {
	sleepCount := 0
	for _, monster := range level.Monsters {
		if monster.IsAlive() {
			// Simple sleep effect - monsters skip next few turns
			sleepCount++
		}
	}

	if sleepCount > 0 {
		return &EffectResult{
			Message:    fmt.Sprintf("You hear %d monster(s) yawn.", sleepCount),
			Success:    true,
			Identified: true,
		}
	}

	return &EffectResult{
		Message:    "You hear a faint snoring sound.",
		Success:    true,
		Identified: true,
	}
}

// useScrollOfEnchantArmor enhances equipped armor
func useScrollOfEnchantArmor(player *actor.Player) *EffectResult {
	if player.Equipment.Armor != nil {
		player.Equipment.Armor.Value += 10
		return &EffectResult{
			Message:    "Your armor glows momentarily.",
			Success:    true,
			Identified: true,
		}
	}

	return &EffectResult{
		Message:    "You are not wearing any armor.",
		Success:    false,
		Identified: true,
	}
}

// useScrollOfEnchantWeapon enhances equipped weapon
func useScrollOfEnchantWeapon(player *actor.Player) *EffectResult {
	if player.Equipment.Weapon != nil {
		player.Equipment.Weapon.Value += 10
		return &EffectResult{
			Message:    "Your weapon glows momentarily.",
			Success:    true,
			Identified: true,
		}
	}

	return &EffectResult{
		Message:    "You are not wielding any weapon.",
		Success:    false,
		Identified: true,
	}
}

// useScrollOfRemoveCurse removes curses from items
func useScrollOfRemoveCurse(player *actor.Player) *EffectResult {
	cursedCount := 0

	// Check equipped items
	if player.Equipment.Weapon != nil && player.Equipment.Weapon.IsCursed {
		player.Equipment.Weapon.IsCursed = false
		cursedCount++
	}
	if player.Equipment.Armor != nil && player.Equipment.Armor.IsCursed {
		player.Equipment.Armor.IsCursed = false
		cursedCount++
	}
	if player.Equipment.RingLeft != nil && player.Equipment.RingLeft.IsCursed {
		player.Equipment.RingLeft.IsCursed = false
		cursedCount++
	}
	if player.Equipment.RingRight != nil && player.Equipment.RingRight.IsCursed {
		player.Equipment.RingRight.IsCursed = false
		cursedCount++
	}

	// Check inventory items
	for _, item := range player.Inventory.Items {
		if item.IsCursed {
			item.IsCursed = false
			cursedCount++
		}
	}

	if cursedCount > 0 {
		return &EffectResult{
			Message:    "You feel as if someone is watching over you.",
			Success:    true,
			Identified: true,
		}
	}

	return &EffectResult{
		Message:    "You feel like someone is watching over you.",
		Success:    true,
		Identified: true,
	}
}

// useScrollOfMagicMapping reveals the entire level layout
func useScrollOfMagicMapping(level *dungeon.Level) *EffectResult {
	// TODO: Implement magic mapping effect
	return &EffectResult{
		Message:    "You see the layout of the dungeon flash before your eyes.",
		Success:    true,
		Identified: true,
	}
}

// useScrollOfLight illuminates the area
func useScrollOfLight(level *dungeon.Level) *EffectResult {
	// TODO: Implement light effect
	return &EffectResult{
		Message:    "The dungeon is lit up.",
		Success:    true,
		Identified: true,
	}
}

// useScrollOfDetection detects specific item types
func useScrollOfDetection(level *dungeon.Level, detectType string) *EffectResult {
	count := 0

	switch detectType {
	case "food":
		for _, itm := range level.Items {
			if itm.Type == item.ItemFood {
				count++
			}
		}
	case "gold":
		for _, itm := range level.Items {
			if itm.Type == item.ItemGold {
				count++
			}
		}
	case "potion":
		for _, itm := range level.Items {
			if itm.Type == item.ItemPotion {
				count++
			}
		}
	case "monster":
		for _, monster := range level.Monsters {
			if monster.IsAlive() {
				count++
			}
		}
	}

	if count > 0 {
		return &EffectResult{
			Message:    fmt.Sprintf("You sense %d %s(s) on this level.", count, detectType),
			Success:    true,
			Identified: true,
		}
	}

	return &EffectResult{
		Message:    fmt.Sprintf("You sense no %ss on this level.", detectType),
		Success:    true,
		Identified: true,
	}
}

// usePotionOfHealing restores HP
func usePotionOfHealing(player *actor.Player, amount int) *EffectResult {
	oldHP := player.HP
	player.Heal(amount)
	healedAmount := player.HP - oldHP

	if healedAmount > 0 {
		return &EffectResult{
			Message:    fmt.Sprintf("You feel better. (%d HP restored)", healedAmount),
			Success:    true,
			Identified: true,
		}
	}

	return &EffectResult{
		Message:    "Nothing happens.",
		Success:    false,
		Identified: true,
	}
}

// usePotionOfHaste speeds up the player
func usePotionOfHaste(player *actor.Player) *EffectResult {
	// TODO: Implement haste effect
	return &EffectResult{
		Message:    "You feel yourself moving much faster.",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfRestoreStrength restores lost strength
func usePotionOfRestoreStrength(player *actor.Player) *EffectResult {
	// TODO: Implement strength restoration
	return &EffectResult{
		Message:    "You feel your strength returning.",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfGainStrength permanently increases strength
func usePotionOfGainStrength(player *actor.Player) *EffectResult {
	player.Attack += 2
	return &EffectResult{
		Message:    "You feel stronger!",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfGainExperience grants experience points
func usePotionOfGainExperience(player *actor.Player) *EffectResult {
	expGain := 100 + rand.Intn(200)
	player.GainExp(expGain)
	return &EffectResult{
		Message:    fmt.Sprintf("You feel more experienced! (%d exp)", expGain),
		Success:    true,
		Identified: true,
	}
}

// usePotionOfSeeInvisible grants ability to see invisible creatures
func usePotionOfSeeInvisible(player *actor.Player) *EffectResult {
	// TODO: Implement see invisible effect
	return &EffectResult{
		Message:    "Your eyes tingle.",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfBlindness temporarily blinds the player
func usePotionOfBlindness(player *actor.Player) *EffectResult {
	// TODO: Implement blindness effect
	return &EffectResult{
		Message:    "A cloak of darkness falls around you.",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfParalysis temporarily paralyzes the player
func usePotionOfParalysis(player *actor.Player) *EffectResult {
	// TODO: Implement paralysis effect
	return &EffectResult{
		Message:    "You can't move!",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfConfusion confuses the player
func usePotionOfConfusion(player *actor.Player) *EffectResult {
	// TODO: Implement confusion effect
	return &EffectResult{
		Message:    "Wait, what's going on here? Huh? What? Who?",
		Success:    true,
		Identified: true,
	}
}

// usePotionOfPoison poisons the player
func usePotionOfPoison(player *actor.Player) *EffectResult {
	damage := 3 + rand.Intn(5)
	player.TakeDamage(damage)
	return &EffectResult{
		Message:    fmt.Sprintf("You feel very sick. (%d damage)", damage),
		Success:    true,
		Identified: true,
	}
}
