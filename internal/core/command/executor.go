package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuru-sha/gorogue/internal/core/entity"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/magic"
	"github.com/yuru-sha/gorogue/internal/game/save"
)

// Context contains the mutable game state needed to execute a gameplay command.
type Context struct {
	Player   *actor.Player
	Level    *dungeon.Level
	Dungeon  *dungeon.DungeonManager
	Save     *save.SaveGameIntegration
	messages []string
}

// Result describes the observable outcome of a gameplay command.
type Result struct {
	Message            string
	Error              bool
	TurnConsumed       bool
	PlayerDied         bool
	Victory            bool
	NeedsItemSelection bool
	TargetTypes        []gameitem.ItemType
	Player             *actor.Player
	Level              *dungeon.Level
	Dungeon            *dungeon.DungeonManager
}

const (
	equipmentSlotWeapon    = "weapon"
	equipmentSlotArmor     = "armor"
	equipmentSlotRingLeft  = "ring_left"
	equipmentSlotRingRight = "ring_right"
)

// Execute runs a gameplay command for either the GUI or CLI entry point.
//
//nolint:gocyclo // This is the single command dispatch point shared by GUI and CLI.
func Execute(ctx *Context, cmd Command, args ...string) Result {
	if ctx == nil {
		return Result{Message: "Game state is unavailable."}
	}
	if ctx.Player != nil && ctx.Player.NoCommandTurns > 0 {
		advanceMonsters(ctx)
		return turnResult(ctx, "You are unable to act.")
	}
	switch cmd.Type {
	case CmdMoveWest, CmdMoveEast, CmdMoveNorth, CmdMoveSouth,
		CmdMoveNorthWest, CmdMoveNorthEast, CmdMoveSouthWest, CmdMoveSouthEast:
		return executeMove(ctx, cmd.Direction)
	case CmdRun:
		return executeRun(ctx, cmd.Direction)
	case CmdLook:
		return executeLook(ctx, args)
	case CmdInventory:
		return executeInventory(ctx)
	case CmdPickUp:
		return executePickUp(ctx, args)
	case CmdDrop:
		return executeDrop(ctx, args)
	case CmdUse, CmdQuaff, CmdRead, CmdEat:
		return executeItem(ctx, cmd.Type, args)
	case CmdWield, CmdWear, CmdRingOn, CmdEquip:
		return executeEquip(ctx, cmd.Type, args)
	case CmdTakeOff, CmdRingOff, CmdUnequip:
		return executeUnequip(ctx, cmd.Type, args)
	case CmdThrow:
		return executeThrow(ctx, cmd, args)
	case CmdZap:
		return executeZap(ctx, cmd, args)
	case CmdWait:
		return executeWait(ctx, args)
	case CmdSearch:
		return executeSearch(ctx)
	case CmdFindTrap:
		return executeFindTrap(ctx, cmd, args)
	case CmdOpen, CmdClose:
		return executeDoor(ctx, cmd, args)
	case CmdCall:
		return executeCall(ctx, args)
	case CmdCharacter:
		return executeCharacter(ctx)
	case CmdDiscover:
		return executeDiscover(ctx)
	case CmdRepeat:
		return result(ctx, "Repeat the last command from the input interface.")
	case CmdFight:
		return executeFight(ctx, cmd, args)
	case CmdGoUpstairs:
		return executeStairs(ctx, true)
	case CmdGoDownstairs:
		return executeStairs(ctx, false)
	case CmdSave:
		return executeSave(ctx)
	case CmdLoad:
		return executeLoad(ctx)
	default:
		return result(ctx, "Unknown command.")
	}
}

func executeMove(ctx *Context, direction Direction) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	if ctx.Player.NoMoveTurns > 0 || ctx.Player.Held {
		advanceMonsters(ctx)
		return turnResult(ctx, "You are still stuck and cannot move.")
	}

	if ctx.Player.ConfusedTurns > 0 && ctx.Player.RandomSource().Intn(5) != 0 {
		direction = Direction{
			X: ctx.Player.RandomSource().Intn(3) - 1,
			Y: ctx.Player.RandomSource().Intn(3) - 1,
		}
	}
	if direction.X == 0 && direction.Y == 0 {
		advanceMonsters(ctx)
		return turnResult(ctx, "You stumble in place.")
	}

	newX := ctx.Player.Position.X + direction.X
	newY := ctx.Player.Position.Y + direction.Y
	if !ctx.Level.IsInBounds(newX, newY) {
		return result(ctx, "Cannot move out of bounds.")
	}
	tile := ctx.Level.GetTile(newX, newY)
	if tile == nil || !tile.Walkable() {
		return result(ctx, "Cannot move there - blocked.")
	}

	if monster := ctx.Level.GetMonsterAt(newX, newY); monster != nil {
		return executeAttack(ctx, monster)
	}

	oldPosition := *ctx.Player.Position
	ctx.Player.Position.X = newX
	ctx.Player.Position.Y = newY
	ctx.Level.UpdateVisibility(newX, newY)
	return finishPlayerMove(ctx, direction, oldPosition)
}

