package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"time"
)

var (
	// Overworld is the Dimension implementation of a normal overworld. It has a
	// blue sky under normal circumstances and has a sun, clouds, stars and a
	// moon. Overworld has a building range of [-64, 320).
	Overworld overworld
	// Nether is a Dimension implementation with a lower base light level and a
	// darker sky without sun/moon. It has a building range of [0, 128).
	Nether nether
	// End is a Dimension implementation with a dark sky. It has a building
	// range of [0, 256).
	End end
)

var dimensionReg = newDimensionRegistry(map[int]Dimension{
	0: Overworld,
	1: Nether,
	2: End,
})

// DimensionByID looks up a Dimension for the ID passed, returning Overworld
// for 0, Nether for 1 and End for 2. If the ID is unknown, the bool returned
// is false. In this case the Dimension returned is Overworld.
func DimensionByID(id int) (Dimension, bool) {
	return dimensionReg.Lookup(id)
}

// DimensionID looks up the ID that a Dimension was registered with. If not
// found, false is returned.
func DimensionID(dim Dimension) (int, bool) {
	return dimensionReg.LookupID(dim)
}

type dimensionRegistry struct {
	dimensions map[int]Dimension
	ids        map[Dimension]int
}

// newDimensionRegistry returns an initialised dimensionRegistry.
func newDimensionRegistry(dim map[int]Dimension) *dimensionRegistry {
	ids := make(map[Dimension]int, len(dim))
	for k, v := range dim {
		ids[v] = k
	}
	return &dimensionRegistry{dimensions: dim, ids: ids}
}

// Lookup looks up a Dimension for the ID passed, returning Overworld for 0,
// Nether for 1 and End for 2. If the ID is unknown, the bool returned is
// false. In this case the Dimension returned is Overworld.
func (reg *dimensionRegistry) Lookup(id int) (Dimension, bool) {
	dim, ok := reg.dimensions[id]
	if !ok {
		dim = Overworld
	}
	return dim, ok
}

// LookupID looks up the ID that a Dimension was registered with. If not found,
// false is returned.
func (reg *dimensionRegistry) LookupID(dim Dimension) (int, bool) {
	id, ok := reg.ids[dim]
	return id, ok
}

type (
	// Dimension is a dimension of a World. It influences a variety of
	// properties of a World such as the building range, the sky colour and the
	// behaviour of liquid blocks.
	Dimension interface {
		// Range returns the lowest and highest valid Y coordinates of a block
		// in the Dimension.
		Range() cube.Range
		WaterEvaporates() bool
		LavaSpreadDuration() time.Duration
		WeatherCycle() bool
		TimeCycle() bool
	}
	overworld struct{}
	nether    struct{}
	end       struct{}
)

func (overworld) Range() cube.Range                 { return cube.Range{-64, 319} }
func (overworld) WaterEvaporates() bool             { return false }
func (overworld) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }
func (overworld) WeatherCycle() bool                { return true }
func (overworld) TimeCycle() bool                   { return true }
func (overworld) String() string                    { return "Overworld" }

func (nether) Range() cube.Range                 { return cube.Range{0, 127} }
func (nether) WaterEvaporates() bool             { return true }
func (nether) LavaSpreadDuration() time.Duration { return time.Second / 4 }
func (nether) WeatherCycle() bool                { return false }
func (nether) TimeCycle() bool                   { return false }
func (nether) String() string                    { return "Nether" }

func (end) Range() cube.Range                 { return cube.Range{0, 255} }
func (end) WaterEvaporates() bool             { return false }
func (end) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }
func (end) WeatherCycle() bool                { return false }
func (end) TimeCycle() bool                   { return false }
func (end) String() string                    { return "End" }
