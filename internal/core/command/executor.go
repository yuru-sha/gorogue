package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/magic"
	"github.com/yuru-sha/gorogue/internal/game/save"
)

// Context contains the mutable game state needed to execute a gameplay command.
type Context struct {
	Player  *actor.Player
	Level   *dungeon.Level
	Dungeon *dungeon.DungeonManager
	Save    *save.SaveGameIntegration
}

// Result describes the observable outcome of a gameplay command.
type Result struct {
	Message      string
	Error        bool
	TurnConsumed bool
	Victory      bool
	Player       *actor.Player
	Level        *dungeon.Level
	Dungeon      *dungeon.DungeonManager
}

// Execute runs a gameplay command for either the GUI or CLI entry point.
func Execute(ctx *Context, cmd Command, args ...string) Result {
	if ctx == nil {
		return Result{Message: "Game state is unavailable."}
	}

	switch cmd.Type {
	case CmdMoveWest, CmdMoveEast, CmdMoveNorth, CmdMoveSouth,
		CmdMoveNorthWest, CmdMoveNorthEast, CmdMoveSouthWest, CmdMoveSouthEast:
		return executeMove(ctx, cmd.Direction)
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
	case CmdWield, CmdEquip:
		return executeEquip(ctx, args)
	case CmdTakeOff, CmdUnequip:
		return executeUnequip(ctx, args)
	case CmdWait:
		return executeWait(ctx, args)
	case CmdSearch:
		return executeSearch(ctx)
	case CmdOpen, CmdClose:
		return executeDoor(ctx, cmd, args)
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
	if direction.X == 0 && direction.Y == 0 {
		return result(ctx, "Invalid movement direction.")
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

	ctx.Player.Position.X = newX
	ctx.Player.Position.Y = newY
	message := fmt.Sprintf("Moved %s to (%d, %d).", directionName(direction), newX, newY)
	if pickupMessage, _ := pickUpAt(ctx, newX, newY); pickupMessage != "" {
		message += "\n" + pickupMessage
	}
	advanceMonsters(ctx)
	return turnResult(ctx, message)
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
	damage := ctx.Player.CalculateDamage(monster.Defense)
	monster.TakeDamage(damage)
	if monster.IsAlive() {
		advanceMonsters(ctx)
		return turnResult(ctx, fmt.Sprintf("Attacked %s for %d damage! (%d HP remaining)",
			monster.Type.Name, damage, monster.HP))
	}

	exp := monster.MaxHP + monster.Attack
	gold := monster.MaxHP / 2
	ctx.Player.GainExp(exp)
	ctx.Player.AddGold(gold)
	advanceMonsters(ctx)
	return turnResult(ctx, fmt.Sprintf("Killed %s! Gained %d experience and %d gold.",
		monster.Type.Name, exp, gold))
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
	return turnResult(ctx, fmt.Sprintf("Dropped %s.", ctx.Player.IdentifyMgr.GetDisplayName(item)))
}

func executeEquip(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	index, item, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}
	if item.Type != gameitem.ItemWeapon && item.Type != gameitem.ItemArmor && item.Type != gameitem.ItemRing {
		return result(ctx, "That item cannot be equipped.")
	}
	if !ctx.Player.Equipment.EquipItem(item) {
		return result(ctx, "Cannot equip that item (slot occupied?).")
	}

	ctx.Player.Inventory.RemoveItem(index)
	return turnResult(ctx, fmt.Sprintf("Equipped %s.", ctx.Player.IdentifyMgr.GetDisplayName(item)))
}

func executeUnequip(ctx *Context, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	if len(args) == 0 {
		return result(ctx, "Usage: unequip <slot>")
	}

	slot := strings.ToLower(args[0])
	slotName := map[string]string{
		"weapon":     "weapon",
		"w":          "weapon",
		"armor":      "armor",
		"a":          "armor",
		"ring-left":  "ring_left",
		"left":       "ring_left",
		"l":          "ring_left",
		"ring-right": "ring_right",
		"right":      "ring_right",
		"r":          "ring_right",
	}[slot]
	if slotName == "" {
		return result(ctx, "Unknown slot. Use: weapon, armor, ring-left, ring-right")
	}
	if equipped := equippedItem(ctx.Player, slotName); equipped != nil && equipped.IsCursed {
		return result(ctx, fmt.Sprintf("You can't remove the cursed %s.", slot))
	}

	item := ctx.Player.Equipment.UnequipItem(slotName)
	if item == nil {
		return result(ctx, fmt.Sprintf("No item equipped in %s slot.", slot))
	}
	if !ctx.Player.Inventory.AddItem(item) {
		ctx.Player.Equipment.EquipItem(item)
		return result(ctx, "Inventory is full! Cannot unequip.")
	}
	return turnResult(ctx, fmt.Sprintf("Unequipped %s.", ctx.Player.IdentifyMgr.GetDisplayName(item)))
}

func executeItem(ctx *Context, commandType Type, args []string) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	index, item, ok := inventoryItem(ctx, args)
	if !ok {
		return result(ctx, inventoryItemError(args))
	}

	var message string
	var identified bool
	turnConsumed := true
	switch commandType {
	case CmdUse:
		switch item.Type {
		case gameitem.ItemPotion:
			effect := magic.UsePotion(item.Name, ctx.Player)
			message, identified = effect.Message, effect.Identified
		case gameitem.ItemScroll:
			effect := magic.UseScroll(item.Name, ctx.Player, ctx.Level)
			message, identified = effect.Message, effect.Identified
		case gameitem.ItemFood:
			ctx.Player.EatFood(item.Value)
			message = fmt.Sprintf("You ate %s.", ctx.Player.IdentifyMgr.GetDisplayName(item))
			identified = true
		default:
			return result(ctx, "That item cannot be used.")
		}
	case CmdQuaff:
		if item.Type != gameitem.ItemPotion {
			return result(ctx, "You can't drink that!")
		}
		effect := magic.UsePotion(item.Name, ctx.Player)
		message, identified = effect.Message, effect.Identified
	case CmdRead:
		if item.Type != gameitem.ItemScroll {
			return result(ctx, "You can't read that!")
		}
		effect := magic.UseScroll(item.Name, ctx.Player, ctx.Level)
		message, identified = effect.Message, effect.Identified
	case CmdEat:
		if item.Type != gameitem.ItemFood {
			return result(ctx, "You can't eat that!")
		}
		ctx.Player.EatFood(item.Value)
		message = fmt.Sprintf("You ate %s.", ctx.Player.IdentifyMgr.GetDisplayName(item))
		identified = true
	}

	if identified {
		ctx.Player.IdentifyMgr.IdentifyByUse(item)
	}
	ctx.Player.Inventory.RemoveItem(index)
	if turnConsumed {
		advanceMonsters(ctx)
	}
	commandResult := result(ctx, message)
	commandResult.TurnConsumed = turnConsumed
	return commandResult
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

	lines := []string{fmt.Sprintf("Position (%d, %d):", x, y)}
	if tile := ctx.Level.GetTile(x, y); tile != nil {
		lines = append(lines, fmt.Sprintf("Terrain: %s", tile.Type.String()))
	}
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
	return result(ctx, strings.Join(lines, "\n"))
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
	healAmount := turns / 5
	if healAmount > 0 {
		ctx.Player.Heal(healAmount)
	}
	for i := 0; i < turns; i++ {
		advanceMonsters(ctx)
	}
	if len(args) == 0 {
		return turnResult(ctx, "You rest.")
	}
	return turnResult(ctx, fmt.Sprintf("Rested for %d turns. (Healed %d HP)", turns, healAmount))
}

func executeSearch(ctx *Context) Result {
	if unavailable(ctx) {
		return result(ctx, "Game state is unavailable.")
	}
	advanceMonsters(ctx)
	return turnResult(ctx, "You search the area.")
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
		return turnResult(ctx, fmt.Sprintf("Climbed to floor %d.", ctx.Dungeon.GetCurrentFloor()))
	}

	if !ctx.Dungeon.CanGoDownstairs() || !ctx.Dungeon.GoDownstairs() {
		return result(ctx, "There are no usable down stairs here.")
	}
	ctx.Level = ctx.Dungeon.GetCurrentLevel()
	if ctx.Dungeon.IsOnFinalFloor() {
		ctx.Dungeon.PlaceAmuletOfYendor()
	}
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
	case "weapon":
		return player.Equipment.Weapon
	case "armor":
		return player.Equipment.Armor
	case "ring_left":
		return player.Equipment.RingLeft
	case "ring_right":
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
	if ctx.Level != nil && ctx.Player != nil {
		ctx.Level.UpdateMonsters(ctx.Player)
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
	result := result(ctx, message)
	result.TurnConsumed = true
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
