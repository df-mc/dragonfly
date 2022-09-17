package item

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/slices"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
)

// Stack represents a stack of items. The stack shares the same item type and has a count which specifies the
// size of the stack.
type Stack struct {
	id int32

	item  world.Item
	count int

	customName string
	lore       []string

	damage int

	anvilCost int

	data map[string]any

	enchantments map[EnchantmentType]Enchantment
}

// NewStack returns a new stack using the item type and the count passed. NewStack panics if the count passed
// is negative or if the item type passed is nil.
func NewStack(t world.Item, count int) Stack {
	if count < 0 {
		panic("cannot use negative count for item stack")
	}
	if t == nil {
		panic("cannot have a stack with item type nil")
	}
	return Stack{item: t, count: count, id: newID()}
}

// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
// negative.
func (s Stack) Count() int {
	return s.count
}

// MaxCount returns the maximum count that the stack is able to hold when added to an inventory or when added
// to an item entity.
func (s Stack) MaxCount() int {
	if counter, ok := s.item.(MaxCounter); ok {
		return counter.MaxCount()
	}
	return 64
}

// Grow grows the Stack's count by n, returning the resulting Stack. If a positive number is passed, the stack
// is grown, whereas if a negative size is passed, the resulting Stack will have a lower count. The count of
// the returned Stack will never be negative.
func (s Stack) Grow(n int) Stack {
	s.count += n
	if s.count < 0 {
		s.count = 0
	}
	s.id = newID()
	return s
}

// Durability returns the current durability of the item stack. If the item is not one that implements the
// Durable interface, BaseDurability will always return -1.
// The closer the durability returned is to 0, the closer the item is to being broken.
func (s Stack) Durability() int {
	if durable, ok := s.Item().(Durable); ok {
		return durable.DurabilityInfo().MaxDurability - s.damage
	}
	return -1
}

// MaxDurability returns the maximum durability that the item stack is able to have. If the item does not
// implement the Durable interface, MaxDurability will always return -1.
func (s Stack) MaxDurability() int {
	if durable, ok := s.Item().(Durable); ok {
		return durable.DurabilityInfo().MaxDurability
	}
	return -1
}

// Damage returns a new stack that is damaged by the amount passed. (Meaning, its durability lowered by the
// amount passed.) If the item does not implement the Durable interface, the original stack is returned.
// The damage passed may be negative to add durability.
// If the final durability reaches 0 or below, the item returned is the resulting item of the breaking of the
// item. If the final durability reaches a number higher than the maximum durability, the stack returned will
// get the maximum durability.
func (s Stack) Damage(d int) Stack {
	durable, ok := s.Item().(Durable)
	if !ok {
		// Not a durable item.
		return s
	}
	durability := s.Durability()
	info := durable.DurabilityInfo()
	if durability-d <= 0 {
		if info.Persistent {
			// Persistent items can't be broken.
			return s
		}
		// A durability of 0, so the item is broken.
		return info.BrokenItem()
	}
	if durability-d > info.MaxDurability {
		// We've passed the maximum durability, so we just need to make sure the final durability of the item
		// will be equal to the max.
		s.damage, d = 0, 0
	}
	s.damage += d
	return s
}

// WithDurability returns a new item stack with the durability passed. If the item does not implement the
// Durable interface, WithDurability returns the original stack.
// The closer the durability d is to 0, the closer the item is to being broken. If a durability of 0 is passed,
// a stack with the item type of the BrokenItem is returned. If a durability is passed that exceeds the
// maximum durability, the stack returned will have the maximum durability.
func (s Stack) WithDurability(d int) Stack {
	durable, ok := s.Item().(Durable)
	if !ok {
		// Not a durable item.
		return s
	}
	maxDurability := durable.DurabilityInfo().MaxDurability
	if d > maxDurability {
		// A durability bigger than the max, so the item has no damage at all.
		s.damage = 0
		return s
	}
	if d == 0 {
		// A durability of 0, so the item is broken.
		return durable.DurabilityInfo().BrokenItem()
	}
	s.damage = maxDurability - d
	return s
}

// Empty checks if the stack is empty (has a count of 0).
func (s Stack) Empty() bool {
	return s.Count() == 0 || s.item == nil
}

// Item returns the item that the stack holds. If the stack is considered empty (Stack.Empty()), Item will
// always return nil.
func (s Stack) Item() world.Item {
	if s.Empty() || s.item == nil {
		return nil
	}
	return s.item
}

