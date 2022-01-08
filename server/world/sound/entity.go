package sound

// Attack is a sound played when an entity, most notably a player, attacks another entity.
type Attack struct {
	// Damage specifies if the attack actually dealt damage to the other entity. If set to false, the sound
	// will be different from when set to true.
	Damage bool

	sound
}

// Burp is a sound played when a player finishes eating an item.
type Burp struct{ sound }

// Pop is a sound played when a chicken lays an egg.
type Pop struct{ sound }

// Explosion is a sound played when an explosion happens, such as from a creeper or TNT.
type Explosion struct{ sound }

// Thunder is a sound played when lightning strikes the ground.
type Thunder struct{ sound }
