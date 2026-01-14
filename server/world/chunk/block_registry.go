package chunk

// BlockRegistry provides the minimal block registry API required by the chunk package.
//
// Implementations must be safe for concurrent read access after construction/finalization, as chunks and chunk
// encoding/decoding may happen from multiple goroutines.
type BlockRegistry interface {
	BlockCount() int
	AirRuntimeID() uint32
	RuntimeIDToState(runtimeID uint32) (name string, properties map[string]any, found bool)
	StateToRuntimeID(name string, properties map[string]any) (runtimeID uint32, found bool)
	FilteringBlock(rid uint32) uint8
	LightBlock(rid uint32) uint8
	RandomTickBlock(rid uint32) bool
	NBTBlock(rid uint32) bool
	LiquidDisplacingBlock(rid uint32) bool
	LiquidBlock(rid uint32) bool
	HashToRuntimeID(hash uint32) (rid uint32, ok bool)
}
