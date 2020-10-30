package world

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/brentp/intintmap"
	"github.com/cespare/xxhash"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/internal/resource"
	"github.com/df-mc/dragonfly/dragonfly/internal/world_internal"
	"github.com/df-mc/dragonfly/dragonfly/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/yourbasic/radix"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unsafe"
)

// Block is a block that may be placed or found in a world. In addition, the block may also be added to an
// inventory: It is also an item.
type Block interface {
	// EncodeBlock converts the block to its encoded representation: It returns the name of the minecraft
	// block, for example 'minecraft:stone' and the block properties (also referred to as states) that the
	// block holds.
	EncodeBlock() (name string, properties map[string]interface{})
	// Hash returns a unique hash of the block including the block properties. No two different blocks must
	// return the same hash.
	Hash() uint64
	// HasNBT specifies if this Block has additional NBT present in the world save, also known as a block
	// entity. If true is returned, Block must implemented the NBTer interface.
	HasNBT() bool
	// Model returns the BlockModel of the Block.
	Model() BlockModel
}

// RandomTicker represents a block that executes an action when it is ticked randomly. Every 20th of a second,
// one random block in each sub chunk are picked to receive a random tick.
type RandomTicker interface {
	// RandomTick handles a random tick of the block at the position passed. Additionally, a rand.Rand
	// instance is passed which may be used to generate values randomly without locking.
	RandomTick(pos BlockPos, w *World, r *rand.Rand)
}

// ScheduledTicker represents a block that executes an action when it has a block update scheduled, such as
// when a block adjacent to it is broken.
type ScheduledTicker interface {
	// ScheduledTick handles a scheduled tick initiated by an event in one of the neighbouring blocks, such as
	// when a block is placed or broken.
	ScheduledTick(pos BlockPos, w *World)
}

// NeighbourUpdateTicker represents a block that is updated when a block adjacent to it is updated, either
// through placement or being broken.
type NeighbourUpdateTicker interface {
	// NeighbourUpdateTick handles a neighbouring block being updated. The position of that block and the
	// position of this block is passed.
	NeighbourUpdateTick(pos, changedNeighbour BlockPos, w *World)
}

// lightEmitter is identical to a block.lightEmitter.
type lightEmitter interface {
	LightEmissionLevel() uint8
}

// lightDiffuser is identical to a block.LightDiffuser.
type lightDiffuser interface {
	LightDiffusionLevel() uint8
}

// liquidRemovable is identical to a block.LiquidRemovable.
type liquidRemovable interface {
	HasLiquidDrops() bool
}

// beaconSource represents a block which is capable of contributing to powering a beacon pyramid.
type beaconSource interface {
	// PowersBeacon returns a bool which indicates whether this block can contribute to powering up a
	// beacon pyramid.
	PowersBeacon() bool
}

// LiquidDisplacer represents a block that is able to displace a liquid to a different world layer, without
// fully removing the liquid.
type LiquidDisplacer interface {
	// CanDisplace specifies if the block is able to displace the liquid passed.
	CanDisplace(b Liquid) bool
	// SideClosed checks if a position on the side of the block placed in the world at a specific position is
	// closed. When this returns true (for example, when the side is below the position and the block is a
	// slab), liquid inside of the displacer won't flow from pos into side.
	SideClosed(pos, side BlockPos, w *World) bool
}

