package session

import (
	"bytes"
	"github.com/cespare/xxhash"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// subChunkRequests is set to true to enable the sub-chunk request system. This can (likely) cause unexpected issues,
// but also solves issues with block entities such as item frames and lecterns as of v1.19.10.
const subChunkRequests = true

// ViewChunk ...
func (s *Session) ViewChunk(pos world.ChunkPos, c *chunk.Chunk, blockEntities map[cube.Pos]world.Block) {
	if !s.conn.ClientCacheEnabled() {
		s.sendNetworkChunk(pos, c, blockEntities)
		return
	}
	s.sendBlobHashes(pos, c, blockEntities)
}

// ViewSubChunks ...
func (s *Session) ViewSubChunks(center world.SubChunkPos, offsets []protocol.SubChunkOffset) {
	w := s.c.World()
	r := w.Range()

	entries := make([]protocol.SubChunkEntry, 0, len(offsets))
	transaction := make(map[uint64]struct{})
	for _, offset := range offsets {
		ch, ok := s.chunkLoader.Chunk(world.ChunkPos{
			center.X() + int32(offset[0]),
			center.Z() + int32(offset[2]),
		})
		if !ok {
			entries = append(entries, protocol.SubChunkEntry{Result: protocol.SubChunkResultChunkNotFound, Offset: offset})
			continue
		}

		ind := int16(center.Y()) + int16(offset[1]) - int16(r[0])>>4
		if ind < 0 || ind >= int16(len(ch.Sub())) {
			entries = append(entries, protocol.SubChunkEntry{Result: protocol.SubChunkResultIndexOutOfBounds, Offset: offset})
			continue
		}

		chunkMap := ch.HeightMap()
		subMapType, subMap := byte(protocol.HeightMapDataHasData), make([]int8, 256)
		higher, lower := true, true
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				y, i := chunkMap.At(x, z), (uint16(z)<<4)|uint16(x)
				otherInd := ch.SubIndex(y)
				if otherInd > ind {
					subMap[i], lower = 16, false
				} else if otherInd < ind {
					subMap[i], higher = -1, false
				} else {
					subMap[i], lower, higher = int8(y-ch.SubY(otherInd)), false, false
				}
			}
		}
		if higher {
			subMapType, subMap = protocol.HeightMapDataTooHigh, nil
		} else if lower {
			subMapType, subMap = protocol.HeightMapDataTooLow, nil
		}

		sub := ch.Sub()[ind]
		if sub.Empty() {
			entries = append(entries, protocol.SubChunkEntry{
				Result:        protocol.SubChunkResultSuccessAllAir,
				HeightMapType: subMapType,
				HeightMapData: subMap,
				Offset:        offset,
			})
			continue
		}

		serialisedSubChunk := chunk.EncodeSubChunk(ch.Chunk, chunk.NetworkEncoding, int(ind))
		blockEntityBuf := bytes.NewBuffer(nil)
		enc := nbt.NewEncoderWithEncoding(blockEntityBuf, nbt.NetworkLittleEndian)
		for pos, b := range ch.BlockEntities() {
			if n, ok := b.(world.NBTer); ok && ch.SubIndex(int16(pos.Y())) == ind {
				d := n.EncodeNBT()
				d["x"], d["y"], d["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
				_ = enc.Encode(d)
			}
		}

		entry := protocol.SubChunkEntry{
			Result:        protocol.SubChunkResultSuccess,
			RawPayload:    append(serialisedSubChunk, blockEntityBuf.Bytes()...),
			HeightMapType: subMapType,
			HeightMapData: subMap,
			Offset:        offset,
		}
		if s.conn.ClientCacheEnabled() {
			if hash := xxhash.Sum64(serialisedSubChunk); s.trackBlob(hash, serialisedSubChunk) {
				transaction[hash] = struct{}{}

				entry.BlobHash = hash
				entry.RawPayload = blockEntityBuf.Bytes()
			}
		}
		entries = append(entries, entry)
	}
	if s.conn.ClientCacheEnabled() && len(transaction) > 0 {
		s.blobMu.Lock()
		s.openChunkTransactions = append(s.openChunkTransactions, transaction)
		s.blobMu.Unlock()
	}
	s.writePacket(&packet.SubChunk{
		Dimension:       int32(w.Dimension().EncodeDimension()),
		Position:        protocol.SubChunkPos(center),
		CacheEnabled:    s.conn.ClientCacheEnabled(),
		SubChunkEntries: entries,
	})
}