// AttackDamage returns the attack damage to the stack. By default, the value returned is 1.0. If the item
// held implements the item.Weapon interface, this damage may be different.
func (s Stack) AttackDamage() float64 {
	if weapon, ok := s.Item().(Weapon); ok {
		// Bonus attack damage from weapons is a bit quirky in Bedrock Edition: Even though tools say they
		// have, for example, + 5 Attack Damage, it is actually 1 + 5, while punching with a hand in Bedrock
		// Edition deals 2 damage, not 1 like in Java Edition.
		// The tooltip displayed in-game is therefore not exactly correct.
		return weapon.AttackDamage() + 1
	}
	return 1.0
}

// WithCustomName returns a copy of the Stack with the custom name passed. The custom name is formatted
// according to the rules of fmt.Sprintln.
func (s Stack) WithCustomName(a ...any) Stack {
	s.customName = format(a)
	if nameable, ok := s.Item().(nameable); ok {
		s.item = nameable.WithName(a...)
	}
	return s
}

// CustomName returns the custom name set for the Stack. An empty string is returned if the Stack has no
// custom name set.
func (s Stack) CustomName() string {
	return s.customName
}

// WithLore returns a copy of the Stack with the lore passed. Each string passed is put on a different line,
// where the first string is at the top and the last at the bottom.
// The lore may be cleared by passing no lines into the Stack.
func (s Stack) WithLore(lines ...string) Stack {
	s.lore = lines
	return s
}

// Lore returns the lore set for the Stack. If no lore is present, the slice returned has a len of 0.
func (s Stack) Lore() []string {
	return s.lore
}

// WithValue returns the current Stack with a value set at a specific key. This method may be used to
// associate custom data with the item stack, which will persist through server restarts.
// The value stored may later be obtained by making a call to Stack.Value().
//
// WithValue may be called with a nil value, in which case the value at the key will be cleared.
//
// WithValue stores Values by encoding them using the encoding/gob package. Users of WithValue must ensure
// that their value is valid for encoding with this package.
func (s Stack) WithValue(key string, val any) Stack {
	s.data = copyMap(s.data)
	if val != nil {
		s.data[key] = val
	} else {
		delete(s.data, key)
	}
	return s
}

// Value attempts to return a value set to the Stack using Stack.WithValue(). If a value is found by the key
// passed, it is returned and ok is true. If not found, the value returned is nil and ok is false.
func (s Stack) Value(key string) (val any, ok bool) {
	val, ok = s.data[key]
	return val, ok
}

// WithEnchantments returns the current stack with the passed enchantments. If an enchantment is not compatible
// with the item stack, it will not be applied.
func (s Stack) WithEnchantments(enchants ...Enchantment) Stack {
	if _, ok := s.item.(Book); ok {
		s.item = EnchantedBook{}
	}
	s.enchantments = copyEnchantments(s.enchantments)
	for _, enchant := range enchants {
		if _, ok := s.Item().(EnchantedBook); !ok && !enchant.t.CompatibleWithItem(s.item) {
			// Enchantment is not compatible with the item.
			continue
		}
		s.enchantments[enchant.t] = enchant
	}
	return s
}

// WithoutEnchantments returns the current stack but with the passed enchantments removed.
func (s Stack) WithoutEnchantments(enchants ...EnchantmentType) Stack {
	s.enchantments = copyEnchantments(s.enchantments)
	for _, enchant := range enchants {
		delete(s.enchantments, enchant)
	}
	if _, ok := s.item.(EnchantedBook); ok && len(s.enchantments) == 0 {
		s.item = Book{}
	}
	return s
}

// Enchantment attempts to return an Enchantment set to the Stack using Stack.WithEnchantment(). If an Enchantment
// is found by the EnchantmentType, the enchantment and the bool true is returned.
func (s Stack) Enchantment(enchant EnchantmentType) (Enchantment, bool) {
	ench, ok := s.enchantments[enchant]
	return ench, ok
}

// Enchantments returns an array of all Enchantments on the item. Enchantments returns the enchantments of a Stack in a
// deterministic order.
func (s Stack) Enchantments() []Enchantment {
	e := make([]Enchantment, 0, len(s.enchantments))
	for _, ench := range s.enchantments {
		e = append(e, ench)
	}
	sort.Slice(e, func(i, j int) bool {
		id1, _ := EnchantmentID(e[i].t)
		id2, _ := EnchantmentID(e[j].t)
		return id1 < id2
	})
	return e
}

