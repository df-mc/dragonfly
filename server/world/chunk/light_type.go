package chunk

var (
	// SkyLight holds a light implementation that can be used for propagating sky light through a sub chunk.
	SkyLight skyLight
	// BlockLight holds a light implementation that can be used for propagating block light through a sub chunk.
	BlockLight blockLight
)

var (
	// LightBlocks is a list of block light levels (0-15) indexed by block runtime IDs. The map is used to do a
	// fast lookup of block light.
	LightBlocks = make([]uint8, 0, 7000)
	// FilteringBlocks is a map for checking if a block runtime ID filters light, and if so, how many levels.
	// Light is able to propagate through these blocks, but will have its level reduced.
	FilteringBlocks = make([]uint8, 0, 7000)
)

type (
	// light is a type that can be used to set and retrieve light of a specific type in a sub chunk.
	light interface {
		light(sub *SubChunk, x, y, z uint8) uint8
		setLight(sub *SubChunk, x, y, z, v uint8)
	}
	skyLight   struct{}
	blockLight struct{}
)

func (skyLight) light(sub *SubChunk, x, y, z uint8) uint8   { return sub.SkyLight(x, y, z) }
func (skyLight) setLight(sub *SubChunk, x, y, z, v uint8)   { sub.SetSkyLight(x, y, z, v) }
func (blockLight) light(sub *SubChunk, x, y, z uint8) uint8 { return sub.BlockLight(x, y, z) }
func (blockLight) setLight(sub *SubChunk, x, y, z, v uint8) { sub.SetBlockLight(x, y, z, v) }
