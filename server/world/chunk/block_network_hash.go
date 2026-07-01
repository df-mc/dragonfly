package chunk

// ConvertBlockNetworkHashesToRuntimeIDs converts block palette values from network hashes to registry runtime IDs.
// Unknown hashes are preserved unchanged.
func (chunk *Chunk) ConvertBlockNetworkHashesToRuntimeIDs() {
	if chunk == nil {
		return
	}
	for _, sub := range chunk.sub {
		sub.ConvertBlockNetworkHashesToRuntimeIDs(chunk.br)
	}
}

// ConvertBlockNetworkHashesToRuntimeIDs converts block palette values from network hashes to registry runtime IDs.
// Unknown hashes are preserved unchanged.
func (sub *SubChunk) ConvertBlockNetworkHashesToRuntimeIDs(br BlockRegistry) {
	if sub == nil || br == nil {
		return
	}
	for _, storage := range sub.storages {
		if storage == nil {
			continue
		}
		storage.palette.Replace(func(runtimeID uint32) uint32 {
			if converted, ok := br.HashToRuntimeID(runtimeID); ok {
				return converted
			}
			return runtimeID
		})
	}
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

func (sub *SubChunk) convertRuntimeIDsToBlockNetworkHashes(br BlockRegistry) {
	if sub == nil || br == nil {
		return
	}
	for _, storage := range sub.storages {
		if storage == nil {
			continue
		}
		storage.palette.Replace(func(runtimeID uint32) uint32 {
			if hash, ok := br.RuntimeIDToHash(runtimeID); ok {
				return hash
			}
			return runtimeID
		})
	}
}
