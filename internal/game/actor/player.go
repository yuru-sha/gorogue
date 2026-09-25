package actor

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/yuru-sha/gorogue/internal/game/identification"
	"github.com/yuru-sha/gorogue/internal/game/inventory"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// Player represents the player character
type Player struct {
	*Actor
	Level                 int
	Hunger                int
	Exp                   int
	Gold                  int
	Strength              int
	MaxStrength           int
	FoodLeft              int
	HungerState           int
	QuietTurns            int
	Running               bool
	NoCommandTurns        int
	NoMoveTurns           int
	BlindTurns            int
	ConfusedTurns         int
	HallucinationTurns    int
	HasteTurns            int
	HasteSkipMonsterTurn  bool
	SeeInvisibleTurns     int
	MonsterDetectionTurns int
	LevitationTurns       int
	ParalyzedTurns        int
	CanConfuse            bool
	CanConfuseTurns       int
	Held                  bool
	Inventory             *inventory.Inventory
	Equipment             *inventory.Equipment
	IdentifyMgr           *identification.IdentificationManager
	rng                   *rand.Rand
}

const (
	HungerSatisfied = iota
	HungerHungry
	HungerWeak
	HungerFainting
	sourceFoodStart    = 1300
	sourceFoodMaximum  = 2000
	sourceMoreTime     = 150
	sourceStarveTime   = 850
	ringProtectionName = "protection"
	ringStealthName    = "stealth"
)

// Saving throw classes from Rogue's rogue.h.
const (
	SavingThrowPoison    = 0
	SavingThrowParalysis = 0
	SavingThrowDeath     = 0
	SavingThrowBreath    = 2
	SavingThrowMagic     = 3
)

var experienceThresholds = []int{
	0, 10, 20, 40, 80, 160, 320, 640, 1300, 2600,
	5200, 13000, 26000, 50000, 100000, 200000, 400000,
	800000, 2000000, 4000000, 8000000,
}

func newRandomSource() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewPlayer creates a new player at the given position
func NewPlayer(x, y int) *Player {
	return NewPlayerWithRand(x, y, newRandomSource())
}

// NewPlayerWithSeed creates a player with deterministic item appearances and effects.
func NewPlayerWithSeed(x, y int, seed int64) *Player {
	return NewPlayerWithRand(x, y, rand.New(rand.NewSource(seed)))
}

// NewPlayerWithRand creates a player using the supplied random source.
func NewPlayerWithRand(x, y int, rng *rand.Rand) *Player {
	if rng == nil {
		rng = newRandomSource()
	}
	player := &Player{
		Actor:       NewActor(x, y, 12, 5, 10),
		Level:       1,
		Hunger:      100,
		Exp:         0,
		Gold:        0,
		Strength:    16,
		MaxStrength: 16,
		FoodLeft:    sourceFoodStart,
		HungerState: HungerSatisfied,
		Inventory:   inventory.NewInventory(),
		Equipment:   inventory.NewEquipment(),
		IdentifyMgr: identification.NewIdentificationManagerWithRand(rng),
		rng:         rng,
	}
	logger.Debug("Created new player",
		"position_x", x,
		"position_y", y,
		"level", player.Level,
		"hp", player.HP,
		"attack", player.Attack,
		"defense", player.Defense,
	)
	return player
}

// SetRandomSource assigns the game's random source to the player and combat.
func (p *Player) SetRandomSource(rng *rand.Rand) {
	if rng != nil {
		p.rng = rng
	}
}

// RandomSource returns the game's random source for player effects.
func (p *Player) RandomSource() *rand.Rand {
	return p.random()
}

func (p *Player) random() *rand.Rand {
	if p.rng == nil {
		p.rng = newRandomSource()
	}
	return p.rng
}

// TakeDamage resets the source recovery counter when the player loses HP.
func (p *Player) TakeDamage(damage int) {
	oldHP := p.HP
	p.Actor.TakeDamage(damage)
	if p.HP < oldHP {
		p.QuietTurns = 0
	}
}

// AddGold adds gold to the player's inventory
func (p *Player) AddGold(amount int) {
	oldGold := p.Gold
	p.Gold += amount
	logger.Debug("Player collected gold",
		"amount", amount,
		"gold_before", oldGold,
		"gold_after", p.Gold,
	)
}

