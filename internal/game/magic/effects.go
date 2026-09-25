package magic

import (
	"fmt"
	"strings"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const invisibilityWandName = "invisibility"

var monsterSymbols = [...]rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

type EffectResult struct {
	Message            string
	Success            bool
	Identified         bool
	NeedsItemSelection bool
	TargetTypes        []item.ItemType
	Experience         int
}

// UseScroll applies a scroll effect when no item target is provided.
func UseScroll(scrollName string, player *actor.Player, level *dungeon.Level) *EffectResult {
	return UseScrollOnItem(scrollName, player, level, nil)
}

// UseScrollOnItem applies a scroll and identifies the selected target when
// reading an identify scroll.
func UseScrollOnItem(scrollName string, player *actor.Player, level *dungeon.Level, target *item.Item) *EffectResult {
	name := strings.ToLower(strings.TrimSpace(scrollName))
	switch name {
	case "identify potion", "identify scroll", "identify weapon", "identify armor", "identify ring, wand or staff":
		return useScrollOfIdentify(scrollName, player, target)
	default:
		if outcome, handled := useScrollEffect(name, player, level); handled {
			return outcome
		}
		return result("That is not a Rogue 5.4.4 scroll.", false, false)
	}
}

func useScrollEffect(name string, player *actor.Player, level *dungeon.Level) (*EffectResult, bool) {
	switch name {
	case "monster confusion":
		player.CanConfuse = true
		return result("Your hands begin to glow red.", true, false), true
	case "teleportation":
		return useScrollOfTeleportation(player, level), true
	case "sleep":
		player.NoCommandTurns += player.RandomSource().Intn(5) + 4
		return result("You fall asleep.", true, true), true
	case "enchant armor":
		return useScrollOfEnchantArmor(player), true
	case "enchant weapon":
		return useScrollOfEnchantWeapon(player), true
	case "create monster":
		return createMonsterNearPlayer(player, level), true
	case "remove curse":
		return useScrollOfRemoveCurse(player), true
	case "magic mapping":
		return useScrollOfMagicMapping(level, player.HallucinationTurns > 0), true
	case "hold monster":
		return useScrollOfHoldMonster(player, level), true
	case "scare monster":
		return result("You hear maniacal laughter in the distance.", true, false), true
	case "food detection":
		return useScrollOfDetection(level, "food"), true
	case "aggravate monsters":
		return useAggravateMonsters(level), true
	case "protect armor":
		if player.Equipment.Armor == nil {
			return result("You feel a strange sense of loss.", true, false), true
		}
		player.Equipment.Armor.IsProtected = true
		return result("Your armor is covered by a shimmering gold shield.", true, false), true
	default:
		return nil, false
	}
}

func useAggravateMonsters(level *dungeon.Level) *EffectResult {
	if level != nil {
		for _, monster := range level.Monsters {
			if monster.IsAlive() {
				monster.IsRunning = true
				monster.IsHeld = false
			}
		}
	}
	return result("You hear a high-pitched humming noise.", true, false)
}

// UsePotion applies a potion effect without level-wide detection information.
func UsePotion(potionName string, player *actor.Player) *EffectResult {
	return UsePotionOnLevel(potionName, player, nil)
}

// UsePotionOnLevel applies a source potion effect with access to level items
// and monsters for detection potions.
func UsePotionOnLevel(potionName string, player *actor.Player, level *dungeon.Level) *EffectResult {
	name := strings.ToLower(strings.TrimSpace(potionName))
	wasHallucinating := player.HallucinationTurns > 0
	if outcome, handled := useBasicPotion(name, player); handled {
		if wasHallucinating && player.HallucinationTurns == 0 && level != nil && player.Position != nil {
			level.UpdateVisibilityForPlayer(player.Position.X, player.Position.Y, false)
		}
		return outcome
	}
	switch name {
	case "monster detection":
		player.MonsterDetectionTurns += 20
		found := level != nil && hasLivingMonsters(level)
		return result(detectionMessage(found, "monsters"), true, false)
	case "magic detection":
		count := countMagicItems(level)
		if count == 0 {
			return result("You have a normal feeling for a moment, then it passes.", true, false)
		}
		return result(fmt.Sprintf("You sense magic on this level (%d items).", count), true, true)
	case "raise level":
		player.GainExp(player.GetExpToNextLevel() + 1)
		return result("You suddenly feel much more skillful.", true, true)
	case "levitation":
		player.LevitationTurns += 30
		return result("You start to float in the air.", true, true)
	default:
		return result("That is not a Rogue 5.4.4 potion.", false, false)
	}
}

func useBasicPotion(name string, player *actor.Player) (*EffectResult, bool) {
	switch name {
	case "healing":
		return usePotionHealing(player, 4, false), true
	case "extra healing":
		return usePotionHealing(player, 8, true), true
	case "haste self":
		return usePotionOfHaste(player), true
	case "restore strength":
		return usePotionOfRestoreStrength(player), true
	case "gain strength":
		return usePotionOfGainStrength(player), true
	case "see invisible":
		return usePotionOfSeeInvisible(player), true
	case "blindness":
		return usePotionOfBlindness(player), true
	case "confusion":
		return usePotionOfConfusion(player), true
	case "hallucination":
		player.HallucinationTurns += 850
		return result("Oh, wow! Everything seems so cosmic!", true, true), true
	case "poison":
		return usePotionOfPoison(player), true
	default:
		return nil, false
	}
}

func result(message string, success, identified bool) *EffectResult {
	return &EffectResult{Message: message, Success: success, Identified: identified}
}

func useScrollOfIdentify(scrollName string, player *actor.Player, target *item.Item) *EffectResult {
	var targets []item.ItemType
	switch strings.ToLower(strings.TrimSpace(scrollName)) {
	case "identify potion":
		targets = []item.ItemType{item.ItemPotion}
	case "identify scroll":
		targets = []item.ItemType{item.ItemScroll}
	case "identify weapon":
		targets = []item.ItemType{item.ItemWeapon}
	case "identify armor":
		targets = []item.ItemType{item.ItemArmor}
	default:
		targets = []item.ItemType{item.ItemRing, item.ItemWand}
	}
	if target == nil {
		return &EffectResult{
			Message:            "Choose an item to identify.",
			Identified:         true,
			Success:            true,
			NeedsItemSelection: true,
			TargetTypes:        targets,
		}
	}
	matches := false
	for _, targetType := range targets {
		matches = matches || target.Type == targetType
	}
	if !matches || player == nil || player.IdentifyMgr == nil {
		return result("That item cannot be identified by this scroll.", false, true)
	}
	player.IdentifyMgr.IdentifyItem(target)
	return result(fmt.Sprintf("You identify the %s.", target.Name), true, true)
}

// useScrollOfTeleportation teleports the player to a random location
func useScrollOfTeleportation(player *actor.Player, level *dungeon.Level) *EffectResult {
	if level == nil {
		return result("The scroll crumbles to dust.", false, true)
	}
	x, y, found := findFreePosition(level, player)
	if !found {
		return result("The scroll crumbles to dust.", false, true)
	}
	previousRoom := roomAt(level, player.Position.X, player.Position.Y)
	player.Held = false
	player.Position.X, player.Position.Y = x, y
	level.UpdateVisibilityForPlayer(x, y, player.HallucinationTurns > 0)
	logger.Debug("Player teleported", "x", x, "y", y)
	return result("You suddenly find yourself somewhere else.", true, roomAt(level, x, y) != previousRoom)
}

// useScrollOfEnchantArmor improves equipped armor and releases its curse.
func useScrollOfEnchantArmor(player *actor.Player) *EffectResult {
	armor := player.Equipment.Armor
	if armor == nil {
		return result("You feel a strange sense of loss.", true, false)
	}
	armor.Enchantment++
	armor.IsCursed = false
	return result("Your armor glows silver for a moment.", true, false)
}

// useScrollOfEnchantWeapon improves an equipped weapon and releases its curse.
func useScrollOfEnchantWeapon(player *actor.Player) *EffectResult {
	weapon := player.Equipment.Weapon
	if weapon == nil {
		return result("You feel a strange sense of loss.", true, false)
	}
	weapon.IsCursed = false
	weapon.Enchantment++
	return result("Your weapon glows blue for a moment.", true, false)
}

// useScrollOfRemoveCurse removes curses from items
func useScrollOfRemoveCurse(player *actor.Player) *EffectResult {
	for _, equipped := range []*item.Item{
		player.Equipment.Weapon,
		player.Equipment.Armor,
		player.Equipment.RingLeft,
		player.Equipment.RingRight,
	} {
		if equipped != nil {
			equipped.IsCursed = false
		}
	}
	return result("You feel as if somebody is watching over you.", true, false)
}

func useScrollOfMagicMapping(level *dungeon.Level, hallucinating bool) *EffectResult {
	if level == nil {
		return result("The map fades before your eyes.", false, true)
	}
	for _, row := range level.Tiles {
		for _, tile := range row {
			if tile != nil {
				tile.Visible = true
				tile.Explored = true
			}
		}
	}
	level.RememberKnownStairs(hallucinating)
	return result("You see the layout of the dungeon flash before your eyes.", true, true)
}

func useScrollOfDetection(level *dungeon.Level, detectType string) *EffectResult {
	if level == nil {
		return result("Your nose tingles.", true, false)
	}
	count := 0
	for _, found := range level.Items {
		if found.Type == item.ItemFood {
			count++
		}
	}
	if count == 0 {
		return result("Your nose tingles.", true, false)
	}
	return result(fmt.Sprintf("Your nose tingles and you smell %d food item(s).", count), true, true)
}

func useScrollOfHoldMonster(player *actor.Player, level *dungeon.Level) *EffectResult {
	if level == nil {
		return result("You feel a strange sense of loss.", true, false)
	}
	count := 0
	for _, monster := range level.Monsters {
		if monster.IsAlive() && monster.IsRunning &&
			abs(monster.Position.X-player.Position.X) <= 2 &&
			abs(monster.Position.Y-player.Position.Y) <= 2 {
			monster.IsRunning = false
			monster.IsHeld = true
			count++
		}
	}
	if count == 0 {
		return result("You feel a strange sense of loss.", true, false)
	}
	return result(fmt.Sprintf("The %d monster(s) around you freeze.", count), true, true)
}

func usePotionHealing(player *actor.Player, sides int, extra bool) *EffectResult {
	healing := rollDice(player, player.Level, sides)
	player.HP += healing
	if player.HP > player.MaxHP {
		if extra && player.HP > player.MaxHP+player.Level+1 {
			player.MaxHP++
		}
		player.MaxHP++
		player.HP = player.MaxHP
	}
	player.BlindTurns = 0
	if extra {
		player.HallucinationTurns = 0
	}
	return result("You begin to feel better.", true, true)
}

func usePotionOfHaste(player *actor.Player) *EffectResult {
	if player.HasteTurns > 0 {
		player.NoCommandTurns += player.RandomSource().Intn(8)
		player.HasteTurns = 0
		return result("You faint from exhaustion.", true, true)
	}
	player.HasteTurns += player.RandomSource().Intn(4) + 4
	return result("You feel yourself moving much faster.", true, true)
}

func usePotionOfRestoreStrength(player *actor.Player) *EffectResult {
	player.RestoreStrength()
	return result("This tastes great and makes you feel warm all over.", true, false)
}

func usePotionOfGainStrength(player *actor.Player) *EffectResult {
	player.AdjustStrength(1)
	return result("You feel stronger now. What bulging muscles!", true, true)
}

func usePotionOfSeeInvisible(player *actor.Player) *EffectResult {
	player.SeeInvisibleTurns += 850
	return result("Your eyes tingle.", true, false)
}

func usePotionOfBlindness(player *actor.Player) *EffectResult {
	player.BlindTurns += 850
	return result("A cloak of darkness falls around you.", true, true)
}

func usePotionOfConfusion(player *actor.Player) *EffectResult {
	wasHallucinating := player.HallucinationTurns > 0
	player.ConfusedTurns += 20
	message := "Wait, what's going on here? Huh? What? Who?"
	if !wasHallucinating {
		message = "What a trippy feeling!"
	}
	return result(message, true, !wasHallucinating)
}

func usePotionOfPoison(player *actor.Player) *EffectResult {
	for _, ring := range []*item.Item{player.Equipment.RingLeft, player.Equipment.RingRight} {
		if ring != nil && ring.Name == "sustain strength" {
			return result("You feel momentarily sick.", true, true)
		}
	}
	loss := player.RandomSource().Intn(3) + 1
	player.ReduceStrength(loss)
	player.HallucinationTurns = 0
	return result(fmt.Sprintf("You feel very sick. Strength falls by %d.", loss), true, true)
}

func rollDice(player *actor.Player, count, sides int) int {
	total := 0
	for range count {
		total += player.RandomSource().Intn(sides) + 1
	}
	return total
}

func hasLivingMonsters(level *dungeon.Level) bool {
	for _, monster := range level.Monsters {
		if monster.IsAlive() {
			return true
		}
	}
	return false
}

func detectionMessage(found bool, noun string) string {
	if found {
		return fmt.Sprintf("You sense the presence of %s.", noun)
	}
	return "You have a strange feeling for a moment, then it passes."
}

func countMagicItems(level *dungeon.Level) int {
	if level == nil {
		return 0
	}
	count := 0
	for _, found := range level.Items {
		switch found.Type {
		case item.ItemPotion, item.ItemScroll, item.ItemRing, item.ItemWand, item.ItemAmulet:
			count++
		case item.ItemWeapon:
			if found.Enchantment != 0 {
				count++
			}
		case item.ItemArmor:
			if found.IsProtected || found.Enchantment != 0 {
				count++
			}
		}
	}
	return count
}

// UseWand applies a charged wand effect along one Rogue direction.
func UseWand(wand *item.Item, player *actor.Player, level *dungeon.Level, dx, dy int) *EffectResult {
	if wand == nil || wand.Type != item.ItemWand || player == nil || level == nil {
		return result("You cannot use that wand here.", false, false)
	}
	if wand.Charges <= 0 {
		return result("Nothing happens.", false, false)
	}
	name := strings.ToLower(strings.TrimSpace(wand.Name))
	name = strings.TrimPrefix(name, "wand of ")
	name = strings.TrimPrefix(name, "staff of ")
	target := monsterAlongRay(player, level, dx, dy)
	if name == "drain life" && player.HP < 2 {
		return result("You are too weak to use it.", false, false)
	}
	outcome := useWandEffect(name, player, level, target, dx, dy)
	if outcome == nil {
		return result("That is not a Rogue 5.4.4 wand.", false, false)
	}
	wand.Charges--
	if outcome.Identified && player.IdentifyMgr != nil {
		player.IdentifyMgr.IdentifyItem(wand)
	}
	if target != nil && target.IsAlive() && outcome.Success {
		target.IsRunning = true
		target.IsHeld = false
	}
	return outcome
}

func useWandEffect(name string, player *actor.Player, level *dungeon.Level, target *actor.Monster, dx, dy int) *EffectResult {
	switch name {
	case "light", "nothing":
		return useLightWand(name, player, level)
	case "drain life":
		return drainLife(player, level)
	case invisibilityWandName, "polymorph":
		return useMonsterWand(name, player, target, level.FloorNumber)
	case "teleport away", "teleport to":
		return useTeleportWand(name, player, level, target, dx, dy)
	case "magic missile":
		outcome := result("The magic missile vanishes with a puff of smoke.", true, true)
		if target != nil {
			outcome.Experience = damageMonster(target, player.RandomSource().Intn(4)+2)
		}
		return outcome
	case "haste monster", "slow monster":
		return useSpeedWand(name, target)
	case "lightning", "fire", "cold":
		outcome := result("The bolt flashes and vanishes.", true, true)
		if target != nil {
			outcome.Experience = damageMonster(target, rollDice(player, 6, 6))
		}
		return outcome
	case "cancellation":
		if target == nil {
			return result("Nothing happens.", true, false)
		}
		target.IsCancelled = true
		target.IsInvisible = false
		target.IsConfused = false
		return result("The monster's powers are canceled.", true, false)
	default:
		return nil
	}
}

func useLightWand(name string, player *actor.Player, level *dungeon.Level) *EffectResult {
	if name == "nothing" {
		return result("Nothing happens.", true, false)
	}
	if level.LightRoomAt(player.Position.X, player.Position.Y) {
		revealCurrentRoom(level, player)
		level.RememberKnownStairs(player.HallucinationTurns > 0)
		return result("The room is lit by a shimmering blue light.", true, true)
	}
	return result("The corridor glows and then fades.", true, true)
}

func useMonsterWand(name string, player *actor.Player, target *actor.Monster, floor int) *EffectResult {
	if target == nil {
		if name == invisibilityWandName {
			return result("You have a tingling feeling for a moment.", true, false)
		}
		return result("Nothing happens.", true, false)
	}
	if name == invisibilityWandName {
		target.IsInvisible = true
		return result("The monster vanishes.", true, false)
	}
	polymorphMonster(target, player, floor)
	return result("The monster changes shape.", true, player.BlindTurns == 0)
}

func useTeleportWand(name string, player *actor.Player, level *dungeon.Level, target *actor.Monster, dx, dy int) *EffectResult {
	if target == nil {
		return result("Nothing happens.", true, false)
	}
	if name == "teleport away" {
		x, y, found := findFreePosition(level, player)
		if !found {
			return result("The monster flickers but stays in place.", true, false)
		}
		target.Position.X, target.Position.Y = x, y
		return result("The monster disappears.", true, false)
	}
	x, y := player.Position.X+dx, player.Position.Y+dy
	if level.IsWalkable(x, y) && level.GetMonsterAt(x, y) == nil {
		target.Position.X, target.Position.Y = x, y
		return result("The monster appears beside you.", true, false)
	}
	return result("Nothing happens.", true, false)
}

func useSpeedWand(name string, target *actor.Monster) *EffectResult {
	if target == nil {
		return result("Nothing happens.", true, false)
	}
	switch name {
	case "haste monster":
		if target.IsSlowed {
			target.IsSlowed = false
		} else {
			target.IsHasted = true
		}
	case "slow monster":
		if target.IsHasted {
			target.IsHasted = false
		} else {
			target.IsSlowed = true
		}
	}
	target.IsRunning = true
	target.IsHeld = false
	return result("The monster's movements change.", true, false)
}

func findFreePosition(level *dungeon.Level, player *actor.Player) (xChoice, yChoice int, found bool) {
	if level == nil {
		return 0, 0, false
	}
	choices := 0
	for y := range level.Height {
		for x := range level.Width {
			if !level.IsWalkable(x, y) ||
				(player != nil && player.Position.X == x && player.Position.Y == y) ||
				level.GetMonsterAt(x, y) != nil || level.GetItemAt(x, y) != nil {
				continue
			}
			choices++
			if player == nil || player.RandomSource().Intn(choices) == 0 {
				xChoice, yChoice, found = x, y, true
			}
		}
	}
	return xChoice, yChoice, found
}

func roomAt(level *dungeon.Level, x, y int) *dungeon.Room {
	if level == nil {
		return nil
	}
	for _, room := range level.Rooms {
		if x >= room.X && x < room.X+room.Width && y >= room.Y && y < room.Y+room.Height {
			return room
		}
	}
	return nil
}

func revealCurrentRoom(level *dungeon.Level, player *actor.Player) {
	room := roomAt(level, player.Position.X, player.Position.Y)
	if room == nil {
		return
	}
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			if tile := level.GetTile(x, y); tile != nil {
				tile.Visible = true
				tile.Explored = true
			}
		}
	}
}

