package block

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math/rand"
)

type Crop interface {
	// LightLevelRequired is the light level required for the crop to grow.
	LightLevelRequired() uint8

	// RequiresHydration is if the crop requires block Hydration in order to grow.
	RequiresHydration() bool

	// RequiresFarmland is if the crop requires farmland to be able to grow.
	RequiresFarmland() bool

	// Grow is ran every random tick to try and grow the crop 1 stage.
	// Growth stages are handled inside of this method, making crops easier to create on a basis.
	// LightLevel and Hydration are passed to make crops in better conditions grow faster.
	Grow(pos world.BlockPos, w *world.World, r *rand.Rand, Hydration uint8)
}
