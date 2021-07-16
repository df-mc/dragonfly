package block_internal

//lint:file-ignore ST1022 Exported variables in this package have compiler directives. These variables are not otherwise exposed to users.

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	_ "unsafe" // Imported for compiler directives.
)

//go:linkname world_breakParticle github.com/df-mc/dragonfly/server/world.breakParticle
//noinspection ALL
var world_breakParticle func(b world.Block) world.Particle

func init() {
	world_breakParticle = func(b world.Block) world.Particle {
		return particle.BlockBreak{Block: b}
	}
}