func finishPlayerMove(ctx *Context, direction Direction, oldPosition entity.Position) Result {
	newX, newY := ctx.Player.Position.X, ctx.Player.Position.Y
	message := fmt.Sprintf("Moved %s to (%d, %d).", directionName(direction), newX, newY)
	if trap := ctx.Level.TrapAt(newX, newY); trap != nil && ctx.Player.LevitationTurns == 0 {
		trap.Discovered = true
		trapMessage, stopMove := triggerTrap(ctx, trap, oldPosition)
		if trapMessage != "" {
			message += "\n" + trapMessage
		}
		if stopMove {
			advanceMonsters(ctx)
			return turnResult(ctx, message)
		}
	}
	if pickupMessage, _ := pickUpAt(ctx, ctx.Player.Position.X, ctx.Player.Position.Y); pickupMessage != "" {
		message += "\n" + pickupMessage
	}
	advanceMonsters(ctx)
	return turnResult(ctx, message)
}

//nolint:gocyclo // One exhaustive switch maps Rogue's eight trap effects to gameplay outcomes.
func triggerTrap(ctx *Context, trap *dungeon.Trap, oldPosition entity.Position) (string, bool) {
	player := ctx.Player
	player.Running = false
	rng := player.RandomSource()
	switch trap.Type {
	case dungeon.TrapDoor:
		if ctx.Dungeon == nil || !ctx.Dungeon.MoveToFloor(ctx.Dungeon.GetCurrentFloor()+1) {
			return "You fell into a trap, but cannot descend further.", true
		}
		ctx.Level = ctx.Dungeon.GetCurrentLevel()
		return "You fell into a trap!", true
	case dungeon.TrapArrow:
		if rogueTrapHits(player, player.Level-1) {
			damage := rng.Intn(6) + 1
			player.TakeDamage(damage)
			return fmt.Sprintf("An arrow hits you for %d damage.", damage), false
		}
		arrow := gameitem.NewItem(oldPosition.X, oldPosition.Y, gameitem.ItemWeapon, "arrow", 1)
		ctx.Level.AddItem(arrow, oldPosition.X, oldPosition.Y)
		return "An arrow shoots past you.", false
	case dungeon.TrapSleep:
		player.NoCommandTurns += 5
		return "A strange white mist envelops you and you fall asleep.", false
	case dungeon.TrapBear:
		player.NoMoveTurns += 3
		return "You are caught in a bear trap.", false
	case dungeon.TrapTeleport:
		if teleportPlayer(ctx) {
			return "You are suddenly elsewhere.", true
		}
		return "The trap flickers, but you remain in place.", true
	case dungeon.TrapDart:
		if !rogueTrapHits(player, player.Level+1) {
			return "A small dart whizzes past your ear and vanishes.", false
		}
		damage := rng.Intn(4) + 1
		player.TakeDamage(damage)
		if player.IsAlive() && !wearingRing(player, "sustain strength") && !player.SaveAgainst(0) {
			player.ReduceStrength(1)
		}
		return fmt.Sprintf("A poisoned dart hits you for %d damage.", damage), false
	case dungeon.TrapRust:
		player.RustArmor()
		return "A gush of water hits you on the head.", false
	case dungeon.TrapMystery:
		messages := [...]string{
			"You are suddenly in a parallel dimension.",
			"The light in here suddenly seems strange.",
			"You feel a sting in the side of your neck.",
			"Multicolored lines swirl around you, then fade.",
			"A strange light flashes in your eyes.",
			"A spike shoots past your ear!",
			"Sparks dance across your armor.",
			"You suddenly feel very thirsty.",
			"You feel time speed up suddenly.",
			"Time now seems to be going slower.",
			"Your pack turns strange colors!",
		}
		return messages[rng.Intn(len(messages))], false
	default:
		return "The trap clicks harmlessly.", false
	}
}

func rogueTrapHits(player *actor.Player, attackerLevel int) bool {
	return player.RandomSource().Intn(20)+2 >= 20-attackerLevel-player.GetTotalDefense()
}

func wearingRing(player *actor.Player, name string) bool {
	for _, ring := range []*gameitem.Item{player.Equipment.RingLeft, player.Equipment.RingRight} {
		if ring != nil && strings.EqualFold(ring.RealName, name) {
			return true
		}
	}
	return false
}
func sourceRingName(ring *gameitem.Item) string {
	if ring == nil {
		return ""
	}
	name := ring.RealName
	if name == "" {
		name = ring.Name
	}
	return strings.ToLower(strings.TrimSpace(name))
}

func teleportPlayer(ctx *Context) bool {
	player, level := ctx.Player, ctx.Level
	rng := player.RandomSource()
	for range level.Width * level.Height {
		x := rng.Intn(level.Width)
		y := rng.Intn(level.Height)
		tile := level.GetTile(x, y)
		if tile == nil || !tile.Walkable() || (x == player.Position.X && y == player.Position.Y) ||
			level.GetMonsterAt(x, y) != nil {
			continue
		}
		player.Position.X, player.Position.Y = x, y
		level.UpdateVisibility(x, y)
		return true
	}
	return false
}

func executeRun(ctx *Context, direction Direction) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	if direction.X == 0 && direction.Y == 0 {
		return result(ctx, "Invalid movement direction.")
	}

	ctx.Player.Running = true
	defer func() {
		ctx.Player.Running = false
	}()
	steps := max(ctx.Level.Width, ctx.Level.Height)
	moved := false
	for range steps {
		x := ctx.Player.Position.X + direction.X
		y := ctx.Player.Position.Y + direction.Y
		tile := ctx.Level.GetTile(x, y)
		if tile == nil || !tile.Walkable() {
			break
		}
		if ctx.Level.GetMonsterAt(x, y) != nil || ctx.Level.GetItemAt(x, y) != nil {
			break
		}
		step := executeMove(ctx, direction)
		if !step.TurnConsumed {
			if moved {
				return turnResult(ctx, "You stop running.")
			}
			return step
		}
		moved = true
		if step.PlayerDied {
			return step
		}
		if !ctx.Player.Running {
			break
		}
	}
	if !moved {
		return result(ctx, "You cannot run that way.")
	}
	return turnResult(ctx, "You stop running.")
}