// RegisterBlock registers a block with the save name passed. The save name is used to save the block to the
// world's database and must not conflict with existing blocks.
// If a saveName is passed which already has a block registered for it, RegisterBlock panics.
func RegisterBlock(states ...Block) {
	if len(states) == 0 {
		panic("at least one block state must be registered")
	}
	for _, state := range states {
		name, props := state.EncodeBlock()
		key := keyStruct{name: name, pHash: hashProperties(props)}

		if _, ok := blocksHash[key]; ok {
			if _, unimplemented := state.(unimplementedBlock); !unimplemented {
				panic(fmt.Sprintf("cannot overwrite an existing block with the same name '%v' and properties %+v", name, props))
			} else {
				continue
			}
		}
		rid := uint32(len(registeredStates))

		runtimeIDsHashes.Put(int64(state.Hash()), int64(rid))
		registeredStates = append(registeredStates, state)

		filterLevel := uint8(15)
		if diffuser, ok := state.(lightDiffuser); ok {
			filterLevel = diffuser.LightDiffusionLevel()
		}
		chunk.FilteringBlocks = append(chunk.FilteringBlocks, filterLevel)
		emissionLevel := uint8(0)
		if emitter, ok := state.(lightEmitter); ok {
			emissionLevel = emitter.LightEmissionLevel()
		}
		chunk.LightBlocks = append(chunk.LightBlocks, emissionLevel)
		if removable, ok := state.(liquidRemovable); ok {
			world_internal.LiquidRemovable[rid] = removable.HasLiquidDrops()
		}
		if source, ok := state.(beaconSource); ok {
			world_internal.BeaconSource[rid] = source.PowersBeacon()
		}

		blocksHash[key] = state
		registerBlockByTypeName(state)
	}
}

// replaceableBlock represents a block that may be replaced by another block automatically. An example is
// grass, which may be replaced by clicking it with another block.
type replaceableBlock interface {
	// ReplaceableBy returns a bool which indicates if the block is replaceable by another block.
	ReplaceableBy(b Block) bool
}

// replaceable checks if the block at the position passed is replaceable with the block passed.
func replaceable(w *World, c *chunkData, pos BlockPos, with Block) bool {
	b, _ := w.blockInChunk(c, pos)
	if replaceable, ok := b.(replaceableBlock); ok {
		return replaceable.ReplaceableBy(with)
	}
	return false
}

// allBlocks returns a list of all registered states of the server. The list is ordered according to the
// runtime ID that the blocks have.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func allBlocks() []Block {
	return registeredStates
}

// blocksByName is a list of blocks indexed by their type names.
var blocksByName = map[string]Block{}

// RegisterBlockByTypeName registers the block passed by its type name. It converts the type name of blocks
// such as 'Leaves' to 'leaves' and the name of blocks suck as 'CoralFan' to 'coral_fan'.
func registerBlockByTypeName(b Block) {
	var name strings.Builder
	for i, r := range reflect.TypeOf(b).Name() {
		if unicode.IsUpper(r) {
			if i != 0 {
				name.WriteByte('_')
			}
			name.WriteRune(unicode.ToLower(r))
			continue
		}
		name.WriteRune(r)
	}
	blocksByName[name.String()] = reflect.New(reflect.TypeOf(b)).Elem().Interface().(Block)
}

// init registers all default states.
func init() {
	chunk.RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]interface{}, found bool) {
		block, ok := blockByRuntimeID(runtimeID)
		if !ok {
			return "", nil, false
		}
		name, properties = block.EncodeBlock()
		return name, properties, true
	}
	chunk.StateToRuntimeID = func(name string, properties map[string]interface{}) (runtimeID uint32, found bool) {
		blockInstance, ok := blockByNameAndProperties(name, properties)
		if !ok {
			return 0, false
		}
		return BlockRuntimeID(blockInstance)
	}
}

var registeredStates []Block
var blocksHash = map[keyStruct]Block{}
var runtimeIDsHashes = intintmap.New(8000, 0.95)

type keyStruct struct {
	name  string
	pHash string
}

// BlockRuntimeID attempts to return a runtime ID of a block state previously registered using
// RegisterBlock(). If the runtime ID is found, the bool returned is true. It is otherwise false.
func BlockRuntimeID(state Block) (uint32, bool) {
	if state == nil {
		return 0, true
	}
	runtimeID, ok := runtimeIDsHashes.Get(int64(state.Hash()))
	return uint32(runtimeID), ok
}

// blockByRuntimeID attempts to return a block state by its runtime ID. If not found, the bool returned is
// false. If found, the block state is non-nil and the bool true.
func blockByRuntimeID(runtimeID uint32) (Block, bool) {
	if runtimeID >= uint32(len(registeredStates)) {
		return nil, false
	}
	return registeredStates[runtimeID], true
}

