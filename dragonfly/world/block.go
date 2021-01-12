package world

import (
	"bytes"
	"encoding/base64"
	"fmt"
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
// Every Block implementation must be able to be hashed as key in a map.
type Block interface {
	// HasNBT specifies if this Block has additional NBT present in the world save, also known as a block
	// entity. If true is returned, Block must implemented the NBTer interface.
	HasNBT() bool
	// Model returns the BlockModel of the Block.
	Model() BlockModel
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

// BlockRuntimeID attempts to return a runtime ID of a block state previously registered using
// RegisterBlock(). If the runtime ID is found, the bool returned is true. It is otherwise false.
func BlockRuntimeID(b Block) (uint32, bool) {
	if b == nil {
		return world_internal.AirRuntimeID, true
	}
	if (b.Model() == unimplementedModel{}) {
		s := b.(unimplementedBlock).BlockState
		return stateRuntimeIDs[stateHash{name: s.Name, properties: s.HashProperties()}], true
	}
	rid, ok := blockRuntimeIDs[b]
	return rid, ok
}

// blockByRuntimeID attempts to return a block state by its runtime ID. If not found, the bool returned is
// false. If found, the block state is non-nil and the bool true.
func blockByRuntimeID(rid uint32) (Block, bool) {
	if rid >= uint32(len(states)) {
		return air(), false
	}
	b, ok := blocks[rid]
	if !ok {
		b = unimplementedBlock{states[rid]}
	}
	return b, true
}

type BlockState struct {
	Name       string                 `nbt:"name"`
	Properties map[string]interface{} `nbt:"states"`
	Version    int32                  `nbt:"version"`
}

// buffers holds a sync.Pool of pooled byte buffers used to create a hash for the properties of a BlockState.
var buffers = sync.Pool{New: func() interface{} {
	return bytes.NewBuffer(make([]byte, 0, 128))
}}

// HashProperties produces a hash for the block properties held by the BlockState.
func (s BlockState) HashProperties() string {
	if s.Properties == nil {
		return ""
	}
	keys := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		keys = append(keys, k)
	}
	radix.Sort(keys)

	b := buffers.Get().(*bytes.Buffer)
	for _, k := range keys {
		switch v := s.Properties[k].(type) {
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

type stateHash struct {
	name, properties string
}

var states []BlockState
var stateRuntimeIDs = map[stateHash]uint32{}
var blockRuntimeIDs = map[Block]uint32{}
var blocks = map[uint32]Block{}

func RegisterBlockState(s BlockState) error {
	h := stateHash{name: s.Name, properties: s.HashProperties()}
	if _, ok := stateRuntimeIDs[h]; ok {
		return fmt.Errorf("cannot register the same state twice (%+v)", s)
	}
	fmt.Println(s.Name, s.Properties)
	rid := uint32(len(states))
	stateRuntimeIDs[h] = rid
	states = append(states, s)

	chunk.FilteringBlocks = append(chunk.FilteringBlocks, 15)
	chunk.LightBlocks = append(chunk.LightBlocks, 0)
	world_internal.LiquidRemovable = append(world_internal.LiquidRemovable, false)
	world_internal.BeaconSource = append(world_internal.BeaconSource, false)

	if s.Name == "minecraft:air" {
		world_internal.AirRuntimeID = rid
	}
	return nil
}

func RegisterBlock(b Block, s BlockState) error {
	h := stateHash{name: s.Name, properties: s.HashProperties()}

	if _, ok := blockRuntimeIDs[b]; ok {
		return fmt.Errorf("cannot register the same block twice (%#v)", b)
	}
	rid, ok := stateRuntimeIDs[h]
	if !ok {
		return fmt.Errorf("block state returned is not currently registered (%+v)", s)
	}
	blockRuntimeIDs[b] = rid
	blocks[rid] = b

	if diffuser, ok := b.(lightDiffuser); ok {
		chunk.FilteringBlocks[rid] = diffuser.LightDiffusionLevel()
	}
	if emitter, ok := b.(lightEmitter); ok {
		chunk.LightBlocks[rid] = emitter.LightEmissionLevel()
	}
	if removable, ok := b.(liquidRemovable); ok {
		world_internal.LiquidRemovable[rid] = removable.HasLiquidDrops()
	}
	if source, ok := b.(beaconSource); ok {
		world_internal.BeaconSource[rid] = source.PowersBeacon()
	}
	registerBlockByTypeName(b)
	return nil
}

// air returns an air block.
func air() Block {
	b, _ := blockByRuntimeID(world_internal.AirRuntimeID)
	return b
}

// unimplementedBlock represents a block that has not yet been implemented. It is used for registering block
// states that haven't yet been added.
type unimplementedBlock struct {
	BlockState
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

// init registers all default states.
func init() {
	b, _ := base64.StdEncoding.DecodeString(resource.BlockStates)
	dec := nbt.NewDecoder(bytes.NewBuffer(b))

	var s BlockState
	for {
		if err := dec.Decode(&s); err != nil {
			break
		}
		if err := RegisterBlockState(s); err != nil {
			// Should never happen.
			panic("duplicate block state registered")
		}
	}

	chunk.RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]interface{}, found bool) {
		if runtimeID >= uint32(len(states)) {
			return "", nil, false
		}
		s := states[runtimeID]
		return s.Name, s.Properties, true
	}
	chunk.StateToRuntimeID = func(name string, properties map[string]interface{}) (runtimeID uint32, found bool) {
		s := BlockState{Name: name, Properties: properties}
		h := stateHash{name: name, properties: s.HashProperties()}

		rid, ok := stateRuntimeIDs[h]
		return rid, ok
	}
}