// AddExp adds experience and applies Rogue's level thresholds and HP roll.
func (p *Player) AddExp(amount int) {
	oldExp := p.Exp
	oldLevel := p.Level
	p.Exp += amount
	newLevel := levelForExperience(p.Exp)
	if newLevel > oldLevel {
		for level := oldLevel; level < newLevel; level++ {
			gain := p.random().Intn(10) + 1
			p.MaxHP += gain
			p.HP += gain
		}
	}
	p.Level = newLevel
	logger.Debug("Player gained experience",
		"amount", amount,
		"exp_before", oldExp,
		"exp_after", p.Exp,
		"level_before", oldLevel,
		"level_after", p.Level,
	)
}

func levelForExperience(exp int) int {
	level := 1
	for level < len(experienceThresholds) && exp >= experienceThresholds[level] {
		level++
	}
	return level
}

// GainExp is an alias for AddExp
func (p *Player) GainExp(amount int) {
	p.AddExp(amount)
}

// AttackResult is the source combat outcome for one player swing.
type AttackResult struct {
	Hit      bool
	Damage   int
	Confuses bool
}

// CalculateDamage retains the existing damage API; misses return zero damage.
func (p *Player) CalculateDamage(targetDefense int) int {
	return p.AttackRoll(targetDefense, true).Damage
}

// AttackRoll resolves one Rogue swing against the target's armor class.
func (p *Player) AttackRoll(targetArmor int, targetRunning bool) AttackResult {
	p.QuietTurns = 0
	dice := sourceDice{count: 1, sides: 4}
	hitBonus, damageBonus := 0, 0
	if p.Equipment != nil && p.Equipment.Weapon != nil {
		weapon := p.Equipment.Weapon
		dice, hitBonus = weaponDamage(weapon), weapon.Enchantment
		damageBonus = weapon.Enchantment
	}
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		switch sourceItemName(ring) {
		case "dexterity":
			hitBonus += ring.Enchantment
		case "increase damage":
			damageBonus += ring.Enchantment
		}
	}
	if !p.RollHit(p.Level, targetArmor, hitBonus, p.Strength, targetRunning) {
		return AttackResult{}
	}

	damage := damageBonus + sourceStrengthDamage(p.Strength)
	for range dice.count {
		damage += p.random().Intn(dice.sides) + 1
	}
	if damage < 0 {
		damage = 0
	}
	result := AttackResult{Hit: true, Damage: damage, Confuses: p.CanConfuse}
	p.CanConfuse = false
	p.CanConfuseTurns = 0
	return result
}

// ProjectileAttackRoll resolves a thrown weapon using Rogue's missile dice and launcher bonuses.
func (p *Player) ProjectileAttackRoll(targetArmor int, targetRunning bool, projectile, launcher *item.Item) AttackResult {
	p.QuietTurns = 0
	definition, ok := item.WeaponForItem(projectile)
	if !ok {
		return AttackResult{}
	}
	diceSpec := definition.ThrownDamage
	hitBonus, damageBonus := projectile.Enchantment, projectile.Enchantment
	if definition.LauncherID != 0 {
		if launcher == nil || launcher.ItemID != definition.LauncherID {
			diceSpec = definition.MeleeDamage
		} else {
			hitBonus += launcher.Enchantment
			damageBonus += launcher.Enchantment
		}
	}
	dice := parseSourceDice(diceSpec)
	if !p.RollHit(p.Level, targetArmor, hitBonus, p.Strength, targetRunning) {
		return AttackResult{}
	}
	damage := damageBonus + sourceStrengthDamage(p.Strength)
	for range dice.count {
		damage += p.random().Intn(dice.sides) + 1
	}
	if damage < 0 {
		damage = 0
	}
	attack := AttackResult{Hit: true, Damage: damage, Confuses: p.CanConfuse}
	p.CanConfuse = false
	p.CanConfuseTurns = 0
	return attack
}

type sourceDice struct {
	count int
	sides int
}

func parseSourceDice(spec string) sourceDice {
	countText, sidesText, ok := strings.Cut(spec, "x")
	if !ok {
		return sourceDice{count: 1, sides: 4}
	}
	count, countErr := strconv.Atoi(countText)
	sides, sidesErr := strconv.Atoi(sidesText)
	if countErr != nil || sidesErr != nil || count <= 0 || sides <= 0 {
		return sourceDice{count: 1, sides: 4}
	}
	return sourceDice{count: count, sides: sides}
}

func weaponDamage(weapon *item.Item) sourceDice {
	switch sourceItemName(weapon) {
	case "mace":
		return sourceDice{count: 2, sides: 4}
	case "long sword":
		return sourceDice{count: 3, sides: 4}
	case "short bow", "bow":
		return sourceDice{count: 1, sides: 1}
	case "arrow":
		return sourceDice{count: 1, sides: 1}
	case "dagger":
		return sourceDice{count: 1, sides: 6}
	case "two handed sword", "two-handed sword":
		return sourceDice{count: 4, sides: 4}
	case "dart":
		return sourceDice{count: 1, sides: 1}
	case "shuriken":
		return sourceDice{count: 1, sides: 2}
	case "spear":
		return sourceDice{count: 2, sides: 3}
	default:
		return sourceDice{count: 1, sides: 4}
	}
}

