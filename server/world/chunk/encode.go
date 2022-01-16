package chunk

import (
	"bytes"
	"sync"
)

const (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written on disk and over network.
	SubChunkVersion = 9
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	CurrentBlockVersion int32 = 17825806
)

var (
	// RuntimeIDToState must hold a function to convert a runtime ID to a name and its state properties.
	RuntimeIDToState func(runtimeID uint32) (name string, properties map[string]interface{}, found bool)
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
		SubChunks [][]byte
		// Biomes is the biome data of the chunk, which is composed of a biome storage for each subchunk.
		Biomes []byte
		// BlockNBT is an encoded NBT array of all blocks that carry additional NBT, such as chests, with all
		// their contents.
		BlockNBT []byte
	}
	// blockEntry represents a block as found in a disk save of a world.
	blockEntry struct {
		Name    string                 `nbt:"name"`
		State   map[string]interface{} `nbt:"states"`
		Version int32                  `nbt:"version"`
		ID      int32                  `nbt:"oldid,omitempty"`
		Meta    int16                  `nbt:"val,omitempty"`
	}
)

// Encode encodes Chunk to an intermediate representation SerialisedData. An Encoding may be passed to encode either for
// network or disk purposed, the most notable difference being that the network encoding generally uses varints and no
// NBT.
func Encode(c *Chunk, e Encoding) SerialisedData {
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	d := encodeSubChunks(buf, c, e)
	for i := range c.biomes {
		encodePalettedStorage(buf, c.biomes[i], e, BiomePaletteEncoding)
	}
	d.Biomes = append([]byte(nil), buf.Bytes()...)

	return d
}

// encodeSubChunks encodes the sub chunks of the Chunk passed into the bytes.Buffer buf. It uses the encoding passed to
// encode the block storages and returns the resulting SerialisedData.
func encodeSubChunks(buf *bytes.Buffer, c *Chunk, e Encoding) (d SerialisedData) {
	d.SubChunks = make([][]byte, len(c.sub))
	for i, sub := range c.sub {
		_, _ = buf.Write([]byte{SubChunkVersion, byte(len(sub.storages)), uint8(i + (c.r[0] >> 4))})
		for _, storage := range sub.storages {
			encodePalettedStorage(buf, storage, e, BlockPaletteEncoding)
		}
		d.SubChunks[i] = make([]byte, buf.Len())
		_, _ = buf.Read(d.SubChunks[i])
	}
	return
}

// encodePalettedStorage encodes a PalettedStorage into a bytes.Buffer. The Encoding passed is used to write the Palette
// of the PalettedStorage.
func encodePalettedStorage(buf *bytes.Buffer, storage *PalettedStorage, e Encoding, pe paletteEncoding) {
	b := make([]byte, len(storage.indices)*4+1)
	b[0] = byte(storage.bitsPerIndex<<1) | e.network()

	for i, v := range storage.indices {
		// Explicitly don't use the binary package to greatly improve performance of writing the uint32s.
		b[i*4+1], b[i*4+2], b[i*4+3], b[i*4+4] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	_, _ = buf.Write(b)

	e.encodePalette(buf, storage.palette, pe)
}
