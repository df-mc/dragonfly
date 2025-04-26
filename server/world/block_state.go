package world

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"maps"
	"math"
	"slices"
	"sort"
	"strings"
	"unsafe"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte

	blockProperties = map[string]map[string]any{}
	// blocks holds a list of all registered Blocks indexed by their runtime ID. Blocks that were not explicitly
	// registered are of the type unknownBlock.
	blocks []Block
	// customBlocks maps a custom block's identifier to a slice of custom blocks.
	customBlocks = map[string]CustomBlock{}
	// stateRuntimeIDs holds a map for looking up the runtime ID of a block by the stateHash it produces.
	stateRuntimeIDs = map[stateHash]uint32{}
	// nbtBlocks holds a list of NBTer implementations for blocks registered that implement the NBTer interface.
	// These are indexed by their runtime IDs. Blocks that do not implement NBTer have a false value in this slice.
	nbtBlocks []bool
	// randomTickBlocks holds a list of RandomTicker implementations for blocks registered that implement the RandomTicker interface.
	// These are indexed by their runtime IDs. Blocks that do not implement RandomTicker have a false value in this slice.
	randomTickBlocks []bool
	// liquidBlocks holds a list of Liquid implementations for blocks registered that implement the Liquid interface.
	// These are indexed by their runtime IDs. Blocks that do not implement Liquid have a false value in this slice.
	liquidBlocks []bool
	// liquidDisplacingBlocks holds a list of LiquidDisplacer implementations for blocks registered that implement the LiquidDisplacer interface.
	// These are indexed by their runtime IDs. Blocks that do not implement LiquidDisplacer have a false value in this slice.
	liquidDisplacingBlocks []bool
	// airRID is the runtime ID of an air block.
	airRID uint32
)

func init() {
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))

	// Register all block states present in the block_states.nbt file. These are all possible options registered
	// blocks may encode to.
	var s blockState
	for {
		if err := dec.Decode(&s); err != nil {
			break
		}
		registerBlockState(s)
	}

	chunk.RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]any, found bool) {
		if runtimeID >= uint32(len(blocks)) {
			return "", nil, false
		}
		name, properties = blocks[runtimeID].EncodeBlock()
		return name, properties, true
	}
	chunk.StateToRuntimeID = func(name string, properties map[string]any) (runtimeID uint32, found bool) {
		if rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]; ok {
			return rid, true
		}
		rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(blockProperties[name])}]
		return rid, ok
	}
}

// registerBlockState registers a new blockState to the states slice. The function panics if the properties the
// blockState hold are invalid or if the blockState was already registered.
func registerBlockState(s blockState) {
	h := stateHash{name: s.Name, properties: hashProperties(s.Properties)}
	if _, ok := stateRuntimeIDs[h]; ok {
		panic(fmt.Sprintf("cannot register the same state twice (%+v)", s))
	}
	if _, ok := blockProperties[s.Name]; !ok {
		blockProperties[s.Name] = s.Properties
	}
	rid := uint32(len(blocks))
	blocks = append(blocks, unknownBlock{blockState: s})

	if s.Name == "minecraft:air" {
		airRID = rid
	}

	nbtBlocks = slices.Insert(nbtBlocks, int(rid), false)
	randomTickBlocks = slices.Insert(randomTickBlocks, int(rid), false)
	liquidBlocks = slices.Insert(liquidBlocks, int(rid), false)
	liquidDisplacingBlocks = slices.Insert(liquidDisplacingBlocks, int(rid), false)
	chunk.FilteringBlocks = slices.Insert(chunk.FilteringBlocks, int(rid), 15)
	chunk.LightBlocks = slices.Insert(chunk.LightBlocks, int(rid), 0)
	stateRuntimeIDs[h] = rid
}

// unknownBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unknownBlock struct {
	blockState
	data map[string]any
}

// EncodeBlock ...
func (b unknownBlock) EncodeBlock() (string, map[string]any) {
	return b.Name, b.Properties
}

// Model ...
func (unknownBlock) Model() BlockModel {
	return unknownModel{}
}

// Hash ...
func (b unknownBlock) Hash() (uint64, uint64) {
	return 0, math.MaxUint64
}

// EncodeNBT ...
func (b unknownBlock) EncodeNBT() map[string]any {
	return maps.Clone(b.data)
}

// DecodeNBT ...
func (b unknownBlock) DecodeNBT(data map[string]any) any {
	b.data = maps.Clone(data)
	return b
}

// blockState holds a combination of a name and properties, together with a version.
type blockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

// stateHash is a struct that may be used as a map key for block states. It contains the name of the block state
// and an encoded version of the properties.
type stateHash struct {
	name, properties string
}

// hashProperties produces a hash for the block properties held by the blockState.
func hashProperties(properties map[string]any) string {
	if properties == nil {
		return ""
	}
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var b strings.Builder
	for _, k := range keys {
		switch v := properties[k].(type) {
		case bool:
			if v {
				b.WriteByte(1)
			} else {
				b.WriteByte(0)
			}
		case uint8:
			b.WriteByte(v)
		case int32:
			a := *(*[4]byte)(unsafe.Pointer(&v))
			b.Write(a[:])
		case string:
			b.WriteString(v)
		default:
			// If block encoding is broken, we want to find out as soon as possible. This saves a lot of time
			// debugging in-game.
			panic(fmt.Sprintf("invalid block property type %T for property %v", v, k))
		}
	}

	return b.String()
}
