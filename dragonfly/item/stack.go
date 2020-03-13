package item

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"reflect"
)

// Stack represents a stack of items. The stack shares the same item type and has a count which specifies the
// size of the stack.
type Stack struct {
	item  world.Item
	count int
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
	return Stack{item: t, count: count}
}

// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
// negative.
func (s Stack) Count() int {
	return s.count
}

// MaxCount returns the maximum count that the stack is able to hold when added to an inventory or when added
// to an item entity.
func (s Stack) MaxCount() int {
	if counter, ok := s.Item().(MaxCounter); ok {
		return counter.MaxCount()
	}
	return 64
}

// Empty checks if the stack is empty (has a count of 0). If this is the case
func (s Stack) Empty() bool {
	return s.Count() == 0
}

// Item returns the item that the stack holds. If the stack is considered empty (Stack.Empty()), Item will
// always return nil.
func (s Stack) Item() world.Item {
	if s.Empty() || s.item == nil {
		return nil
	}
	return s.item
}

// AttackDamage returns the attack damage of the stack. By default, the value returned is 2.0. If the item
// held implements the item.Weapon interface, this damage may be different.
func (s Stack) AttackDamage() float32 {
	if weapon, ok := s.Item().(Weapon); ok {
		return weapon.AttackDamage()
	}
	return 2.0
}

// AddStack adds another stack to the stack and returns both stacks. The first stack returned will have as
// many items in it as possible to fit in the stack, according to a max count of either 64 or otherwise as
// returned by Item.MaxCount(). The second stack will have the leftover items: It may be empty if the count of
// both stacks together don't exceed the max count.
// If the two stacks are not comparable, AddStack will return both the original stack and the stack passed.
func (s Stack) AddStack(s2 Stack) (a, b Stack) {
	if !s.Comparable(s2) {
		// The items are not comparable and thus cannot be stacked together.
		return s, s2
	}
	if s.Count() >= s.MaxCount() {
		// No more items could be added to the original stack.
		return s, s2
	}
	diff := s.MaxCount() - s.Count()
	if s2.Count() < diff {
		diff = s2.Count()
	}

	s.count, s2.count = s.count+diff, s2.count-diff
	return s, s2
}

// Grow grows the Stack's count by n, returning the resulting Stack. If a positive number is passed, the stack
// is grown, whereas if a negative size is passed, the resulting Stack will have a lower count. The count of
// the returned Stack will never be negative.
func (s Stack) Grow(n int) Stack {
	s.count += n
	if s.count < 0 {
		s.count = 0
	}
	return s
}

// Comparable checks if two stacks can be considered comparable. True is returned if the two stacks have an
// equal item type and have equal enchantments, lore and custom names, or if one of the stacks is empty.
func (s Stack) Comparable(s2 Stack) bool {
	if s.count == 0 || s2.count == 0 {
		return true
	}
	// Make sure the counts are equal so that we can deep compare.
	s.count = s2.count
	return reflect.DeepEqual(s, s2)
}

// String implements the fmt.Stringer interface.
func (s Stack) String() string {
	if s.item == nil {
		return fmt.Sprintf("Stack<nil> x%v", s.count)
	}
	return fmt.Sprintf("Stack<%T%+v>x%v", s.item, s.item, s.count)
}