func executeFight(ctx *Context, cmd Command, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}

	var monster *actor.Monster
	if len(args) == 2 {
		x, errX := strconv.Atoi(args[0])
		y, errY := strconv.Atoi(args[1])
		if errX != nil || errY != nil {
			return result(ctx, "Invalid coordinates.")
		}
		if !adjacent(ctx.Player.Position.X, ctx.Player.Position.Y, x, y) {
			return result(ctx, "You can only attack an adjacent monster.")
		}
		monster = ctx.Level.GetMonsterAt(x, y)
	} else {
		direction, ok := resolveDirection(cmd, args)
		if !ok {
			return result(ctx, "Attack which direction?")
		}
		monster = ctx.Level.GetMonsterAt(
			ctx.Player.Position.X+direction.X,
			ctx.Player.Position.Y+direction.Y,
		)
	}
	if monster == nil {
		return result(ctx, "There is no monster there.")
	}
	return executeAttack(ctx, monster)
}

func executeAttack(ctx *Context, monster *actor.Monster) Result {
	attack := ctx.Player.AttackRoll(monster.Defense, monster.IsRunning)
	monster.IsRunning = true
	monster.IsHeld = false
	if !attack.Hit {
		advanceMonsters(ctx)
		return turnResult(ctx, fmt.Sprintf("You miss %s.", monster.Type.Name))
	}
	monster.TakeDamage(attack.Damage)
	if attack.Confuses {
		monster.IsConfused = true
	}
	if monster.IsAlive() {
		advanceMonsters(ctx)
		return turnResult(ctx, fmt.Sprintf("Attacked %s for %d damage! (%d HP remaining)",
			monster.Type.Name, attack.Damage, monster.HP))
	}

	exp := monster.Type.Experience
	ctx.Player.GainExp(exp)
	advanceMonsters(ctx)
	return turnResult(ctx, fmt.Sprintf("Killed %s! Gained %d experience.",
		monster.Type.Name, exp))
}

func executePickUp(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}

	items := make([]*gameitem.Item, 0)
	for _, candidate := range ctx.Level.Items {
		if candidate.Position.X == ctx.Player.Position.X && candidate.Position.Y == ctx.Player.Position.Y {
			items = append(items, candidate)
		}
	}
	if len(items) == 0 {
		return result(ctx, "There is nothing here to pick up.")
	}

	if len(args) > 0 && strings.EqualFold(args[0], "all") {
		pickedUp := 0
		for _, candidate := range items {
			if ctx.Player.Inventory.AddItem(candidate) {
				ctx.Level.RemoveItem(candidate)
				pickedUp++
			}
		}
		if pickedUp > 0 {
			advanceMonsters(ctx)
			return turnResult(ctx, fmt.Sprintf("Picked up %d items.", pickedUp))
		}
		return result(ctx, "Picked up 0 items.")
	}

	message, pickedUp := pickUpAt(ctx, ctx.Player.Position.X, ctx.Player.Position.Y)
	if !pickedUp {
		return result(ctx, message)
	}
	advanceMonsters(ctx)
	return turnResult(ctx, message)
}

func pickUpAt(ctx *Context, x, y int) (string, bool) {
	item := ctx.Level.GetItemAt(x, y)
	if item == nil {
		return "", false
	}
	if !ctx.Player.Inventory.AddItem(item) {
		return "Your pack is full!", false
	}

	ctx.Level.RemoveItem(item)
	displayName := ctx.Player.IdentifyMgr.GetDisplayName(item)
	switch item.Type {
	case gameitem.ItemGold:
		return fmt.Sprintf("You found %d gold pieces.", item.Value), true
	case gameitem.ItemAmulet:
		return fmt.Sprintf("You picked up the %s!", displayName), true
	default:
		return fmt.Sprintf("You picked up %s.", displayName), true
	}
}

func executeDrop(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	index, item, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}

	ctx.Level.AddItem(item, ctx.Player.Position.X, ctx.Player.Position.Y)
	ctx.Player.Inventory.RemoveItem(index)
	advanceMonsters(ctx)
	return turnResult(ctx, fmt.Sprintf("Dropped %s.", ctx.Player.IdentifyMgr.GetDisplayName(item)))
}

func executeEquip(ctx *Context, action Type, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	index, gameItem, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}

	if !canEquipWithAction(gameItem, action) {
		return result(ctx, "That item cannot be used for this equipment action.")
	}
	if gameItem.Type == gameitem.ItemRing &&
		ctx.Player.Equipment.RingLeft != nil && ctx.Player.Equipment.RingRight != nil {
		return result(ctx, "Both ring hands are occupied.")
	}

	var oldItem *gameitem.Item
	var slot string
	switch gameItem.Type {
	case gameitem.ItemWeapon:
		oldItem, slot = ctx.Player.Equipment.Weapon, equipmentSlotWeapon
	case gameitem.ItemArmor:
		oldItem, slot = ctx.Player.Equipment.Armor, equipmentSlotArmor
	}
	if oldItem != nil && oldItem.IsCursed {
		return result(ctx, "You can't replace your cursed equipment.")
	}

	if message := equipInventoryItem(ctx, index, gameItem, oldItem, slot); message != "" {
		return result(ctx, message)
	}
	if gameItem.Type == gameitem.ItemRing && sourceRingName(gameItem) == "add strength" {
		ctx.Player.AdjustStrength(gameItem.Enchantment)
	}

	advanceMonsters(ctx)
	actionName := equipmentActionName(action, gameItem.Type)
	return turnResult(ctx, fmt.Sprintf("%s %s.", actionName, ctx.Player.IdentifyMgr.GetDisplayName(gameItem)))
}

