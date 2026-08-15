package item

// SwingDuration represents an item with a custom duration in seconds of the swing animation played when the
// item is used to attack.
type SwingDuration interface {
	// SwingDuration returns the duration in seconds of the swing animation.
	SwingDuration() float64
}

// SwingSounds represents an item with custom sounds played when the item is swung.
type SwingSounds interface {
	// SwingSounds returns the sounds played when the item is swung.
	SwingSounds() SwingSoundsInfo
}

// SwingSoundsInfo is a struct returned by items that implement SwingSounds. It contains the sounds played
// when the item is swung.
type SwingSoundsInfo struct {
	// AttackHit is the sound played when an attack made with the item hits.
	AttackHit string
	// AttackMiss is the sound played when an attack made with the item misses.
	AttackMiss string
}
