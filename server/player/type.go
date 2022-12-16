package player

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Type is a world.EntityType implementation for Player.
type Type struct{}

func (Type) EncodeEntity() string   { return "minecraft:player" }
func (Type) NetworkOffset() float64 { return 1.62 }
func (Type) BBox(e world.Entity) cube.BBox {
	p := e.(*Player)
	s := p.Scale()

	// TODO: Shrink BBox for sneaking once implemented in Bedrock Edition. This is already a thing in Java Edition.
	gliding := p.Gliding()
	swimming := p.Swimming()
	_, sleeping := p.Sleeping()
	switch {
	case sleeping:
		return cube.Box(-0.1*s, 0, -0.1*s, 0.1*s, 0.2*s, 0.1*s)
	case gliding, swimming:
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 0.6*s, 0.3*s)
	default:
		return cube.Box(-0.3*s, 0, -0.3*s, 0.3*s, 1.8*s, 0.3*s)
	}
}