func equipInventoryItem(ctx *Context, index int, gameItem, oldItem *gameitem.Item, slot string) string {
	ctx.Player.Inventory.RemoveItem(index)
	if oldItem != nil {
		ctx.Player.Equipment.UnequipItem(slot)
		if !ctx.Player.Inventory.AddItem(oldItem) {
			ctx.Player.Inventory.AddItem(gameItem)
			ctx.Player.Equipment.EquipItem(oldItem)
			return "Inventory is full! Cannot replace equipment."
		}
	}
	if !ctx.Player.Equipment.EquipItem(gameItem) {
		if oldItem != nil {
			ctx.Player.Inventory.RemoveItem(len(ctx.Player.Inventory.Items) - 1)
			ctx.Player.Equipment.EquipItem(oldItem)
		}
		ctx.Player.Inventory.AddItem(gameItem)
		return "Cannot equip that item."
	}
	return ""
}

func canEquipWithAction(gameItem *gameitem.Item, action Type) bool {
	switch action {
	case CmdWield:
		return gameItem.Type == gameitem.ItemWeapon
	case CmdWear:
		return gameItem.Type == gameitem.ItemArmor
	case CmdRingOn:
		return gameItem.Type == gameitem.ItemRing
	default:
		return gameItem.Type == gameitem.ItemWeapon || gameItem.Type == gameitem.ItemArmor || gameItem.Type == gameitem.ItemRing
	}
}

func equipmentActionName(action Type, itemType gameitem.ItemType) string {
	switch action {
	case CmdWield:
		return "Wielded"
	case CmdWear:
		return "Wearing"
	case CmdRingOn:
		return "Put on"
	default:
		switch itemType {
		case gameitem.ItemWeapon:
			return "Wielded"
		case gameitem.ItemArmor:
			return "Wearing"
		case gameitem.ItemRing:
			return "Put on"
		default:
			return "Equipped"
		}
	}
}

func executeUnequip(ctx *Context, action Type, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	if len(args) == 0 {
		return result(ctx, "Select an equipment slot.")
	}

	slot := strings.ToLower(args[0])
	slotName := map[string]string{
		"weapon":     equipmentSlotWeapon,
		"w":          equipmentSlotWeapon,
		"armor":      equipmentSlotArmor,
		"a":          equipmentSlotArmor,
		"ring-left":  equipmentSlotRingLeft,
		"ring_left":  equipmentSlotRingLeft,
		"left":       equipmentSlotRingLeft,
		"l":          equipmentSlotRingLeft,
		"ring-right": equipmentSlotRingRight,
		"ring_right": equipmentSlotRingRight,
		"right":      equipmentSlotRingRight,
		"r":          equipmentSlotRingRight,
	}[slot]
	if slotName == "" {
		return result(ctx, "Unknown equipment slot.")
	}
	if action == CmdTakeOff && slotName != equipmentSlotArmor {
		return result(ctx, "T only takes off armor.")
	}
	if action == CmdRingOff && slotName != equipmentSlotRingLeft && slotName != equipmentSlotRingRight {
		return result(ctx, "R only removes a ring.")
	}
	if equipped := equippedItem(ctx.Player, slotName); equipped != nil && equipped.IsCursed {
		return result(ctx, fmt.Sprintf("You can't remove the cursed %s.", slot))
	}

	gameItem := ctx.Player.Equipment.UnequipItem(slotName)
	if gameItem == nil {
		return result(ctx, fmt.Sprintf("No item equipped in %s slot.", slot))
	}
	if !ctx.Player.Inventory.AddItem(gameItem) {
		ctx.Player.Equipment.EquipItem(gameItem)
		return result(ctx, "Inventory is full! Cannot unequip.")
	}
	if gameItem.Type == gameitem.ItemRing && sourceRingName(gameItem) == "add strength" {
		ctx.Player.AdjustStrength(-gameItem.Enchantment)
	}
	advanceMonsters(ctx)
	return turnResult(ctx, fmt.Sprintf("Removed %s.", ctx.Player.IdentifyMgr.GetDisplayName(gameItem)))
}

