package world

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/world_internal"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"math/rand"
)

// Block is a block that may be placed or found in a world. In addition, the block may also be added to an
// inventory: It is also an item.
// Every Block implementation must be able to be hashed as key in a map.
type Block interface {
	// EncodeBlock encodes the block to a string ID such as 'minecraft:grass' and properties associated
	// with the block.
	EncodeBlock() (string, map[string]interface{})
	// Model returns the BlockModel of the Block.
	Model() BlockModel
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
	// LiquidType returns an int unique for the liquid, used to check if two liquids are considered to be
	// of the same type.
	LiquidType() int
	// Harden checks if the block should harden when looking at the surrounding blocks and sets the position
	// to the hardened block when adequate. If the block was hardened, the method returns true.
	Harden(pos cube.Pos, w *World, flownIntoBy *cube.Pos) bool
}

// RegisterBlock registers the Block passed. The EncodeBlock method will be used to encode and decode the
// block passed. RegisterBlock panics if the block properties returned were not valid, existing properties.
func RegisterBlock(b Block) {
	name, properties := b.EncodeBlock()
	h := stateHash{name: name, properties: hashProperties(properties)}

	rid, ok := stateRuntimeIDs[h]
	if !ok {
		// We assume all blocks must have all their states registered beforehand. Vanilla blocks will have
		// this done through registering of all states present in the block_states.nbt file.
		panic(fmt.Sprintf("block state returned is not registered (%v {%#v})", name, properties))
	}
	if _, ok := blocks[rid].(unknownBlock); !ok {
		panic(fmt.Sprintf("block with name and properties %v {%#v} already registered", name, properties))
	}
	blocks[rid] = b

	if diffuser, ok := b.(lightDiffuser); ok {
		chunk.FilteringBlocks[rid] = diffuser.LightDiffusionLevel()
	}
	if emitter, ok := b.(lightEmitter); ok {
		chunk.LightBlocks[rid] = emitter.LightEmissionLevel()
	}
	if _, ok := b.(liquidRemovable); ok {
		world_internal.LiquidRemovable[rid] = true
	}
	if source, ok := b.(beaconSource); ok {
		world_internal.BeaconSource[rid] = source.PowersBeacon()
	}
	if _, ok := b.(NBTer); ok {
		nbtBlocks[rid] = true
	}
}

// BlockRuntimeID attempts to return a runtime ID of a block previously registered using RegisterBlock().
// If the runtime ID is found, the bool returned is true. It is otherwise false.
func BlockRuntimeID(b Block) (uint32, bool) {
	if b == nil {
		return world_internal.AirRuntimeID, true
	}
	name, properties := b.EncodeBlock()

	rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	return rid, ok
}

// BlockByRuntimeID attempts to return a Block by its runtime ID. If not found, the bool returned is
// false. If found, the block is non-nil and the bool true.
func BlockByRuntimeID(rid uint32) (Block, bool) {
	if rid >= uint32(len(blocks)) {
		return air(), false
	}
	return blocks[rid], true
}

// BlockByName attempts to return a Block by its name and properties. If not found, the bool returned is
// false.
func BlockByName(name string, properties map[string]interface{}) (Block, bool) {
	rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		return nil, false
	}
	return blocks[rid], true
}

// air returns an air block.
func air() Block {
	b, _ := BlockByRuntimeID(world_internal.AirRuntimeID)
	return b
}

// RandomTicker represents a block that executes an action when it is ticked randomly. Every 20th of a second,
// one random block in each sub chunk are picked to receive a random tick.
type RandomTicker interface {
	// RandomTick handles a random tick of the block at the position passed. Additionally, a rand.Rand
	// instance is passed which may be used to generate values randomly without locking.
	RandomTick(pos cube.Pos, w *World, r *rand.Rand)
}

// ScheduledTicker represents a block that executes an action when it has a block update scheduled, such as
// when a block adjacent to it is broken.
type ScheduledTicker interface {
	// ScheduledTick handles a scheduled tick initiated by an event in one of the neighbouring blocks, such as
	// when a block is placed or broken.
	ScheduledTick(pos cube.Pos, w *World)
}

// TickerBlock is an implementation of NBTer with an additional Tick method that is called on every world
// tick for loaded blocks that implement this interface.
type TickerBlock interface {
	NBTer
	Tick(currentTick int64, pos cube.Pos, w *World)
}

// NeighbourUpdateTicker represents a block that is updated when a block adjacent to it is updated, either
// through placement or being broken.
type NeighbourUpdateTicker interface {
	// NeighbourUpdateTick handles a neighbouring block being updated. The position of that block and the
	// position of this block is passed.
	NeighbourUpdateTick(pos, changedNeighbour cube.Pos, w *World)
}

// NBTer represents either an item or a block which may decode NBT data and encode to NBT data. Typically
// this is done to store additional data.
type NBTer interface {
	// DecodeNBT returns the item or block, depending on which of those the NBTer was, with the NBT data
	// decoded into it.
	DecodeNBT(data map[string]interface{}) interface{}
	EncodeNBT() map[string]interface{}
}

// LiquidDisplacer represents a block that is able to displace a liquid to a different world layer, without
// fully removing the liquid.
type LiquidDisplacer interface {
	// CanDisplace specifies if the block is able to displace the liquid passed.
	CanDisplace(b Liquid) bool
	// SideClosed checks if a position on the side of the block placed in the world at a specific position is
	// closed. When this returns true (for example, when the side is below the position and the block is a
	// slab), liquid inside of the displacer won't flow from pos into side.
	SideClosed(pos, side cube.Pos, w *World) bool
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

// replaceableBlock represents a block that may be replaced by another block automatically. An example is
// grass, which may be replaced by clicking it with another block.
type replaceableBlock interface {
	// ReplaceableBy returns a bool which indicates if the block is replaceable by another block.
	ReplaceableBy(b Block) bool
}

// replaceable checks if the block at the position passed is replaceable with the block passed.
func replaceable(w *World, c *chunkData, pos cube.Pos, with Block) bool {
	b, _ := w.blockInChunk(c, pos)
	if r, ok := b.(replaceableBlock); ok {
		return r.ReplaceableBy(with)
	}
	return false
}
