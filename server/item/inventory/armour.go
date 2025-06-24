package inventory

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"math/rand/v2"
)

// Armour represents an inventory for armour. It has 4 slots, one for a helmet, chestplate, leggings and
// boots respectively. NewArmour() must be used to create a valid armour inventory.
// Armour inventories, like normal Inventories, are safe for concurrent usage.
type Armour struct {
	inv *Inventory
}

// NewArmour returns an armour inventory that is ready to be used. The zero value of an inventory.Armour is
// not valid for usage.
// The function passed is called when a slot is changed. It may be nil to not call anything.
func NewArmour(f func(slot int, before, after item.Stack)) *Armour {
	inv := New(4, f)
	inv.validator = canAddArmour
	return &Armour{inv: inv}
}

// canAddArmour checks if the item passed can be worn as armour in the slot passed.
func canAddArmour(s item.Stack, slot int) bool {
	if s.Empty() {
		return true
	}
	switch slot {
	case 0:
		if h, ok := s.Item().(item.HelmetType); ok {
			return h.Helmet()
		}
	case 1:
		if c, ok := s.Item().(item.ChestplateType); ok {
			return c.Chestplate()
		}
	case 2:
		if l, ok := s.Item().(item.LeggingsType); ok {
			return l.Leggings()
		}
	case 3:
		if b, ok := s.Item().(item.BootsType); ok {
			return b.Boots()
		}
	}
	return false
}

// Set sets all individual pieces of armour in one go. It is equivalent to calling SetHelmet, SetChestplate, SetLeggings
// and SetBoots sequentially.
func (a *Armour) Set(helmet, chestplate, leggings, boots item.Stack) {
	a.SetHelmet(helmet)
	a.SetChestplate(chestplate)
	a.SetLeggings(leggings)
	a.SetBoots(boots)
}

// SetHelmet sets the item stack passed as the helmet in the inventory.
func (a *Armour) SetHelmet(helmet item.Stack) {
	_ = a.inv.SetItem(0, helmet)
}

// Helmet returns the item stack set as helmet in the inventory.
func (a *Armour) Helmet() item.Stack {
	i, _ := a.inv.Item(0)
	return i
}

// SetChestplate sets the item stack passed as the chestplate in the inventory.
func (a *Armour) SetChestplate(chestplate item.Stack) {
	_ = a.inv.SetItem(1, chestplate)
}

// Chestplate returns the item stack set as chestplate in the inventory.
func (a *Armour) Chestplate() item.Stack {
	i, _ := a.inv.Item(1)
	return i
}

// SetLeggings sets the item stack passed as the leggings in the inventory.
func (a *Armour) SetLeggings(leggings item.Stack) {
	_ = a.inv.SetItem(2, leggings)
}

// Leggings returns the item stack set as leggings in the inventory.
func (a *Armour) Leggings() item.Stack {
	i, _ := a.inv.Item(2)
	return i
}

// SetBoots sets the item stack passed as the boots in the inventory.
func (a *Armour) SetBoots(boots item.Stack) {
	_ = a.inv.SetItem(3, boots)
}

// Boots returns the item stack set as boots in the inventory.
func (a *Armour) Boots() item.Stack {
	i, _ := a.inv.Item(3)
	return i
}

// DamageReduction returns the amount of damage that is reduced by the Armour for
// an amount of damage and damage source. The value returned takes into account
// the armour itself and its enchantments.
func (a *Armour) DamageReduction(dmg float64, src world.DamageSource) float64 {
	var (
		original                 = dmg
		defencePoints, toughness float64
		enchantments             []item.Enchantment
	)

	for _, it := range a.Items() {
		enchantments = append(enchantments, it.Enchantments()...)
		if armour, ok := it.Item().(item.Armour); ok {
			defencePoints += armour.DefencePoints()
			toughness += armour.Toughness()
		}
	}

	dmg -= dmg * enchantment.ProtectionFactor(src, enchantments)
	if src.ReducedByArmour() {
		// Armour in Bedrock edition reduces the damage taken by 4% for each effective armour point. Effective
		// armour point decreases as damage increases, with 1 point lost for every 2 HP of damage. The defense
		// reduction is decreased by the toughness armor value. Effective armour points will at minimum be 20% of
		// armour points.
		dmg -= dmg * 0.04 * math.Max(defencePoints*0.2, defencePoints-dmg/(2+toughness/4))
	}
	return original - dmg
}

