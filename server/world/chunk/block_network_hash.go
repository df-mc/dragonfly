package chunk

import "math"

// ConvertBlockNetworkHashesToRuntimeIDs converts block palette values from network hashes to registry runtime IDs.
// Unknown hashes are preserved opaquely so values that overlap registry runtime IDs survive re-encoding.
func (chunk *Chunk) ConvertBlockNetworkHashesToRuntimeIDs() {
	if chunk == nil {
		return
	}
	for _, sub := range chunk.sub {
		sub.ConvertBlockNetworkHashesToRuntimeIDs(chunk.br)
	}
}

// ConvertBlockNetworkHashesToRuntimeIDs converts block palette values from network hashes to registry runtime IDs.
// Unknown hashes are preserved opaquely so values that overlap registry runtime IDs survive re-encoding.
func (sub *SubChunk) ConvertBlockNetworkHashesToRuntimeIDs(br BlockRegistry) {
	if sub == nil || br == nil {
		return
	}
	for _, storage := range sub.storages {
		if storage == nil {
			continue
		}
		storage.palette.Replace(func(networkHash uint32) uint32 {
			if converted, ok := br.HashToRuntimeID(networkHash); ok {
				return converted
			}
			if _, collides := br.RuntimeIDToHash(networkHash); collides {
				return sub.opaqueBlockNetworkHashRuntimeID(br, networkHash)
			}
			return networkHash
		})
	}
}

func (sub *SubChunk) opaqueBlockNetworkHashRuntimeID(br BlockRegistry, networkHash uint32) uint32 {
	for runtimeID, hash := range sub.opaqueBlockNetworkHashes {
		if hash == networkHash {
			return runtimeID
		}
	}
	if sub.opaqueBlockNetworkHashes == nil {
		sub.opaqueBlockNetworkHashes = make(map[uint32]uint32)
	}
	for runtimeID := uint32(math.MaxUint32); ; runtimeID-- {
		if _, registered := br.RuntimeIDToHash(runtimeID); registered {
			continue
		}
		if _, registered := br.HashToRuntimeID(runtimeID); registered {
			continue
		}
		if _, used := sub.opaqueBlockNetworkHashes[runtimeID]; used || sub.hasPaletteValue(runtimeID) {
			continue
		}
		sub.opaqueBlockNetworkHashes[runtimeID] = networkHash
		return runtimeID
	}
}

func (sub *SubChunk) hasPaletteValue(value uint32) bool {
	for _, storage := range sub.storages {
		if storage != nil && storage.palette.Index(value) >= 0 {
			return true
		}
	}
	return false
}

// EncodeWithBlockNetworkHashes encodes c for the network with block palette runtime IDs converted to network hashes.
// The chunk is cloned before conversion, so the cached chunk remains in registry runtime-ID form. Runtime IDs without a
// registered network hash are preserved unchanged.
func EncodeWithBlockNetworkHashes(c *Chunk) SerialisedData {
	if c == nil {
		return SerialisedData{}
	}
	networkChunk := c.Clone()
	for _, sub := range networkChunk.sub {
		sub.convertRuntimeIDsToBlockNetworkHashes(c.br)
	}
	return Encode(networkChunk, NetworkEncoding)
}

// EncodeSubChunkWithBlockNetworkHashes encodes one sub-chunk with registry runtime IDs converted to network hashes.
// The source sub-chunk is cloned before conversion.
func EncodeSubChunkWithBlockNetworkHashes(c *Chunk, ind int) []byte {
	if c == nil || ind < 0 || ind >= len(c.sub) {
		return nil
	}
	networkChunk := &Chunk{
		r:   c.r,
		br:  c.br,
		sub: make([]*SubChunk, len(c.sub)),
	}
	networkChunk.sub[ind] = c.sub[ind].Clone()
	networkChunk.sub[ind].convertRuntimeIDsToBlockNetworkHashes(c.br)
	return EncodeSubChunk(networkChunk, NetworkEncoding, ind)
}

func (sub *SubChunk) convertRuntimeIDsToBlockNetworkHashes(br BlockRegistry) {
	if sub == nil || br == nil {
		return
	}
	for _, storage := range sub.storages {
		if storage == nil {
			continue
		}
		storage.palette.Replace(func(runtimeID uint32) uint32 {
			if hash, ok := sub.opaqueBlockNetworkHashes[runtimeID]; ok {
				return hash
			}
			if hash, ok := br.RuntimeIDToHash(runtimeID); ok {
				return hash
			}
			return runtimeID
		})
	}
}