// blockByNameAndProperties attempts to return a block instance by a name and its properties passed.
func blockByNameAndProperties(name string, properties map[string]interface{}) (b Block, found bool) {
	blockInstance, ok := blocksHash[keyStruct{name: name, pHash: hashProperties(properties)}]
	return blockInstance, ok
}

// registerAllStates registers all block states present in the game, skipping ones that have already been
// registered before this is called.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func registerAllStates() {
	var m []unimplementedBlock
	b, _ := base64.StdEncoding.DecodeString(resource.BlockStates)
	_ = nbt.Unmarshal(b, &m)

	for _, b := range m {
		key := keyStruct{name: b.Block.Name, pHash: hashProperties(b.Block.Properties)}
		if _, ok := blocksHash[key]; ok {
			// Duplicate state, don't add it.
			continue
		}
		RegisterBlock(b)
	}
}

var buffers = sync.Pool{New: func() interface{} {
	return bytes.NewBuffer(make([]byte, 0, 128))
}}

// hashProperties produces a hash for the block properties map passed.
// Passing the same map into hashProperties will always result in the same hash.
func hashProperties(properties map[string]interface{}) string {
	// TODO: Find a way to speed this up even more. Even though a lot of effort has been put into reducing the
	//  time this function takes, it is still too much. Any improvements to its performance have a large
	//  impact on large world modifications.

	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	radix.Sort(keys)

	b := buffers.Get().(*bytes.Buffer)
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
			a := uint32(v)
			b.WriteByte(byte(a))
			b.WriteByte(byte(a >> 8))
			b.WriteByte(byte(a >> 16))
			b.WriteByte(byte(a >> 24))
		case string:
			b.WriteString(v)
		}
	}

	data := append([]byte(nil), b.Bytes()...)
	b.Reset()
	buffers.Put(b)
	return *(*string)(unsafe.Pointer(&data))
}

// air returns an air block.
func air() Block {
	b, _ := blockByRuntimeID(0)
	return b
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

// Name ...
func (u unimplementedBlock) Name() string {
	return u.Block.Name
}

// EncodeBlock ...
func (u unimplementedBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return u.Block.Name, u.Block.Properties
}

// Hash ...
func (u unimplementedBlock) Hash() uint64 {
	return xxhash.Sum64String(u.Block.Name + hashProperties(u.Block.Properties))
}

// HasNBT ...
func (unimplementedBlock) HasNBT() bool {
	return false
}

// Model ...
func (unimplementedBlock) Model() BlockModel {
	return unimplementedModel{}
}

// unimplementedModel is the model used for unimplementedBlocks. It is the equivalent of a fully solid model.
type unimplementedModel struct{}

// AABB ...
func (u unimplementedModel) AABB(BlockPos, *World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
}

// FaceSolid ...
func (u unimplementedModel) FaceSolid(BlockPos, Face, *World) bool {
	return true
}

// Liquid represents a block that can be moved through and which can flow in the world after placement. There
// are two liquids in vanilla, which are lava and water.
type Liquid interface {
	Block
	// LiquidDepth returns the current depth of the liquid.
	LiquidDepth() int
	// SpreadDecay returns the amount of depth that is subtracted from the liquid's depth when it spreads to
	// a next block.
	SpreadDecay() int
	// WithDepth returns the liquid with the depth passed.
	WithDepth(depth int, falling bool) Liquid
	// LiquidFalling checks if the liquid is currently considered falling down.
	LiquidFalling() bool
	// LiquidType returns a string unique for the liquid, used to check if two liquids are considered to be
	// of the same type.
	LiquidType() string
	// Harden checks if the block should harden when looking at the surrounding blocks and sets the position
	// to the hardened block when adequate. If the block was hardened, the method returns true.
	Harden(pos BlockPos, w *World, flownIntoBy *BlockPos) bool
}
