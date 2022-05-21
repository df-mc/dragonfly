package world

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"unsafe"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte
	// blocks holds a list of all registered Blocks indexed by their runtime ID. Blocks that were not explicitly
	// registered are of the type unknownBlock.
	blocks []Block
	// customBlocks ...
	customBlocks []CustomBlock
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
		registerBlockState(s, false)
	}

	chunk.RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]any, found bool) {
		if runtimeID >= uint32(len(blocks)) {
			return "", nil, false
		}
		name, properties = blocks[runtimeID].EncodeBlock()
		return name, properties, true
	}
	chunk.StateToRuntimeID = func(name string, properties map[string]any) (runtimeID uint32, found bool) {
		rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
		return rid, ok
	}
}

// registerBlockState registers a new blockState to the states slice. The function panics if the properties the
// blockState hold are invalid or if the blockState was already registered.
func registerBlockState(s blockState, order bool) {
	h := stateHash{name: s.Name, properties: hashProperties(s.Properties)}
	if _, ok := stateRuntimeIDs[h]; ok {
		panic(fmt.Sprintf("cannot register the same state twice (%+v)", s))
	}

	blocks = append(blocks, unknownBlock{s})
	if order {
		sort.SliceStable(blocks, func(i, j int) bool {
			one, _ := blocks[i].EncodeBlock()
			two, _ := blocks[j].EncodeBlock()

			f := fnv.New64()
			_, _ = f.Write([]byte(one))
			hashOne := f.Sum64()

			f = fnv.New64()
			_, _ = f.Write([]byte(two))
			hashTwo := f.Sum64()

			return hashOne < hashTwo
		})
	}

	nbtBlocks = append(nbtBlocks, false)
	randomTickBlocks = append(randomTickBlocks, false)
	liquidBlocks = append(liquidBlocks, false)
	liquidDisplacingBlocks = append(liquidDisplacingBlocks, false)
	chunk.FilteringBlocks = append(chunk.FilteringBlocks, 15)
	chunk.LightBlocks = append(chunk.LightBlocks, 0)

	if !order {
		rid := uint32(len(blocks)) - 1
		if s.Name == "minecraft:air" {
			airRID = rid
		}
		stateRuntimeIDs[h] = rid
		return
	}

	updatedNBTBlocks := make([]bool, len(nbtBlocks))
	updatedRandomTickBlocks := make([]bool, len(randomTickBlocks))
	updatedLiquidBlocks := make([]bool, len(liquidBlocks))
	updatedLiquidDisplacingBlocks := make([]bool, len(liquidDisplacingBlocks))
	updatedFilteringBlocks := make([]uint8, len(chunk.FilteringBlocks))
	updatedLightBlocks := make([]uint8, len(chunk.LightBlocks))

	newStateRuntimeIDs := make(map[stateHash]uint32, len(stateRuntimeIDs))
	for id, b := range blocks {
		name, properties := b.EncodeBlock()
		i := stateHash{name: name, properties: hashProperties(properties)}
		if name == "minecraft:air" {
			airRID = uint32(id)
		}

		if oldID, ok := stateRuntimeIDs[i]; ok {
			updatedNBTBlocks[id] = nbtBlocks[oldID]
			updatedRandomTickBlocks[id] = randomTickBlocks[oldID]
			updatedLiquidBlocks[id] = liquidBlocks[oldID]
			updatedLiquidDisplacingBlocks[id] = liquidDisplacingBlocks[oldID]
			updatedFilteringBlocks[id] = chunk.FilteringBlocks[oldID]
			updatedLightBlocks[id] = chunk.LightBlocks[oldID]
		}
		newStateRuntimeIDs[i] = uint32(id)
	}

	stateRuntimeIDs = newStateRuntimeIDs
	nbtBlocks, randomTickBlocks = updatedNBTBlocks, updatedRandomTickBlocks
	liquidBlocks, liquidDisplacingBlocks = updatedLiquidBlocks, updatedLiquidDisplacingBlocks
	chunk.FilteringBlocks, chunk.LightBlocks = updatedFilteringBlocks, updatedLightBlocks
}

// unknownBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unknownBlock struct {
	blockState
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
func (b unknownBlock) Hash() uint64 {
	return math.MaxUint64
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

// HashProperties produces a hash for the block properties held by the blockState.
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
