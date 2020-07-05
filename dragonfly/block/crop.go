package block

// Crop is an interface for all plants that are grown. A crop uses random chances on ticks to make sure that the crop should grow.
// A crop doesn't necessarily have to be on a farmland block, as there are non-hydrated plants like trees and cactus that are planted on grass
type Crop interface {
	// LightLevelRequired is the light level required for the crop to grow.
	LightLevelRequired() uint8

	// RequiresHydration is if the crop requires block Hydration in order to grow.
	RequiresHydration() bool

	// RequiresFarmland is if the crop requires farmland to be able to grow.
	RequiresFarmland() bool
}
