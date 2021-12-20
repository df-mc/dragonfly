package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"time"
)

var (
	Overworld overworld
	Nether    nether
	End       end
)

type (
	Dimension interface {
		Range() cube.Range
		EncodeDimension() int
		WaterEvaporates() bool
		LavaSpreadDuration() time.Duration
	}
	overworld struct{}
	nether    struct{}
	end       struct{}
)

func (overworld) Range() cube.Range                 { return cube.Range{-64, 320} }
func (overworld) EncodeDimension() int              { return 0 }
func (overworld) WaterEvaporates() bool             { return false }
func (overworld) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }

func (nether) Range() cube.Range                 { return cube.Range{0, 256} }
func (nether) EncodeDimension() int              { return 1 }
func (nether) WaterEvaporates() bool             { return true }
func (nether) LavaSpreadDuration() time.Duration { return time.Second / 4 }

func (end) Range() cube.Range                 { return cube.Range{0, 256} }
func (end) EncodeDimension() int              { return 2 }
func (end) WaterEvaporates() bool             { return false }
func (end) LavaSpreadDuration() time.Duration { return time.Second * 3 / 2 }
