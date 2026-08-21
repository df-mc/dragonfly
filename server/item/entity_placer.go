package item

// EntityPlacer represents an item that can place entities into the world, such as spawn eggs.
type EntityPlacer interface {
	// EntityPlacerInfo returns the entity placer information of the item.
	EntityPlacerInfo() EntityPlacerInfo
}

// EntityPlacerInfo is a struct returned by items that implement EntityPlacer. It contains the
// entity placement configuration of the item.
type EntityPlacerInfo struct {
	// Entity is the identifier of the entity that will be placed.
	Entity string
	// UseOn is a list of block identifiers that this item can be used on. If empty, all blocks
	// are allowed.
	UseOn []string
	// DispenseOn is a list of block identifiers that this item can be dispensed on. If empty,
	// all blocks are allowed.
	DispenseOn []string
}