// sendBlobHashes sends chunk blob hashes of the data of the chunk and stores the data in a map of blobs. Only
// data that the client doesn't yet have will be sent over the network.
func (s *Session) sendBlobHashes(pos world.ChunkPos, c *chunk.Chunk, blockEntities map[cube.Pos]world.Block) {
	if subChunkRequests {
		biomes := chunk.EncodeBiomes(c, chunk.NetworkEncoding)
		if hash := xxhash.Sum64(biomes); s.trackBlob(hash, biomes) {
			s.writePacket(&packet.LevelChunk{
				SubChunkRequestMode: protocol.SubChunkRequestModeLimited,
				Position:            protocol.ChunkPos(pos),
				SubChunkCount:       uint32(len(c.Sub())),
				HighestSubChunk:     uint16(len(c.Sub())), // This is always going to be the highest sub-chunk, anyway.
				BlobHashes:          []uint64{hash},
				RawPayload:          []byte{0},
				CacheEnabled:        true,
			})
			return
		}
	}

	var (
		data   = chunk.Encode(c, chunk.NetworkEncoding)
		count  = uint32(len(data.SubChunks))
		blobs  = append(data.SubChunks, data.Biomes)
		hashes = make([]uint64, len(blobs))
		m      = make(map[uint64]struct{}, len(blobs))
	)
	for i, blob := range blobs {
		h := xxhash.Sum64(blob)
		hashes[i], m[h] = h, struct{}{}
	}

	s.blobMu.Lock()
	s.openChunkTransactions = append(s.openChunkTransactions, m)
	if l := len(s.blobs); l > 4096 {
		s.blobMu.Unlock()
		s.log.Errorf("player %v has too many blobs pending %v: disconnecting", s.c.Name(), l)
		_ = s.c.Close()
		return
	}
	for i := range hashes {
		s.blobs[hashes[i]] = blobs[i]
	}
	s.blobMu.Unlock()

	// Length of 1 byte for the border block count.
	raw := bytes.NewBuffer(make([]byte, 1, 32))
	enc := nbt.NewEncoderWithEncoding(raw, nbt.NetworkLittleEndian)
	for bp, b := range blockEntities {
		if n, ok := b.(world.NBTer); ok {
			d := n.EncodeNBT()
			d["x"], d["y"], d["z"] = int32(bp[0]), int32(bp[1]), int32(bp[2])
			_ = enc.Encode(d)
		}
	}

	s.writePacket(&packet.LevelChunk{
		Position:      protocol.ChunkPos{pos.X(), pos.Z()},
		SubChunkCount: count,
		CacheEnabled:  true,
		BlobHashes:    hashes,
		RawPayload:    raw.Bytes(),
	})
}

// sendNetworkChunk sends a network encoded chunk to the client.
func (s *Session) sendNetworkChunk(pos world.ChunkPos, c *chunk.Chunk, blockEntities map[cube.Pos]world.Block) {
	if subChunkRequests {
		s.writePacket(&packet.LevelChunk{
			SubChunkRequestMode: protocol.SubChunkRequestModeLimited,
			Position:            protocol.ChunkPos(pos),
			SubChunkCount:       uint32(len(c.Sub())),
			HighestSubChunk:     uint16(len(c.Sub())), // This is always going to be the highest sub-chunk, anyway.
			RawPayload:          append(chunk.EncodeBiomes(c, chunk.NetworkEncoding), 0),
		})
		return
	}

	data := chunk.Encode(c, chunk.NetworkEncoding)
	chunkBuf := bytes.NewBuffer(nil)
	for _, s := range data.SubChunks {
		_, _ = chunkBuf.Write(s)
	}
	_, _ = chunkBuf.Write(data.Biomes)

	// Length of 1 byte for the border block count.
	chunkBuf.WriteByte(0)

	enc := nbt.NewEncoderWithEncoding(chunkBuf, nbt.NetworkLittleEndian)
	for bp, b := range blockEntities {
		if n, ok := b.(world.NBTer); ok {
			d := n.EncodeNBT()
			d["x"], d["y"], d["z"] = int32(bp[0]), int32(bp[1]), int32(bp[2])
			_ = enc.Encode(d)
		}
	}

	s.writePacket(&packet.LevelChunk{
		Position:      protocol.ChunkPos{pos.X(), pos.Z()},
		SubChunkCount: uint32(len(data.SubChunks)),
		RawPayload:    append([]byte(nil), chunkBuf.Bytes()...),
	})
}

// trackBlob attempts to track the given blob. If the player has too many pending blobs, it returns false and closes the
// connection.
func (s *Session) trackBlob(hash uint64, blob []byte) bool {
	s.blobMu.Lock()
	defer s.blobMu.Unlock()

	if l := len(s.blobs); l > 4096 {
		s.blobMu.Unlock()
		s.log.Errorf("player %v has too many blobs pending %v: disconnecting", s.c.Name(), l)
		_ = s.c.Close()
		return false
	}
	s.blobs[hash] = blob
	return true
}
