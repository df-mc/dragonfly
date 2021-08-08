package chunk

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"sync"
)

const (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written on disk and over network.
	SubChunkVersion = 8
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	CurrentBlockVersion int32 = 17825806
)

var (
	// RuntimeIDToState must hold a function to convert a runtime ID to a name and its state properties.
	RuntimeIDToState func(runtimeID uint32) (name string, properties map[string]interface{}, found bool)
	// emptyHeightMap holds an empty height map. It is written as 256 int16s, or 512 bytes.
	emptyHeightMap = make([]byte, 512)
	// pool is used to pool byte buffers used for encoding chunks.
	pool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	}
)

type (
	// SerialisedData holds the serialised data of a chunk. It consists of the chunk's block data itself, a height
	// map, the biomes and entities and block entities.
	SerialisedData struct {
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
	// blockEntry represents a block as found in a disk save of a world.
	blockEntry struct {
		Name    string                 `nbt:"name"`
		State   map[string]interface{} `nbt:"states"`
		Version int32                  `nbt:"version"`
	}
)

// NetworkEncode encodes a chunk passed to its network representation and returns it as a SerialisedData,
// which may be sent over network.
func NetworkEncode(c *Chunk) (d SerialisedData) {
	buf := pool.Get().(*bytes.Buffer)
	d = encodeSubChunks(buf, c, networkEncodeBlockStorage)
	d.Data2D = append(c.biomes[:], 0)

	enc := nbt.NewEncoder(buf)
	for _, data := range c.blockEntities {
		_ = enc.Encode(data)
	}
	d.BlockNBT = append([]byte(nil), buf.Bytes()...)

	buf.Reset()
	pool.Put(buf)
	return
}

// DiskEncode encodes a chunk to its disk representation, so that it may be stored in a database, giving other
// servers the ability to read the chunk.
func DiskEncode(c *Chunk) (d SerialisedData) {
	buf := pool.Get().(*bytes.Buffer)
	d = encodeSubChunks(buf, c, diskEncodeBlockStorage)
	// We simply write a zero slice for the height map, as there is little profit of writing it here.
	buf.Write(emptyHeightMap)
	buf.Write(c.biomes[:])
	d.Data2D = append([]byte(nil), buf.Bytes()...)

	buf.Reset()
	pool.Put(buf)
	return d
}

// encodeSubChunks encodes the sub chunks of the Chunk passed into the bytes.Buffer buf. It uses the function passed to
// encode the block storages and returns the resulting SerialisedData.
func encodeSubChunks(buf *bytes.Buffer, c *Chunk, f func(buf *bytes.Buffer, storage *BlockStorage)) (d SerialisedData) {
	for y, sub := range c.sub {
		if sub == nil || len(sub.storages) == 0 {
			// The sub chunk at this Y value is empty, so don't write it.
			continue
		}
		_, _ = buf.Write([]byte{SubChunkVersion, byte(len(sub.storages))})
		for _, storage := range sub.storages {
			f(buf, storage)
		}
		d.SubChunks[y] = make([]byte, buf.Len())
		_, _ = buf.Read(d.SubChunks[y])
	}
	return
}

// diskEncodeBlockStorage encodes a block storage to its network representation into the buffer passed.
func networkEncodeBlockStorage(buf *bytes.Buffer, storage *BlockStorage) {
	_ = buf.WriteByte(byte(storage.bitsPerBlock<<1) | 1)

	b := make([]byte, len(storage.blocks)*4)
	for i, v := range storage.blocks {
		// Explicitly don't use the binary package to greatly improve performance of writing the uint32s.
		b[i*4], b[i*4+1], b[i*4+2], b[i*4+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	_, _ = buf.Write(b)

	_ = protocol.WriteVarint32(buf, int32(storage.palette.Len()))
	for _, runtimeID := range storage.palette.blockRuntimeIDs {
		_ = protocol.WriteVarint32(buf, int32(runtimeID))
	}
}

// diskEncodeBlockStorage encodes a block storage to its disk representation into the buffer passed.
func diskEncodeBlockStorage(buf *bytes.Buffer, storage *BlockStorage) {
	_ = buf.WriteByte(byte(storage.bitsPerBlock << 1))
	for _, b := range storage.blocks {
		_ = binary.Write(buf, binary.LittleEndian, b)
	}
	_ = binary.Write(buf, binary.LittleEndian, int32(storage.palette.Len()))

	blocks := make([]blockEntry, storage.palette.Len())
	for index, runtimeID := range storage.palette.blockRuntimeIDs {
		// Get the block state registered with the runtime IDs we have in the palette of the block storage
		// as we need the name and data value to store.
		name, props, _ := RuntimeIDToState(runtimeID)
		blocks[index] = blockEntry{Name: name, State: props, Version: CurrentBlockVersion}
	}
	// Marshal the slice of block states into NBT and add it to the byte slice.
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, b := range blocks {
		_ = enc.Encode(b)
	}
}