//nolint:gocyclo // Item command validation stays in the shared dispatch path.
func executeItem(ctx *Context, commandType Type, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	index, gameItem, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}

	var message string
	var identified bool
	turnConsumed := true
	switch commandType {
	case CmdUse:
		switch gameItem.Type {
		case gameitem.ItemPotion:
			effect := magic.UsePotionOnLevel(gameItem.Name, ctx.Player, ctx.Level)
			message, identified = effect.Message, effect.Identified
		case gameitem.ItemScroll:
			var target *gameitem.Item
			if len(args) > 1 {
				_, target, ok = inventoryItem(ctx, args[1:])
				if !ok {
					return result(ctx, inventoryItemError(args[1:]))
				}
			}
			effect := magic.UseScrollOnItem(gameItem.Name, ctx.Player, ctx.Level, target)
			if effect.NeedsItemSelection {
				return Result{
					Message:            effect.Message,
					NeedsItemSelection: true,
					TargetTypes:        effect.TargetTypes,
					Player:             ctx.Player,
					Level:              ctx.Level,
					Dungeon:            ctx.Dungeon,
				}
			}
			if !effect.Success {
				return failure(ctx, effect.Message)
			}
			message, identified = effect.Message, effect.Identified
		case gameitem.ItemFood:
			ctx.Player.EatFood(gameItem.Value)
			message = fmt.Sprintf("You ate %s.", ctx.Player.IdentifyMgr.GetDisplayName(gameItem))
			identified = true
		default:
			return result(ctx, "That item cannot be used.")
		}
	case CmdQuaff:
		if gameItem.Type != gameitem.ItemPotion {
			return result(ctx, "You can't drink that!")
		}
		effect := magic.UsePotionOnLevel(gameItem.Name, ctx.Player, ctx.Level)
		message, identified = effect.Message, effect.Identified
	case CmdRead:
		if gameItem.Type != gameitem.ItemScroll {
			return result(ctx, "You can't read that!")
		}
		var target *gameitem.Item
		if len(args) > 1 {
			_, target, ok = inventoryItem(ctx, args[1:])
			if !ok {
				return result(ctx, inventoryItemError(args[1:]))
			}
		}
		effect := magic.UseScrollOnItem(gameItem.Name, ctx.Player, ctx.Level, target)
		if effect.NeedsItemSelection {
			return Result{
				Message:            effect.Message,
				NeedsItemSelection: true,
				TargetTypes:        effect.TargetTypes,
				Player:             ctx.Player,
				Level:              ctx.Level,
				Dungeon:            ctx.Dungeon,
			}
		}
		if !effect.Success {
			return failure(ctx, effect.Message)
		}
		message, identified = effect.Message, effect.Identified
	case CmdEat:
		if gameItem.Type != gameitem.ItemFood {
			return result(ctx, "You can't eat that!")
		}
		ctx.Player.EatFood(gameItem.Value)
		message = fmt.Sprintf("You ate %s.", ctx.Player.IdentifyMgr.GetDisplayName(gameItem))
		identified = true
	}

	if identified {
		ctx.Player.IdentifyMgr.IdentifyByUse(gameItem)
	}
	ctx.Player.Inventory.RemoveItem(index)
	if turnConsumed {
		advanceMonsters(ctx)
	}
	if turnConsumed {
		return turnResult(ctx, message)
	}
	return result(ctx, message)
}

func executeLook(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	x, y := ctx.Player.Position.X, ctx.Player.Position.Y
	if len(args) == 2 {
		var err error
		x, err = strconv.Atoi(args[0])
		if err != nil {
			return result(ctx, "Invalid coordinates.")
		}
		y, err = strconv.Atoi(args[1])
		if err != nil {
			return result(ctx, "Invalid coordinates.")
		}
	} else if len(args) != 0 {
		return result(ctx, "Usage: look [x] [y]")
	}
	if !ctx.Level.IsInBounds(x, y) {
		return result(ctx, "Out of bounds.")
	}
	tile := ctx.Level.GetTile(x, y)
	if tile == nil || (!tile.Visible && !tile.Explored) {
		return result(ctx, fmt.Sprintf("Position (%d, %d):\nYou cannot see that location.", x, y))
	}

	lines := []string{fmt.Sprintf("Position (%d, %d):", x, y)}
	lines = append(lines, fmt.Sprintf("Terrain: %s", tile.Type.String()))
	if tile.Visible {
		lines = appendVisibleLookDetails(ctx, lines, x, y)
	}
	return result(ctx, strings.Join(lines, "\n"))
}

func appendVisibleLookDetails(ctx *Context, lines []string, x, y int) []string {
	if monster := ctx.Level.GetMonsterAt(x, y); monster != nil {
		lines = append(lines, fmt.Sprintf("Monster: %s (HP: %d/%d)", monster.Type.Name, monster.HP, monster.MaxHP))
	}
	items := make([]string, 0)
	for _, item := range ctx.Level.Items {
		if item.Position.X == x && item.Position.Y == y {
			items = append(items, ctx.Player.IdentifyMgr.GetDisplayName(item))
		}
	}
	if len(items) > 0 {
		lines = append(lines, "Items: "+strings.Join(items, ", "))
	}
	return lines
}

func executeInventory(ctx *Context) Result {
	if ctx == nil || ctx.Player == nil {
		return result(ctx, "Game state is unavailable.")
	}
	return result(ctx, strings.Join(ctx.Player.Inventory.GetInventoryListing(ctx.Player.IdentifyMgr), "\n"))
}

func executeWait(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	turns := 1
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil || parsed <= 0 {
			return result(ctx, "Invalid rest turns.")
		}
		turns = parsed
	}
	if turns > 100 {
		turns = 100
	}
	for turn := range turns {
		advanceMonsters(ctx)
		if !ctx.Player.IsAlive() {
			turns = turn + 1
			break
		}
	}
	if len(args) == 0 {
		return turnResult(ctx, "You rest.")
	}
	return turnResult(ctx, fmt.Sprintf("Rested for %d turns.", turns))
}

