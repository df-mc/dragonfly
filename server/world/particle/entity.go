package particle

import "image/color"

// HugeExplosion is a particle shown when TNT or a creeper explodes.
type HugeExplosion struct{ particle }

// EndermanTeleportParticle is a particle that shows up when an enderman teleports.
type EndermanTeleportParticle struct{ particle }

// SnowballPoof is a particle shown when a snowball collides with something.
type SnowballPoof struct{ particle }

// Splash is a particle that shows up when a splash potion is splashed.
type Splash struct {
	particle

	// Colour is the colour that should be splashed.
	Colour color.RGBA
}
