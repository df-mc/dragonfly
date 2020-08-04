package sound

// Attack is a sound played when an entity, most notably a player, attacks another entity.
type Attack struct {
	// Damage specifies if the attack actually dealt damage to the other entity. If set to false, the sound
	// will be different than when set to true.
	Damage bool

	sound
}

// Burp is a sound played when a player finishes eating an item.
type Burp struct{ sound }
