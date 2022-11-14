package playerdb

import "github.com/df-mc/dragonfly/server/world"

const (
	overworld = uint8(iota)
	nether
	end
)

func idToDimension(mode uint8) world.Dimension {
	switch mode {
	case nether:
		return world.Nether
	case end:
		return world.End
	case overworld:
		return world.Overworld
	}
	panic("should never happen")
}