// HighestEnchantmentLevel looks up the highest level of an item.EnchantmentType
// that any of the Armour items have and returns it, or 0 if none of the items
// have the enchantment.
func (a *Armour) HighestEnchantmentLevel(t item.EnchantmentType) int {
	lvl := 0
	for _, it := range a.Items() {
		if e, ok := it.Enchantment(t); ok && e.Level() > lvl {
			lvl = e.Level()
		}
	}
	return lvl
}

// DamageFunc is a function that deals d damage points to an item stack s. The
// resulting item.Stack is returned. Depending on the game mode of a player,
// damage may not be dealt at all.
type DamageFunc func(s item.Stack, d int) item.Stack

// Damage deals damage (hearts) to Armour. The resulting item damage depends on the
// dmg passed and the DamageFunc used.
func (a *Armour) Damage(dmg float64, f DamageFunc) {
	armourDamage := int(math.Max(math.Floor(dmg/4), 1))
	for slot, it := range a.Slots() {
		_ = a.inv.SetItem(slot, f(it, armourDamage))
	}
}

// ThornsDamage checks if any of the Armour items are enchanted with thorns. If
// this is the case and the thorns enchantment activates (15% chance per level),
// a random Armour piece is damaged. The damage to be dealt to the attacker is
// returned.
func (a *Armour) ThornsDamage(f DamageFunc) float64 {
	slots := a.Slots()
	dmg := 0.0

	for _, i := range slots {
		thorns, _ := i.Enchantment(enchantment.Thorns)
		if level := float64(thorns.Level()); rand.Float64() < level*0.15 {
			// 15%/level chance of thorns activation per item. Total damage from
			// normal thorns armour (max thorns III) should never exceed 4.0 in
			// total.
			dmg = math.Min(dmg+float64(1+rand.IntN(4)), 4.0)
		}
	}
	if highest := a.HighestEnchantmentLevel(enchantment.Thorns); highest > 10 {
		// When we find an armour piece with thorns XI or above, the logic
		// changes: We have to find the armour piece with the highest level
		// of thorns and subtract 10 from its level to calculate the final
		// damage.
		dmg = float64(highest - 10)
	}
	if dmg > 0 {
		// Deal 2 damage to one random thorns item. Bedrock Edition and Java Edition
		// both have different behaviour here and neither seem to match the expected
		// behaviour. Java Edition deals 2 damage to a random thorns item for every
		// thorns armour item worn, while Bedrock Edition deals 1 additional damage
		// for every thorns item and another 2 for every thorns item when it
		// activates.
		slot := rand.IntN(len(slots))
		_ = a.Inventory().SetItem(slot, f(slots[slot], 2))
	}
	return dmg
}

// KnockBackResistance returns the combined knock back resistance of all Armour
// items. A value of 0 means normal knock back force, while a value of 1 means
// all knock back is ignored.
func (a *Armour) KnockBackResistance() float64 {
	resistance := 0.0
	for _, i := range a.Items() {
		if a, ok := i.Item().(item.Armour); ok {
			resistance += a.KnockBackResistance()
		}
	}
	return resistance
}

// Slots returns all items (including) air of the armour inventory in the order of helmet, chestplate, leggings,
// boots.
func (a *Armour) Slots() []item.Stack {
	return a.inv.Slots()
}

// Items returns a slice of all non-empty armour items equipped.
func (a *Armour) Items() []item.Stack {
	return a.inv.Items()
}

// Clear clears the armour inventory, removing all items currently present.
func (a *Armour) Clear() []item.Stack {
	return a.inv.Clear()
}

// String converts the armour to a readable string representation.
func (a *Armour) String() string {
	return fmt.Sprintf("(helmet: %v, chestplate: %v, leggings: %v, boots: %v)", a.Helmet(), a.Chestplate(), a.Leggings(), a.Boots())
}

// Inventory returns the underlying Inventory instance.
func (a *Armour) Inventory() *Inventory {
	return a.inv
}

// Handle assigns a Handler to an Armour inventory so that its methods are called for the respective events. Nil may be
// passed to set the default NopHandler.
// Handle is the equivalent of calling (*Armour).Inventory().H.
func (a *Armour) Handle(h Handler) {
	a.inv.Handle(h)
}

// Close closes the armour inventory, removing the slot change function.
func (a *Armour) Close() error {
	return a.inv.Close()
}
