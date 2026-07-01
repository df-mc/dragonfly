package chunk

var (
	// SkyLight holds a light implementation that can be used for propagating sky light through a sub chunk.
	SkyLight skyLight
	// BlockLight holds a light implementation that can be used for propagating block light through a sub chunk.
	BlockLight blockLight
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
