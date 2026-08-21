package item

// Seed represents an item that can be planted on blocks to grow a crop.
type Seed interface {
	// SeedInfo returns the information of the item related to planting it as a seed.
	SeedInfo() SeedInfo
}

// SeedInfo is a struct returned by items that implement Seed. It contains the information required for the
// client to allow planting the item on the specified blocks.
type SeedInfo struct {
	// CropResult is the identifier of the crop that grows when the seed is planted.
	CropResult string
	// PlantAt is a list of block identifiers that the seed may be planted on. This list is ignored if
	// PlantAtAnySolidSurface is true.
	PlantAt []string
	// PlantAtAnySolidSurface is true if the seed may be planted on any solid surface.
	PlantAtAnySolidSurface bool
	// PlantAtFace is the face the seed is planted on, such as "up" or "down".
	PlantAtFace string
}