func executeSearch(ctx *Context) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	blind := ctx.Player.BlindTurns > 0
	hallucinating := ctx.Player.HallucinationTurns > 0
	foundTraps := ctx.Level.SearchTraps(ctx.Player.Position.X, ctx.Player.Position.Y, blind, hallucinating)
	foundSecrets := ctx.Level.SearchSecrets(ctx.Player.Position.X, ctx.Player.Position.Y, blind, hallucinating)
	var message strings.Builder
	message.WriteString("You search the area.")
	for _, trap := range foundTraps {
		message.WriteString("\nYou found a ")
		message.WriteString(trap.Type.String())
		message.WriteByte('.')
	}
	if foundSecrets > 0 {
		message.WriteString("\nYou discover a hidden door or passage.")
	}
	advanceMonsters(ctx)
	return turnResult(ctx, message.String())
}

func executeFindTrap(ctx *Context, cmd Command, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	direction, ok := resolveDirection(cmd, args)
	if !ok {
		return result(ctx, "Check for a trap in which direction?")
	}
	x := ctx.Player.Position.X + direction.X
	y := ctx.Player.Position.Y + direction.Y
	trap := ctx.Level.TrapAt(x, y)
	if trap == nil || !trap.Discovered {
		return result(ctx, "You don't see a trap there.")
	}
	return result(ctx, "You found a "+trap.Type.String()+".")
}

func executeCall(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	if len(args) < 2 {
		return result(ctx, "Usage: call <item_letter> <name>")
	}
	_, gameItem, ok := inventoryItem(ctx, args[:1])
	if !ok {
		return result(ctx, inventoryItemError(args[:1]))
	}
	name := strings.TrimSpace(strings.Join(args[1:], " "))
	if !ctx.Player.IdentifyMgr.SetCall(gameItem, name) {
		return result(ctx, "That item cannot be called.")
	}
	return result(ctx, "You name it "+name+".")
}

func executeDoor(ctx *Context, cmd Command, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	direction, ok := resolveDirection(cmd, args)
	if !ok {
		return result(ctx, "Which direction?")
	}
	targetX := ctx.Player.Position.X + direction.X
	targetY := ctx.Player.Position.Y + direction.Y
	tile := ctx.Level.GetTile(targetX, targetY)
	if tile == nil {
		return result(ctx, "There is no door there.")
	}

	if cmd.Type == CmdOpen {
		switch tile.Type {
		case dungeon.TileDoor, dungeon.TileDoorClosed:
			ctx.Level.SetTile(targetX, targetY, dungeon.TileOpenDoor)
			ctx.Level.UpdateVisibility(ctx.Player.Position.X, ctx.Player.Position.Y)
			advanceMonsters(ctx)
			return turnResult(ctx, "You open the door.")
		case dungeon.TileDoorOpen, dungeon.TileOpenDoor:
			return result(ctx, "The door is already open.")
		default:
			return result(ctx, "There is no door there.")
		}
	}

	switch tile.Type {
	case dungeon.TileDoorOpen, dungeon.TileOpenDoor:
		if ctx.Level.GetMonsterAt(targetX, targetY) != nil {
			return result(ctx, "There's a monster in the doorway!")
		}
		for _, item := range ctx.Level.Items {
			if item.Position.X == targetX && item.Position.Y == targetY {
				return result(ctx, "There's something in the doorway!")
			}
		}
		ctx.Level.SetTile(targetX, targetY, dungeon.TileDoor)
		ctx.Level.UpdateVisibility(ctx.Player.Position.X, ctx.Player.Position.Y)
		advanceMonsters(ctx)
		return turnResult(ctx, "You close the door.")
	case dungeon.TileDoor, dungeon.TileDoorClosed:
		return result(ctx, "The door is already closed.")
	default:
		return result(ctx, "There is no door there.")
	}
}

func executeStairs(ctx *Context, goUp bool) Result {
	if ctx == nil || ctx.Dungeon == nil {
		return result(ctx, "Dungeon navigation is unavailable.")
	}
	if goUp {
		if !ctx.Dungeon.CanGoUpstairs() || !ctx.Dungeon.GoUpstairs() {
			return result(ctx, "There are no usable up stairs here.")
		}
		ctx.Level = ctx.Dungeon.GetCurrentLevel()
		if ctx.Dungeon.CheckVictoryCondition() {
			return victoryResult(ctx, "You returned to the surface and won.")
		}
		advanceMonsters(ctx)
		return turnResult(ctx, fmt.Sprintf("Climbed to floor %d.", ctx.Dungeon.GetCurrentFloor()))
	}

	if !ctx.Dungeon.CanGoDownstairs() || !ctx.Dungeon.GoDownstairs() {
		return result(ctx, "There are no usable down stairs here.")
	}
	ctx.Level = ctx.Dungeon.GetCurrentLevel()
	if ctx.Dungeon.IsOnFinalFloor() {
		ctx.Dungeon.PlaceAmuletOfYendor()
	}
	advanceMonsters(ctx)
	return turnResult(ctx, fmt.Sprintf("Descended to floor %d.", ctx.Dungeon.GetCurrentFloor()))
}

func executeSave(ctx *Context) Result {
	if ctx == nil || ctx.Save == nil {
		return failure(ctx, "Save/load is unavailable.")
	}
	if ctx.Player != nil && ctx.Dungeon != nil {
		ctx.Save.SetGameState(ctx.Player, ctx.Dungeon)
	}
	if err := ctx.Save.SaveGame(); err != nil {
		return failure(ctx, fmt.Sprintf("Save failed: %v", err))
	}
	return result(ctx, "Game saved.")
}

