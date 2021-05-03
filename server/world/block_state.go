package world

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/df-mc/dragonfly/server/internal/world_internal"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
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
	// stateRuntimeIDs holds a map for looking up the runtime ID of a block by the stateHash it produces.
	stateRuntimeIDs = map[stateHash]uint32{}
	// nbtBlocks holds a list of NBTer implementations for blocks registered that implement the NBTer interface.
	// These are indexed by their runtime IDs. Blocks that do not implement NBTer have a nil implementation in
	// this slice.
	nbtBlocks []bool
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

	chunk.RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]interface{}, found bool) {
		if runtimeID >= uint32(len(blocks)) {
			return "", nil, false
		}
		name, properties = blocks[runtimeID].EncodeBlock()
		return name, properties, true
	}
	chunk.StateToRuntimeID = func(name string, properties map[string]interface{}) (runtimeID uint32, found bool) {
		rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
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
	rid := uint32(len(blocks))
	if s.Name == "minecraft:air" {
		world_internal.AirRuntimeID = rid
	}
	stateRuntimeIDs[h] = rid
	blocks = append(blocks, unknownBlock{s})

	nbtBlocks = append(nbtBlocks, false)
	world_internal.LiquidRemovable = append(world_internal.LiquidRemovable, false)
	chunk.FilteringBlocks = append(chunk.FilteringBlocks, 15)
	chunk.LightBlocks = append(chunk.LightBlocks, 0)
	world_internal.BeaconSource = append(world_internal.BeaconSource, false)
}

// unknownBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unknownBlock struct {
	blockState
}

// EncodeBlock ...
func (b unknownBlock) EncodeBlock() (string, map[string]interface{}) {
	return b.Name, b.Properties
}

// Model ...
func (unknownBlock) Model() BlockModel {
	return unknownModel{}
}

// blockState holds a combination of a name and properties, together with a version.
type blockState struct {
	Name       string                 `nbt:"name"`
	Properties map[string]interface{} `nbt:"states"`
	Version    int32                  `nbt:"version"`
}

// stateHash is a struct that may be used as a map key for block states. It contains the name of the block state
// and an encoded version of the properties.
type stateHash struct {
	name, properties string
}

// HashProperties produces a hash for the block properties held by the blockState.
func hashProperties(properties map[string]interface{}) string {
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
