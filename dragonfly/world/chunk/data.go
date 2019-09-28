package chunk

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"sync"
)

const (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written as over network and on disk.
	SubChunkVersion = 8
)

// SerialisedData holds the serialised data of a chunk. It consists of the chunk's block data itself, a height
// map, the biomes and entities and block entities.
type SerialisedData struct {
	// sub holds the data of the serialised sub chunks in a chunk. Sub chunks that are empty or that otherwise
	// don't exist are represented as an empty slice (or technically, nil).
	SubChunks [16][]byte
	// Data2D is the 2D data of the chunk, which is composed of the biome IDs (256 bytes) and optionally the
	// height map of the chunk.
	Data2D []byte
	// BlockNBT is an encoded NBT array of all blocks that carry additional NBT, such as chests, with all
	// their contents.
	BlockNBT []byte
}

// pool is used to pool byte buffers used for encoding chunks.
var pool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

// NetworkEncode encodes a chunk passed to its network representation and returns it as a SerialisedData,
// which may be sent over network.
func NetworkEncode(c *Chunk) (d SerialisedData) {
	buf := pool.Get().(*bytes.Buffer)

	for y, sub := range c.sub {
		if sub == nil {
			// No need to put empty sub chunks in the SerialisedData.
			continue
		}
		_ = buf.WriteByte(SubChunkVersion)
		_ = buf.WriteByte(byte(len(sub.storages)))
		for _, storage := range sub.storages {
			_ = buf.WriteByte(byte(storage.bitsPerBlock<<1) | 1)
			for _, word := range storage.blocks {
				_ = binary.Write(buf, binary.LittleEndian, word)
			}
			_ = protocol.WriteVarint32(buf, int32(storage.palette.Len()))
			for _, runtimeID := range storage.palette.blockRuntimeIDs {
				_ = protocol.WriteVarint32(buf, int32(runtimeID))
			}
		}
		d.SubChunks[y] = make([]byte, buf.Len())
		_, _ = buf.Read(d.SubChunks[y])
	}
	d.Data2D = append(c.biomes[:], 0)

	buf.Reset()
	pool.Put(buf)
	return
}