func executeLoad(ctx *Context) Result {
	if ctx == nil || ctx.Save == nil {
		return failure(ctx, "Save/load is unavailable.")
	}
	if err := ctx.Save.LoadGame(); err != nil {
		return failure(ctx, fmt.Sprintf("Load failed: %v", err))
	}
	ctx.Player, ctx.Dungeon = ctx.Save.GetGameState()
	if ctx.Dungeon != nil {
		ctx.Level = ctx.Dungeon.GetCurrentLevel()
	}
	return result(ctx, "Game loaded.")
}

func inventoryItem(ctx *Context, args []string) (int, *gameitem.Item, bool) {
	if len(args) == 0 {
		return -1, nil, false
	}
	letter := []rune(strings.ToLower(args[0]))
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		return -1, nil, false
	}
	index, item := ctx.Player.Inventory.GetItemByLetter(letter[0])
	if item == nil {
		return -1, nil, false
	}
	return index, item, true
}

func inventoryItemError(args []string) string {
	if len(args) == 0 {
		return "Usage: <item_letter>"
	}
	letter := []rune(strings.ToLower(args[0]))
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		return "Invalid item letter. Use a-z."
	}
	return fmt.Sprintf("No item at slot %s.", args[0])
}

func equippedItem(player *actor.Player, slot string) *gameitem.Item {
	switch slot {
	case equipmentSlotWeapon:
		return player.Equipment.Weapon
	case equipmentSlotArmor:
		return player.Equipment.Armor
	case equipmentSlotRingLeft:
		return player.Equipment.RingLeft
	case equipmentSlotRingRight:
		return player.Equipment.RingRight
	default:
		return nil
	}
}

func resolveDirection(cmd Command, args []string) (Direction, bool) {
	if cmd.Direction != (Direction{}) {
		return cmd.Direction, true
	}
	if len(args) == 0 {
		return Direction{}, false
	}
	return ParseDirection(args[0])
}

func adjacent(x1, y1, x2, y2 int) bool {
	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1 && (dx != 0 || dy != 0)
}

func unavailable(ctx *Context) bool {
	return ctx == nil || ctx.Player == nil || ctx.Level == nil || ctx.Player.Position == nil
}

func advanceMonsters(ctx *Context) {
	if ctx == nil || ctx.Level == nil || ctx.Player == nil {
		return
	}
	hasteActive := ctx.Player.HasteTurns > 0
	skipMonsterTurn := hasteActive && ctx.Player.HasteSkipMonsterTurn
	if hasteActive {
		ctx.Player.HasteSkipMonsterTurn = !ctx.Player.HasteSkipMonsterTurn
	} else {
		ctx.Player.HasteSkipMonsterTurn = false
	}
	ctx.Player.AdvanceStatuses()
	ctx.Player.UpdateHunger()
	if ctx.Player.IsAlive() && !skipMonsterTurn {
		ctx.Level.UpdateMonsters(ctx.Player)
	}
	releaseDeadFlytrapHold(ctx)
	if ctx.Player.IsAlive() {
		ctx.Player.Recover(ctx.Player.QuietTurns)
		applyRingTurnEffects(ctx)
	}
}
func releaseDeadFlytrapHold(ctx *Context) {
	if !ctx.Player.Held {
		return
	}
	for _, monster := range ctx.Level.Monsters {
		if monster == nil || !monster.IsAlive() || monster.Type.Code != 'F' {
			continue
		}
		dx := monster.Position.X - ctx.Player.Position.X
		dy := monster.Position.Y - ctx.Player.Position.Y
		if dx >= -1 && dx <= 1 && dy >= -1 && dy <= 1 {
			return
		}
	}
	ctx.Player.Held = false
}

func applyRingTurnEffects(ctx *Context) {
	if ctx.Player.Equipment == nil {
		return
	}
	for _, ring := range []*gameitem.Item{ctx.Player.Equipment.RingLeft, ctx.Player.Equipment.RingRight} {
		switch sourceRingName(ring) {
		case "searching":
			foundTraps := ctx.Level.SearchTraps(ctx.Player.Position.X, ctx.Player.Position.Y, ctx.Player.BlindTurns > 0, ctx.Player.HallucinationTurns > 0)
			foundSecrets := ctx.Level.SearchSecrets(ctx.Player.Position.X, ctx.Player.Position.Y, ctx.Player.BlindTurns > 0, ctx.Player.HallucinationTurns > 0)
			if len(foundTraps) > 0 || foundSecrets > 0 {
				ctx.messages = append(ctx.messages, "Your ring of searching reveals something.")
			}
		case "teleportation":
			if ctx.Player.RandomSource().Intn(50) == 0 && teleportPlayer(ctx) {
				ctx.messages = append(ctx.messages, "Your ring of teleportation makes you vanish.")
			}
		case "aggravate monster":
			for _, monster := range ctx.Level.Monsters {
				if monster != nil && monster.IsAlive() {
					monster.IsRunning = true
				}
			}
		}
	}
}

func result(ctx *Context, message string) Result {
	result := Result{Message: message}
	if ctx != nil {
		result.Player = ctx.Player
		result.Level = ctx.Level
		result.Dungeon = ctx.Dungeon
	}
	return result
}