// RollHit applies fight.c's d20 swing check using the player's managed RNG.
func (p *Player) RollHit(attackerLevel, targetArmor, hitBonus, strength int, targetRunning bool) bool {
	p.QuietTurns = 0
	if !targetRunning {
		hitBonus += 4
	}
	return p.random().Intn(20)+hitBonus+sourceStrengthHit(strength) >= 20-attackerLevel-targetArmor
}

// GetTotalDefense returns Rogue's armor class (lower values are harder to hit).
func (p *Player) GetTotalDefense() int {
	armorClass := p.Defense
	if p.Equipment == nil {
		return armorClass
	}
	if p.Equipment.Armor != nil {
		armorClass = 10 - (p.Equipment.Armor.Defense + p.Equipment.Armor.Enchantment)
	}
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		if sourceItemName(ring) == ringProtectionName {
			armorClass -= ring.Enchantment
		}
	}
	return armorClass
}

func (p *Player) ringLeft() *item.Item {
	if p.Equipment == nil {
		return nil
	}
	return p.Equipment.RingLeft
}

func (p *Player) ringRight() *item.Item {
	if p.Equipment == nil {
		return nil
	}
	return p.Equipment.RingRight
}

func sourceItemName(itm *item.Item) string {
	if itm == nil {
		return ""
	}
	name := itm.RealName
	if name == "" {
		name = itm.Name
	}
	return strings.ToLower(strings.TrimSpace(name))
}

var strengthHitAdjustment = [...]int{
	-7, -6, -5, -4, -3, -2, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3,
}

var strengthDamageAdjustment = [...]int{
	-7, -6, -5, -4, -3, -2, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 2, 3, 3, 4, 5, 5, 5, 5, 5, 5, 5, 5, 5, 6,
}

func strengthIndex(strength int) int {
	return min(max(strength, 0), len(strengthHitAdjustment)-1)
}

func sourceStrengthHit(strength int) int {
	return strengthHitAdjustment[strengthIndex(strength)]
}

func sourceStrengthDamage(strength int) int {
	return strengthDamageAdjustment[strengthIndex(strength)]
}

// UpdateHunger advances Rogue's stomach daemon by one consumed turn.
func (p *Player) UpdateHunger() {
	oldFood := p.FoodLeft
	if p.FoodLeft <= 0 {
		if p.FoodLeft < -sourceStarveTime {
			p.HP = 0
			return
		}
		p.FoodLeft--
		if p.NoCommandTurns == 0 && p.random().Intn(5) == 0 {
			p.NoCommandTurns += p.random().Intn(8) + 4
			p.HungerState = HungerFainting
		}
	} else {
		p.FoodLeft -= p.foodDrain()
		if p.FoodLeft < sourceMoreTime && oldFood >= sourceMoreTime {
			p.HungerState = HungerWeak
		} else if p.FoodLeft < 2*sourceMoreTime && oldFood >= 2*sourceMoreTime {
			p.HungerState = HungerHungry
		}
	}
	p.syncHunger()
}

func (p *Player) foodDrain() int {
	drain := 1
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		switch sourceItemName(ring) {
		case ringProtectionName, "add strength", "sustain strength", ringStealthName, "maintain armor":
			drain++
		case "regeneration":
			drain += 2
		case "searching", "dexterity", "increase damage":
			if p.random().Intn(3) == 0 {
				drain++
			}
		case "see invisible":
			if p.random().Intn(5) == 0 {
				drain++
			}
		case "slow digestion":
			if p.random().Intn(2) != 0 {
				drain--
			}
		}
	}
	if p.Inventory != nil && p.Inventory.HasItemType(item.ItemAmulet) {
		drain--
	}
	return drain
}

func (p *Player) syncHunger() {
	if p.FoodLeft <= 0 {
		p.Hunger = 0
		return
	}
	p.Hunger = min(100, p.FoodLeft*100/sourceFoodStart)
}

// EatFood preserves its caller API while applying Rogue's stomach refill.
func (p *Player) EatFood(_ int) {
	if p.FoodLeft < 0 {
		p.FoodLeft = 0
	}
	p.FoodLeft += sourceFoodStart - 200 + p.random().Intn(400)
	if p.FoodLeft > sourceFoodMaximum {
		p.FoodLeft = sourceFoodMaximum
	}
	p.HungerState = HungerSatisfied
	p.syncHunger()
}

