package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"time"
)

var (
	// Overworld is the Dimension implementation of a normal overworld. It has a blue sky under normal circumstances and
	// has a sun, clouds, stars and a moon. Overworld has a building range of [-64, 320].
	Overworld overworld
	// Nether is a Dimension implementation with a lower base light level and a darker sky without sun/moon. It has a
	// building range of [0, 256].
	Nether nether
	// End is a Dimension implementation with a dark sky. It has a building range of [0, 256].
	End end
)

type (
	// Dimension is a dimension of a World. It influences a variety of properties of a World such as the building range,
	// the sky colour and the behaviour of liquid blocks.
	Dimension interface {
		Range() cube.Range
		EncodeDimension() int
		WaterEvaporates() bool
		LavaSpreadDuration() time.Duration
		WeatherCycle() bool
		TimeCycle() bool
	}
	overworld struct{}
	nether    struct{}
	end       struct{}
)

func (overworld) Range() cube.Range                 { return cube.Range{-64, 320} }
func (overworld) EncodeDimension() int              { return 0 }
func (overworld) WaterEvaporates() bool             { return false }
func (overworld) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }
func (overworld) WeatherCycle() bool                { return true }
func (overworld) TimeCycle() bool                   { return true }
func (overworld) String() string                    { return "Overworld" }

func (nether) Range() cube.Range                 { return cube.Range{0, 256} }
func (nether) EncodeDimension() int              { return 1 }
func (nether) WaterEvaporates() bool             { return true }
func (nether) LavaSpreadDuration() time.Duration { return time.Second / 4 }
func (nether) WeatherCycle() bool                { return false }
func (nether) TimeCycle() bool                   { return false }
func (nether) String() string                    { return "Nether" }

func (end) Range() cube.Range                 { return cube.Range{0, 256} }
func (end) EncodeDimension() int              { return 2 }
func (end) WaterEvaporates() bool             { return false }
func (end) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }
func (end) WeatherCycle() bool                { return false }
func (end) TimeCycle() bool                   { return false }
func (end) String() string                    { return "End" }
