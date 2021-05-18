package armour

// Armour represents an item that may be worn as armour. Generally, these items provide armour points, which
// reduce damage taken. Some pieces of armour also provide toughness, which negates damage proportional to
// the total damage dealt.
type Armour interface {
	// DefencePoints returns the defence points that the armour provides when worn.
	DefencePoints() float64
	// KnockBackResistance returns a number from 0-1 that decides the amount of knock back force that is
	// resisted upon being attacked. 1 knock back resistance point client-side translates to 10% knock back
	// reduction.
	KnockBackResistance() float64
}

// Helmet is an Armour item that can be worn in the helmet slot.
type Helmet interface {
	Armour
	Helmet() bool
}

// Chestplate is an Armour item that can be worn in the chestplate slot.
type Chestplate interface {
	Armour
	Chestplate() bool
}

// Leggings are an Armour item that can be worn in the leggings slot.
type Leggings interface {
	Armour
	Leggings() bool
}

// Boots are an Armour item that can be worn in the boots slot.
type Boots interface {
	Armour
	Boots() bool
}