// Recover applies one Rogue doctor daemon tick after the supplied quiet turns.
// It returns true when HP changes so the caller can reset its quiet counter.
func (p *Player) Recover(quietTurns int) bool {
	if !p.IsAlive() {
		p.QuietTurns = 0
		return false
	}
	quiet := quietTurns + 1
	p.QuietTurns = quiet
	oldHP := p.HP
	if p.Level < 8 {
		if quiet+(p.Level<<1) > 20 {
			p.HP++
		}
	} else if quiet >= 3 {
		p.HP += p.random().Intn(p.Level-7) + 1
	}
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		if sourceItemName(ring) == "regeneration" {
			p.HP++
		}
	}
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	if p.HP != oldHP {
		p.QuietTurns = 0
		return true
	}
	return false
}

// AdvanceStatuses advances player fuses once after a consumed turn.
func (p *Player) AdvanceStatuses() {
	if p.NoCommandTurns > 0 {
		p.NoCommandTurns--
	}
	if p.NoMoveTurns > 0 {
		p.NoMoveTurns--
	}
	if p.CanConfuseTurns > 0 {
		p.CanConfuseTurns--
		if p.CanConfuseTurns == 0 {
			p.CanConfuse = false
		}
	}
	for _, turns := range []*int{
		&p.BlindTurns,
		&p.ConfusedTurns,
		&p.HallucinationTurns,
		&p.HasteTurns,
		&p.SeeInvisibleTurns,
		&p.MonsterDetectionTurns,
		&p.LevitationTurns,
		&p.ParalyzedTurns,
	} {
		if *turns > 0 {
			*turns--
		}
	}
	if p.HasteTurns == 0 {
		p.HasteSkipMonsterTurn = false
	}
}

// AdjustStrength applies a Rogue strength change, bounded to 3..31.
func (p *Player) AdjustStrength(amount int) {
	p.Strength = min(max(p.Strength+amount, 3), 31)
	if p.Strength > p.MaxStrength {
		p.MaxStrength = p.Strength
	}
}

func (p *Player) ReduceStrength(amount int) {
	if p.wearingRing("sustain strength") {
		return
	}
	p.AdjustStrength(-amount)
}

func (p *Player) wearingRing(name string) bool {
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		if sourceItemName(ring) == name {
			return true
		}
	}
	return false
}
func (p *Player) RestoreStrength() {
	p.Strength = p.MaxStrength
}

// DrainLevel applies the Wraith's source experience-level drain.
func (p *Player) DrainLevel() {
	if p.Exp == 0 {
		p.HP = 0
		return
	}
	p.Level--
	if p.Level <= 0 {
		p.Exp = 0
		p.Level = 1
		return
	}
	p.Exp = experienceThresholds[p.Level-1] + 1
}

// DrainMaxHP applies the Vampire's current and maximum HP drain.
func (p *Player) DrainMaxHP(amount int) {
	p.MaxHP -= amount
	p.HP -= amount
	if p.HP <= 0 {
		p.HP = 1
	}
	if p.MaxHP <= 0 {
		p.HP = 0
	}
}

// Hold prevents movement until the source holding effect is cleared.
func (p *Player) Hold() {
	p.Held = true
}

// Freeze adds source no-command turns and applies the Rogue boredom limit.
func (p *Player) Freeze(turns int) {
	if turns <= 0 {
		return
	}
	p.NoCommandTurns += turns
	p.ParalyzedTurns += turns
	if p.NoCommandTurns > 50 {
		p.HP = 0
	}
}

// RustArmor reduces equipped armor's protection unless source protection applies.
func (p *Player) RustArmor() {
	if p.Equipment == nil || p.Equipment.Armor == nil || p.Equipment.Armor.IsProtected {
		return
	}
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		if sourceItemName(ring) == "maintain armor" {
			return
		}
	}
	p.Equipment.Armor.Defense--
}

// SaveAgainst resolves a Rogue saving throw with protection-ring modifiers.
func (p *Player) SaveAgainst(saveType int) bool {
	for _, ring := range []*item.Item{p.ringLeft(), p.ringRight()} {
		if sourceItemName(ring) == ringProtectionName {
			saveType -= ring.Enchantment
		}
	}
	return p.random().Intn(20)+1 >= 14+saveType-p.Level/2
}

// GetExpToNextLevel returns experience needed to reach next level.
func (p *Player) GetExpToNextLevel() int {
	if p.Level < 1 || p.Level >= len(experienceThresholds) {
		return 0
	}
	nextLevelExp := experienceThresholds[p.Level]
	return nextLevelExp - p.Exp
}
