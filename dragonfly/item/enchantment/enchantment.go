package enchantment

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"

// enchantment is used internally to store the level of enchantments. It implements the Level method so
// parent enchantments do not need to implement it themselves.
type enchantment struct {
	Lvl int
}

// Level ...
func (e enchantment) Level() int {
	return e.Lvl
}

// withLevel returns the same enchantment instance but with the level provided. If the provided level
// is greater than the enchantment's maximum level, it will return the enchantment with the maximum level.
func (e enchantment) withLevel(level int, ench item.Enchantment) enchantment {
	if level > ench.MaxLevel() {
		e.Lvl = ench.MaxLevel()
		return e
	}
	e.Lvl = level
	return e
}
