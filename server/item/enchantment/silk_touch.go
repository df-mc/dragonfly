package enchantment

import "github.com/df-mc/dragonfly/server/item"

// SilkTouch is an enchantment that allows many blocks to drop themselves instead of their usual items when mined.
type SilkTouch struct{ enchantment }

// Name ...
func (e SilkTouch) Name() string {
	return "Silk Touch"
}

// MaxLevel ...
func (e SilkTouch) MaxLevel() int {
	return 1
}

// WithLevel ...
func (e SilkTouch) WithLevel(level int) item.Enchantment {
	return SilkTouch{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e SilkTouch) CompatibleWith(s item.Stack) bool {
	it := s.Item()
	_, pickaxe := it.(item.Pickaxe)
	_, axe := it.(item.Axe)
	_, shovel := it.(item.Shovel)
	_, hoe := it.(item.Hoe)
	_, shears := it.(item.Shears)
	//TODO: Fortune
	return pickaxe || axe || shovel || hoe || shears
}
