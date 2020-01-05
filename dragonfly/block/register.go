package block

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"sort"
	"unsafe"
)

// init registers all blocks implemented by Dragonfly.
func init() {
	Register(Air{})
	Register(Stone{})
	Register(Granite{}, Granite{Polished: true})
	Register(Diorite{}, Diorite{Polished: true})
	Register(Andesite{}, Andesite{Polished: true})
	Register(Grass{})
	Register(Dirt{}, Dirt{Coarse: true})
	Register(allLogs()...)
	Register(Bedrock{}, Bedrock{InfiniteBurning: true})

	registerAllStates()
}

var registeredStates []Block
var runtimeIDs = map[string]uint32{}
var blocksHash = map[string]Block{}

// Register registers a block with the save name passed. The save name is used to save the block to the
// world's database and must not conflict with existing blocks.
// If a saveName is passed which already has a block registered for it, Register panics.
func Register(states ...Block) {
	if len(states) == 0 {
		panic("at least one block state must be registered")
	}
	for _, state := range states {
		name, props := state.Minecraft()
		key := name + HashProperties(props)

		if _, ok := blocksHash[key]; ok {
			panic(fmt.Sprintf("cannot overwrite an existing block with the same name '%v' and properties %+v", name, props))
		}

		runtimeIDs[key] = uint32(len(registeredStates))
		registeredStates = append(registeredStates, state)

		blocksHash[key] = state
	}
}

// Get attempts to return a block by its Minecraft save name combined with the hash of its block properties.
// If found, the Block returned is non-nil and the bool true.
func Get(key string) (Block, bool) {
	b, ok := blocksHash[key]
	return b, ok
}

// RuntimeID attempts to return a runtime ID of a block state previously registered using Register(). If the
// runtime ID is found, the bool returned is true. It is otherwise false.
func RuntimeID(state Block) (uint32, bool) {
	name, props := state.Minecraft()
	runtimeID, ok := runtimeIDs[name+HashProperties(props)]
	return runtimeID, ok
}

// ByRuntimeID attempts to return a block state by its runtime ID. If not found, the bool returned is false.
// If found, the block state is non-nil and the bool true.
func ByRuntimeID(runtimeID uint32) (Block, bool) {
	if runtimeID >= uint32(len(registeredStates)) {
		return nil, false
	}
	return registeredStates[runtimeID], true
}

// All returns a list of all registered states of the server. The list is ordered according to the runtime ID
// that the blocks have.
func All() []Block {
	return registeredStates
}

// registerAllStates registers all block states present in the game, skipping ones that have already been
// registered before this is called.
func registerAllStates() {
	var m []unimplementedBlock
	b, _ := base64.StdEncoding.DecodeString(allStates)
	_ = nbt.Unmarshal(b, &m)

	for _, b := range m {
		key := b.Block.Name + HashProperties(b.Block.Properties)
		if _, ok := blocksHash[key]; ok {
			// Duplicate state, don't add it.
			continue
		}
		Register(b)
	}
}

// HashProperties produces a hash for the block properties map passed.
// Passing the same map into HashProperties will always result in the same hash.
func HashProperties(properties map[string]interface{}) string {
	l := len(properties)

	keys := make([]string, 0, l)
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]interface{}, 0, l)
	for _, k := range keys {
		values = append(values, properties[k])
	}

	a, _ := nbt.Marshal(keys)
	b, _ := nbt.Marshal(values)

	h := sha256.New()
	h.Write(a)
	v := h.Sum(b)

	return *(*string)(unsafe.Pointer(&v))
}

// unimplementedBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unimplementedBlock struct {
	Block struct {
		Name       string                 `nbt:"name"`
		Properties map[string]interface{} `nbt:"states"`
		Version    int32                  `nbt:"version"`
	} `nbt:"block"`
	ID int16 `nbt:"id"`
}

func (u unimplementedBlock) Name() string {
	return u.Block.Name
}

func (u unimplementedBlock) Minecraft() (name string, properties map[string]interface{}) {
	return u.Block.Name, u.Block.Properties
}