func monsterAlongRay(player *actor.Player, level *dungeon.Level, dx, dy int) *actor.Monster {
	if dx == 0 && dy == 0 {
		return nil
	}
	x, y := player.Position.X, player.Position.Y
	for level.IsInBounds(x+dx, y+dy) {
		x, y = x+dx, y+dy
		if monster := level.GetMonsterAt(x, y); monster != nil {
			return monster
		}
		if !level.IsWalkable(x, y) {
			return nil
		}
	}
	return nil
}

func damageMonster(monster *actor.Monster, damage int) int {
	if monster == nil || !monster.IsAlive() {
		return 0
	}
	monster.TakeDamage(damage)
	if !monster.IsAlive() {
		return monster.Type.Experience
	}
	return 0
}

func drainLife(player *actor.Player, level *dungeon.Level) *EffectResult {
	room := roomAt(level, player.Position.X, player.Position.Y)
	targetCount := 0
	for _, monster := range level.Monsters {
		if monster.IsAlive() && (room != nil && roomAt(level, monster.Position.X, monster.Position.Y) == room ||
			room == nil && max(abs(monster.Position.X-player.Position.X), abs(monster.Position.Y-player.Position.Y)) <= 6) {
			targetCount++
		}
	}
	if targetCount == 0 {
		return result("You have a tingling feeling.", true, false)
	}
	player.HP /= 2
	damage := player.HP / targetCount
	experience := 0
	for _, monster := range level.Monsters {
		if monster.IsAlive() && (room != nil && roomAt(level, monster.Position.X, monster.Position.Y) == room ||
			room == nil && max(abs(monster.Position.X-player.Position.X), abs(monster.Position.Y-player.Position.Y)) <= 6) {
			experience += damageMonster(monster, damage)
			if monster.IsAlive() {
				monster.IsRunning = true
				monster.IsHeld = false
			}
		}
	}
	outcome := result("A wave of life drains from you into the monsters.", true, false)
	outcome.Experience = experience
	return outcome
}

