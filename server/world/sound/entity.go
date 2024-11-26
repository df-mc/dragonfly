package sound

// Attack is a sound played when an entity, most notably a player, attacks another entity.
type Attack struct {
	// Damage specifies if the attack actually dealt damage to the other entity. If set to false, the sound
	// will be different from when set to true.
	Damage bool

	sound
}

// Drowning is a sound played when an entity is drowning in water.
type Drowning struct{ sound }

// Burning is a sound played when an entity is on fire.
type Burning struct{ sound }

// Fall is a sound played when an entity falls and hits ground.
type Fall struct {
	// Distance is the distance the entity has fallen.
	Distance float64

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

// LevelUp is a sound played for a player whenever they level up.
type LevelUp struct{ sound }

// Experience is a sound played whenever a player picks up an XP orb.
type Experience struct{ sound }

// GhastWarning is a sound played when a ghast is ready to attack.
type GhastWarning struct{ sound }

// GhastShoot is a sound played when a ghast shoots a fire charge.
type GhastShoot struct{ sound }

// FireworkLaunch is a sound played when a firework is launched.
type FireworkLaunch struct{ sound }

// FireworkHugeBlast is a sound played when a huge sphere firework explodes.
type FireworkHugeBlast struct{ sound }

// FireworkBlast is a sound played when a small sphere firework explodes.
type FireworkBlast struct{ sound }

// FireworkTwinkle is a sound played when a firework explodes and should twinkle.
type FireworkTwinkle struct{ sound }
