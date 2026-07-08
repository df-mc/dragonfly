package chunk

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// EncodeLevelChunkPayload builds the RawPayload for a LevelChunk packet from
// network-serialised chunk data and trailing block entity compounds.
func EncodeLevelChunkPayload(data SerialisedData, blockEntities []BlockEntity) ([]byte, error) {
	blockNBTs := make([]map[string]any, 0, len(blockEntities))
	for _, blockEntity := range blockEntities {
		if blockEntity.Data == nil {
			continue
		}
		blockNBT := make(map[string]any, len(blockEntity.Data)+3)
		for k, v := range blockEntity.Data {
			blockNBT[k] = v
		}
		blockNBT["x"] = int32(blockEntity.Pos[0])
		blockNBT["y"] = int32(blockEntity.Pos[1])
		blockNBT["z"] = int32(blockEntity.Pos[2])
		blockNBTs = append(blockNBTs, blockNBT)
	}
	return EncodeLevelChunkPayloadWithBlockNBTs(data, blockNBTs)
}

// EncodeLevelChunkPayloadFromMap builds the RawPayload for a LevelChunk packet
// from network-serialised chunk data and raw block entity NBT indexed by
// absolute block position.
func EncodeLevelChunkPayloadFromMap(data SerialisedData, blockEntities map[cube.Pos]map[string]any) ([]byte, error) {
	entries := make([]BlockEntity, 0, len(blockEntities))
	for pos, blockNBT := range blockEntities {
		entries = append(entries, BlockEntity{Pos: pos, Data: blockNBT})
	}
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i].Pos, entries[j].Pos
		if a[1] != b[1] {
			return a[1] < b[1]
		}
		if a[2] != b[2] {
			return a[2] < b[2]
		}
		return a[0] < b[0]
	})
	return EncodeLevelChunkPayload(data, entries)
}

// EncodeLevelChunkPayloadWithBlockNBTs builds the RawPayload for a LevelChunk
// packet from network-serialised chunk data and block entity NBT compounds that
// already include their id/x/y/z fields.
func EncodeLevelChunkPayloadWithBlockNBTs(data SerialisedData, blockNBTs []map[string]any) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for _, sub := range data.SubChunks {
		_, _ = buf.Write(sub)
	}
	_, _ = buf.Write(data.Biomes)

	// Length of 1 byte for the border block count.
	buf.WriteByte(0)

	staging := bytes.NewBuffer(nil)
	for i, blockNBT := range blockNBTs {
		if blockNBT == nil {
			continue
		}
		staging.Reset()
		if err := nbt.NewEncoderWithEncoding(staging, nbt.NetworkLittleEndian).Encode(blockNBT); err != nil {
			return nil, fmt.Errorf("encode block entity %d at %v/%v/%v: %w", i, blockNBT["x"], blockNBT["y"], blockNBT["z"], err)
		}
		_, _ = buf.Write(staging.Bytes())
	}
	return append([]byte(nil), buf.Bytes()...), nil
}