// AnvilCost returns the number of experience levels to add to the base level cost when repairing, combining, or
// renaming this item with an anvil.
func (s Stack) AnvilCost() int {
	return s.anvilCost
}

// WithAnvilCost returns the current Stack with the anvil cost set to the passed value.
func (s Stack) WithAnvilCost(anvilCost int) Stack {
	i := s.Item()
	_, repairable := i.(Repairable)
	_, enchantedBook := i.(EnchantedBook)
	if !repairable && !enchantedBook {
		// This item can't have a repair cost.
		return s
	}
	s.anvilCost = anvilCost
	return s
}

// AddStack adds another stack to the stack and returns both stacks. The first stack returned will have as
// many items in it as possible to fit in the stack, according to a max count of either 64 or otherwise as
// returned by Item.MaxCount(). The second stack will have the leftover items: It may be empty if the count of
// both stacks together don't exceed the max count.
// If the two stacks are not comparable, AddStack will return both the original stack and the stack passed.
func (s Stack) AddStack(s2 Stack) (a, b Stack) {
	if s.Count() >= s.MaxCount() {
		// No more items could be added to the original stack.
		return s, s2
	}
	if !s.Comparable(s2) {
		// The items are not comparable and thus cannot be stacked together.
		return s, s2
	}
	diff := s.MaxCount() - s.Count()
	if s2.Count() < diff {
		diff = s2.Count()
	}

	s.count, s2.count = s.count+diff, s2.count-diff
	s.id, s2.id = newID(), newID()
	return s, s2
}

// Equal checks if the two stacks are equal. Equal is equivalent to a Stack.Comparable check while also
// checking the count and durability.
func (s Stack) Equal(s2 Stack) bool {
	return s.Comparable(s2) && s.count == s2.count && s.damage == s2.damage
}

// Comparable checks if two stacks can be considered comparable. True is returned if the two stacks have an
// equal item type and have equal enchantments, lore and custom names, or if one of the stacks is empty.
// Comparable does not check if the two stacks have the same durability.
func (s Stack) Comparable(s2 Stack) bool {
	if s.Empty() || s2.Empty() {
		return true
	}

	name, meta := s.Item().EncodeItem()
	name2, meta2 := s2.Item().EncodeItem()
	if name != name2 || meta != meta2 || s.anvilCost != s2.anvilCost || s.customName != s2.customName {
		return false
	}
	for !slices.Equal(s.lore, s2.lore) {
		return false
	}
	if len(s.enchantments) != len(s2.enchantments) {
		return false
	}
	for i := range s.enchantments {
		if s.enchantments[i] != s2.enchantments[i] {
			return false
		}
	}
	if !reflect.DeepEqual(s.data, s2.data) {
		return false
	}
	if nbt, ok := s.Item().(world.NBTer); ok {
		nbt2, ok := s2.Item().(world.NBTer)
		return ok && reflect.DeepEqual(nbt.EncodeNBT(), nbt2.EncodeNBT())
	}
	return true
}

// String implements the fmt.Stringer interface.
func (s Stack) String() string {
	if s.item == nil {
		return fmt.Sprintf("Stack<nil> x%v", s.count)
	}
	return fmt.Sprintf("Stack<%T%+v>(custom name='%v', lore='%v', damage=%v, anvilCost=%v) x%v", s.item, s.item, s.customName, s.lore, s.damage, s.anvilCost, s.count)
}

// Values returns all values associated with the stack by users. The map returned is a copy of the original:
// Modifying it will not modify the item stack.
func (s Stack) Values() map[string]any {
	return copyMap(s.data)
}

// stackID is a counter for unique stack IDs.
var stackID = new(int32)

// newID returns a new unique stack ID.
func newID() int32 {
	return atomic.AddInt32(stackID, 1)
}

// id returns the unique ID of the stack passed.
//lint:ignore U1000 Function is used through compiler directives.
//noinspection GoUnusedFunction
func id(s Stack) int32 {
	if s.Empty() {
		return 0
	}
	return s.id
}

// format is a utility function to format a list of Values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []any) string {
	return strings.TrimSuffix(fmt.Sprintln(a...), "\n")
}

// copyMap makes a copy of the map passed. It does not recursively copy the map.
func copyMap(m map[string]any) map[string]any {
	cp := make(map[string]any, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// copyEnchantments makes a copy of the enchantments map passed. It does not recursively copy the map.
func copyEnchantments(m map[EnchantmentType]Enchantment) map[EnchantmentType]Enchantment {
	cp := make(map[EnchantmentType]Enchantment, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
