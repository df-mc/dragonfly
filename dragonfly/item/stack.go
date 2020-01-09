package item

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"sync/atomic"
)

// Stack represents a stack of items. The stack shares the same item type and has a count which specifies the
// size of the stack.
type Stack struct {
	item  Item
	count uint32
}

func NewStack(t Item, count int) Stack {
	return Stack{item: t, count: uint32(count)}
}

// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
// negative.
func (s Stack) Count() int {
	return int(atomic.LoadUint32(&s.count))
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
