package explosion

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Explosion struct {
	// W is the world where the explosion will occur in. If none is given, the explosion will not occur.
	w *world.World

	// Power is the power of the explosion. The higher the power, the bigger the explosion.
	power float64

	// Fire will cause the explosion to create fire in 1/3 of the air in the explosion that are above opaque blocks
	fire bool

	// TODO: Damage entities
}

var rays int8 = 16

// Explode explodes the blocks in the center of the provided vec3
func (e Explosion) Explode(center mgl64.Vec3) {
	if e.w == nil {
		return
	}

	if e.power < 0 {
		return
	}

}
