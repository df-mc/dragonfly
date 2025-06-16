package item

import (
	"time"

	"github.com/df-mc/dragonfly/server/world"
)

// Smeltable represents an item that can be input into a smelter, such as a blast furnace, furnace, or smoker, to cook and
// transform it into a different item.
type Smeltable interface {
	// SmeltInfo returns information of the item related to it's smelting capabilities.
	SmeltInfo() SmeltInfo
}

// Fuel represents an item that can be used as fuel in a smelter, such as a blast furnace, furnace, or smoker.
type Fuel interface {
	// FuelInfo returns information of the item related to its fuel capabilities.
	FuelInfo() FuelInfo
}

// SmeltInfo is a struct returned by items that implement Smeltable. It contains information about the product, experience
// gained, and more.
type SmeltInfo struct {
	// Product returns the resulting item stack from smelting the item.
	Product Stack
	// Experience returns the experience gained from performing the smelt, alongside the Product.
	Experience float64
	// Food returns true if the smelt is food, for smelters such as smokers or regular furnaces.
	Food bool
	// Ores returns true if the smelt is ores, for smelters such as blast furnaces or regular furnaces.
	Ores bool
}

// newSmeltInfo returns a new SmeltInfo with the given values.
func newSmeltInfo(product Stack, experience float64) SmeltInfo {
	return SmeltInfo{
		Product:    product,
		Experience: experience,
	}
}

// newFoodSmeltInfo returns a new SmeltInfo with the given values that allows smelting in a smelter.
func newFoodSmeltInfo(product Stack, experience float64) SmeltInfo {
	return SmeltInfo{
		Product:    product,
		Experience: experience,
		Food:       true,
	}
}

// newOreSmeltInfo returns a new SmeltInfo with the given values that allows smelting in a blast furnace.
func newOreSmeltInfo(product Stack, experience float64) SmeltInfo {
	return SmeltInfo{
		Product:    product,
		Experience: experience,
		Ores:       true,
	}
}

// FuelInfo is a struct returned by items that implement Fuel. It contains information about the amount of fuel time
// it gives, and the residue created from burning the fuel.
type FuelInfo struct {
	// Duration returns the amount of time the fuel can be used to burn an input in a smelter.
	Duration time.Duration
	// Residue is the resulting item from burning the fuel in a smelter.
	Residue world.ItemStack
}

// WithResidue returns a new FuelInfo with a residue.
func (f FuelInfo) WithResidue(residue Stack) FuelInfo {
	f.Residue = residue
	return f
}

// newFuelInfo returns a new FuelInfo with the given values.
func newFuelInfo(duration time.Duration) FuelInfo {
	return FuelInfo{Duration: duration}
}
