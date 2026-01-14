package world

import (
	"bytes"
	_ "embed"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"unsafe"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

var (
	//go:embed block_states.nbt
	blockStateData []byte
)

func init() {
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))

	// Register all block states present in the block_states.nbt file. These are all possible options registered
	// blocks may encode to.
	for {
		var s BlockState
		if err := dec.Decode(&s); err != nil {
			break
		}
		DefaultBlockRegistry.RegisterBlockState(s)
	}
}

// BlockState holds a combination of a name and properties, together with a version.
type BlockState struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

// unknownBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unknownBlock struct {
	BlockState
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

// stateHash is a struct that may be used as a map key for block states. It contains the name of the block state
// and an encoded version of the properties.
type stateHash struct {
	name, properties string
}

// hashProperties produces a hash for the block properties held by the BlockState.
func hashProperties(properties map[string]any) string {
	if properties == nil {
		return ""
	}
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	slices.Sort(keys)

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
