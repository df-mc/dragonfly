package item

// Mace is a melee weapon that can do smash attacks which deal more damage the farther the user falls before hitting an
// entity. Landing a successful smash attack causes the user to land without taking any fall damage accumulated before
// attacking.
type Mace struct{}

// MaxCount always returns 1.
func (Mace) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (Mace) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    501,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 1,
		BreakDurability:  1,
	}
}

// AttackDamage ...
func (Mace) AttackDamage() float64 {
	return 5
}

// EncodeItem ...
func (Mace) EncodeItem() (name string, meta int16) {
	return "minecraft:mace", 0
}
