package playerdb

import "github.com/df-mc/dragonfly/server/world"

const (
	overworld = uint8(iota)
	nether
	end
)

func dimensionToID(dimension world.Dimension) uint8 {
	switch dimension {
	case world.Nether:
		return nether
	case world.End:
		return end
	default:
		return overworld
	}
}

func idToDimension(mode uint8) world.Dimension {
	switch mode {
	case nether:
		return world.Nether
	case end:
		return world.End
	default:
		return world.Overworld
	}
}