func turnResult(ctx *Context, message string) Result {
	if ctx != nil && len(ctx.messages) > 0 {
		message += "\n" + strings.Join(ctx.messages, "\n")
		ctx.messages = ctx.messages[:0]
	}
	result := result(ctx, message)
	result.TurnConsumed = true
	if ctx != nil && ctx.Player != nil && !ctx.Player.IsAlive() {
		result.PlayerDied = true
		result.Message += "\nYou died."
		if ctx.Save != nil {
			ctx.Save.OnPlayerDeath("Player died during a turn.")
		}
	}
	return result
}

func victoryResult(ctx *Context, message string) Result {
	result := turnResult(ctx, message)
	result.Victory = true
	return result
}

func failure(ctx *Context, message string) Result {
	result := result(ctx, message)
	result.Error = true
	return result
}

func executeZap(ctx *Context, cmd Command, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	_, wand, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}
	if wand.Type != gameitem.ItemWand {
		return result(ctx, "You can only zap a wand.")
	}
	direction, ok := resolveDirection(cmd, nil)
	if !ok {
		return result(ctx, "Zap in which direction?")
	}
	outcome := magic.UseWand(wand, ctx.Player, ctx.Level, direction.X, direction.Y)
	if !outcome.Success {
		return result(ctx, outcome.Message)
	}
	message := outcome.Message
	if outcome.Experience > 0 {
		ctx.Player.GainExp(outcome.Experience)
		message = fmt.Sprintf("%s You gain %d experience.", message, outcome.Experience)
	}
	advanceMonsters(ctx)
	return turnResult(ctx, message)
}

func executeThrow(ctx *Context, cmd Command, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	index, thrown, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}
	direction, ok := resolveDirection(cmd, nil)
	if !ok {
		return result(ctx, "Throw in which direction?")
	}

	x, y := ctx.Player.Position.X, ctx.Player.Position.Y
	var target *actor.Monster
	for range 5 {
		nextX, nextY := x+direction.X, y+direction.Y
		tile := ctx.Level.GetTile(nextX, nextY)
		if tile == nil || !tile.Walkable() {
			break
		}
		x, y = nextX, nextY
		if target = ctx.Level.GetMonsterAt(x, y); target != nil {
			break
		}
	}

	if ctx.Player.Equipment.Weapon == thrown {
		return result(ctx, "You can't throw the weapon you're wielding.")
	}
	message := fmt.Sprintf("You throw %s.", ctx.Player.IdentifyMgr.GetDisplayName(thrown))
	hit := false
	if target != nil {
		hit, message = resolveThrownAttack(ctx, target, thrown)
	}
	consumeThrownItem(ctx, index, thrown, hit, x, y)
	advanceMonsters(ctx)
	return turnResult(ctx, message)
}

func resolveThrownAttack(ctx *Context, target *actor.Monster, thrown *gameitem.Item) (hit bool, message string) {
	attack := ctx.Player.ProjectileAttackRoll(target.Defense, target.IsRunning, thrown, ctx.Player.Equipment.Weapon)
	target.IsRunning = true
	target.IsHeld = false
	if !attack.Hit {
		return false, fmt.Sprintf("Your %s misses %s.", thrown.Name, target.Type.Name)
	}
	target.TakeDamage(attack.Damage)
	if attack.Confuses {
		target.IsConfused = true
	}
	if !target.IsAlive() {
		ctx.Player.GainExp(target.Type.Experience)
		return true, fmt.Sprintf("Your %s hits and kills %s.", thrown.Name, target.Type.Name)
	}
	return true, fmt.Sprintf("Your %s hits %s for %d damage.", thrown.Name, target.Type.Name, attack.Damage)
}

func consumeThrownItem(ctx *Context, index int, thrown *gameitem.Item, hit bool, x, y int) {
	if thrown.Quantity > 1 {
		thrown.Quantity--
		if !hit {
			dropped := *thrown
			dropped.Quantity = 1
			dropped.Entity = entity.NewEntity(x, y)
			ctx.Level.AddItem(&dropped, x, y)
		}
		return
	}
	ctx.Player.Inventory.RemoveItem(index)
	if !hit {
		ctx.Level.AddItem(thrown, x, y)
	}
}

func executeCharacter(ctx *Context) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	return result(ctx, fmt.Sprintf(
		"Level %d, experience %d, strength %d/%d, HP %d/%d, gold %d, food %d.",
		ctx.Player.Level, ctx.Player.Exp, ctx.Player.Strength, ctx.Player.MaxStrength,
		ctx.Player.HP, ctx.Player.MaxHP, ctx.Player.Gold, ctx.Player.FoodLeft))
}

func executeDiscover(ctx *Context) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	discovered := ctx.Player.IdentifyMgr.DiscoveredItems()
	if len(discovered) == 0 {
		return result(ctx, "You have not discovered any items.")
	}
	lines := make([]string, len(discovered))
	for i, discoveredItem := range discovered {
		category := "item"
		switch discoveredItem.Type {
		case gameitem.ItemPotion:
			category = "potion"
		case gameitem.ItemScroll:
			category = "scroll"
		case gameitem.ItemRing:
			category = "ring"
		case gameitem.ItemWand:
			category = "wand"
		}
		lines[i] = fmt.Sprintf("%s: %s", category, discoveredItem.Name)
	}
	return result(ctx, strings.Join(lines, "\n"))
}