func polymorphMonster(monster *actor.Monster, player *actor.Player, floor int) {
	symbol := monsterSymbols[player.RandomSource().Intn(len(monsterSymbols))]
	replacement := actor.NewMonsterWithRandAndFloor(monster.Position.X, monster.Position.Y, symbol, floor, player.RandomSource())
	if replacement == nil {
		return
	}
	monster.Type = replacement.Type
	monster.HP = replacement.HP
	monster.MaxHP = replacement.MaxHP
	monster.Attack = replacement.Attack
	monster.Defense = replacement.Defense
	monster.Mean = replacement.Mean
	monster.Flying = replacement.Flying
	monster.Greedy = replacement.Greedy
	monster.Regenerates = replacement.Regenerates
	monster.Floor = replacement.Floor
	monster.IsActive = true
	monster.IsRunning = false
	monster.IsHeld = false
	monster.IsConfused = false
	monster.IsInvisible = replacement.IsInvisible
	monster.IsHasted = false
	monster.IsSlowed = false
	monster.IsCancelled = false
}

func createMonsterNearPlayer(player *actor.Player, level *dungeon.Level) *EffectResult {
	if level == nil {
		return result("You hear a faint cry of anguish in the distance.", true, false)
	}
	selectedX, selectedY, choices := 0, 0, 0
	for y := player.Position.Y - 1; y <= player.Position.Y+1; y++ {
		for x := player.Position.X - 1; x <= player.Position.X+1; x++ {
			if (x == player.Position.X && y == player.Position.Y) || !level.IsWalkable(x, y) || level.GetMonsterAt(x, y) != nil {
				continue
			}
			choices++
			if player.RandomSource().Intn(choices) == 0 {
				selectedX, selectedY = x, y
			}
		}
	}
	if choices == 0 {
		return result("You hear a faint cry of anguish in the distance.", true, false)
	}
	var symbol rune
	for {
		candidate := monsterSymbols[player.RandomSource().Intn(len(monsterSymbols))]
		monsterType := actor.MonsterTypes[candidate]
		if level.FloorNumber >= monsterType.MinFloor && level.FloorNumber <= monsterType.MaxFloor {
			symbol = candidate
			break
		}
	}
	monster := actor.NewMonsterWithRandAndFloor(selectedX, selectedY, symbol, level.FloorNumber, player.RandomSource())
	if monster == nil {
		return result("You hear a faint cry of anguish in the distance.", true, false)
	}
	level.Monsters = append(level.Monsters, monster)
	return result("A monster appears nearby.", true, false)
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
