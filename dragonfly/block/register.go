package block

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/hashstructure"
)

// init registers all blocks implemented by Dragonfly.
func init() {
	Register("air", Air{})
	Register("stone", Stone{})
	Register("granite", Granite{}, Granite{Polished: true})
	Register("diorite", Diorite{}, Diorite{Polished: true})
	Register("andesite", Andesite{}, Andesite{Polished: true})
	Register("grass", Grass{})
	Register("dirt", Dirt{}, Dirt{Coarse: true})
	Register("log", allLogs()...)
	Register("bedrock", Bedrock{}, Bedrock{InfiniteBurning: true})

	registerAllStates()
}

var blocks = map[string]Block{}
var saveNames = map[Block]string{}

var registeredStates []Block
var runtimeIDs = map[Block]uint32{}

var existingStates = map[string]struct{}{}

// Register registers a block with the save name passed. The save name is used to save the block to the
// world's database and must not conflict with existing blocks.
// If a saveName is passed which already has a block registered for it, Register panics.
func Register(saveName string, states ...Block) {
	if len(states) == 0 {
		panic("at least one block state must be registered")
	}
	if _, ok := blocks[saveName]; ok {
		panic("cannot overwrite an existing block with the save name " + saveName)
	}
	blocks[saveName] = states[0]

	for _, state := range states {
		runtimeIDs[state] = uint32(len(registeredStates))
		saveNames[state] = saveName
		registeredStates = append(registeredStates, state)

		name, props := state.Minecraft()
		h, _ := hashstructure.Hash(props, nil)
		existingStates[fmt.Sprint(name, h)] = struct{}{}
	}
}

// Get attempts to return a block by a save name registered using Register. If found, the block is returned
// and the bool is true. If not found, false is returned and the block is nil.
func Get(saveName string) (Block, bool) {
	b, ok := blocks[saveName]
	return b, ok
}

// SaveName attempts to return the save name of a particular block state. If found, the string is non-empty
// and the bool returned true. The bool is false if the state was not registered.
func SaveName(state Block) (string, bool) {
	saveName, ok := saveNames[state]
	return saveName, ok
}

// RuntimeID attempts to return a runtime ID of a block state previously registered using Register(). If the
// runtime ID is found, the bool returned is true. It is otherwise false.
func RuntimeID(state Block) (uint32, bool) {
	runtimeID, ok := runtimeIDs[state]
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
	_ = json.Unmarshal([]byte(allStates), &m)

	for _, b := range m {
		h, _ := hashstructure.Hash(b.Block.Properties, nil)
		key := fmt.Sprint(b.Block.Name, h)
		if _, ok := existingStates[key]; ok {
			// Duplicate state, don't add it.
			continue
		}
		registeredStates = append(registeredStates, b)
	}
}

// unimplementedBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unimplementedBlock struct {
	Block struct {
		Name       string                 `json:"name"`
		Properties map[string]interface{} `json:"states"`
	} `json:"block"`
}

func (u unimplementedBlock) Name() string {
	return u.Block.Name
}

func (u unimplementedBlock) Minecraft() (name string, properties map[string]interface{}) {
	return u.Block.Name, u.Block.Properties
}
