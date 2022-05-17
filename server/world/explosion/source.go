package explosion

import "github.com/df-mc/dragonfly/server/world"

// Source is the source of the explosion. This can range from creepers, TNT, and other sources.
type Source interface {
	// ExplodeInfo returns the explosion information for the source.
	ExplodeInfo() *Explosion
}

// TNTSource is an explosion source that is caused by TNT.
type TNTSource struct {
	World *world.World
}

// ExplodeInfo ...
func (t TNTSource) ExplodeInfo() *Explosion {
	return &Explosion{
		w:     t.World,
		power: 4,
		fire:  false,
	}
}

type Fireball struct {
	World *world.World
}

func (f Fireball) ExplodeInfo() *Explosion {
	return &Explosion{
		w:     f.World,
		power: 1,
		fire:  true,
	}
}
