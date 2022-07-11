package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Mending is an enchantment that repairs the item when experience orbs are collected.
type Mending struct{}

// Name ...
func (Mending) Name() string {
	return "Mending"
}

// MaxLevel ...
func (Mending) MaxLevel() int {
	return 1
}

// CompatibleWith ...
func (Mending) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Durable)
	//_, infinity := s.Enchantment(Infinity{}) todo: infinity
	return ok // && !infinity
}
