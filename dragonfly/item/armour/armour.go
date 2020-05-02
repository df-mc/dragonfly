package armour

// Armour represents an item that may be worn as armour. Generally, these items provide armour points, which
// reduce damage taken. Some pieces of armour also provide toughness, which negates damage proportional to
// the total damage dealt.
type Armour interface {
	// DefencePoints returns the defence points that the armour provides when worn.
	DefencePoints() float64
}
