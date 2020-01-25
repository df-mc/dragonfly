package item

import (
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"reflect"
)

// Stack represents a stack of items. The stack shares the same item type and has a count which specifies the
// size of the stack.
type Stack struct {
	item  Item
	count int
}

func NewStack(t Item, count int) Stack {
	if count < 0 {
		panic("cannot use negative count for item stack")
	}
	return Stack{item: t, count: count}
}

// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
// negative.
func (s Stack) Count() int {
	return s.count
}

// Empty checks if the stack is empty (has a count of 0). If this is the case
func (s Stack) Empty() bool {
	return s.Count() == 0
}

// Item returns the item that the stack holds. If the stack is considered empty (Stack.Empty()), Item will
// always return block.Air.
func (s Stack) Item() Item {
	if s.Empty() || s.item == nil {
		return block.Air{}
	}
	return s.item
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
	max := 64
	if counter, ok := s.Item().(MaxCounter); ok {
		max = counter.MaxCount()
	}
	if s.Count() >= max {
		// No more items could be added to the original stack.
		return s, s2
	}
	diff := max - s.Count()
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
// equal item type and have equal enchantments, lore and custom names.
func (s Stack) Comparable(s2 Stack) bool {
	if s.count == 0 {
		s.item = s2.item
	}
	if s2.count == 0 {
		s2.item = s.item
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
