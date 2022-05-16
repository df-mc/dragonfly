package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// Fuel represents a block that can be used as fuel in a smelter, such as a blast furnace, furnace, or smoker.
type Fuel interface {
	// FuelInfo returns information of the block related to its fuel capabilities.
	FuelInfo() FuelInfo
}

// FuelInfo is a struct returned by blocks that implement Fuel. It contains information about the amount of fuel time
// it gives, and the residue created from burning the fuel.
type FuelInfo struct {
	// Duration returns the amount of time the fuel can be used to burn an input in a smelter.
	Duration time.Duration
	// Residue is the resulting item from burning the fuel in a smelter.
	Residue item.Stack
}
