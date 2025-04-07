package world

import (
	"fmt"
	"github.com/brentp/intintmap"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/segmentio/fasthash/fnv1"
	"image"
	"math"
	"math/bits"
	"math/rand/v2"
	"slices"
	"sort"
)

// Block is a block that may be placed or found in a world. In addition, the block may also be added to an
// inventory: It is also an item.
// Every Block implementation must be able to be hashed as key in a map.
type Block interface {
	// EncodeBlock encodes the block to a string ID such as 'minecraft:grass' and properties associated
	// with the block.
	EncodeBlock() (string, map[string]any)
	// Hash returns two different identifiers for the block. The first is the base hash which is unique for
	// each type of block at runtime. For vanilla blocks, this is an auto-incrementing constant and for custom
	// blocks, you can call block.NextHash() to get a unique identifier. The second is the hash of the block's
	// own state and does not need to worry about colliding with other types of blocks. This is later combined
	// with the base hash to create a unique identifier for the full block.
	Hash() (uint64, uint64)
	// Model returns the BlockModel of the Block.
	Model() BlockModel
}

// CustomBlock represents a block that is non-vanilla and requires a resource pack and extra steps to show it to the
// client.
type CustomBlock interface {
	Block
	Properties() customblock.Properties
}

type CustomBlockBuildable interface {
	CustomBlock
	// Name is the name displayed to clients using the block.
	Name() string
	// Geometry is the geometries for the block that define the shape of the block. If false is returned, no custom
	// geometry will be applied. Permutation-specific geometry can be defined by returning a map of permutations to
	// geometry.
	Geometry() []byte
	// Textures is a map of images indexed by their target, used to map textures on to the block. Permutation-specific
	// textures can be defined by returning a map of permutations to textures.
	Textures() map[string]image.Image
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
	// BlastResistance is the blast resistance of the liquid, which influences the liquid's ability to withstand an
	// explosive blast.
	BlastResistance() float64
	// LiquidType returns an int unique for the liquid, used to check if two liquids are considered to be
	// of the same type.
	LiquidType() string
	// Harden checks if the block should harden when looking at the surrounding blocks and sets the position
	// to the hardened block when adequate. If the block was hardened, the method returns true.
	Harden(pos cube.Pos, tx *Tx, flownIntoBy *cube.Pos) bool
}

// hashes holds a list of runtime IDs indexed by the hash of the Block that implements the blocks pointed to by those
// runtime IDs. It is used to look up a block's runtime ID quickly.
var (
	bitSize int
	hashes  = intintmap.New(7000, 0.999)
)

// RegisterBlock registers the Block passed. The EncodeBlock method will be used to encode and decode the
// block passed. RegisterBlock panics if the block properties returned were not valid, existing properties.
func RegisterBlock(b Block) {
	if bitSize > 0 {
		panic(fmt.Errorf("tried to register a block after the block registry was finalised"))
	}
	name, properties := b.EncodeBlock()
	if _, ok := b.(CustomBlock); ok {
		registerBlockState(blockState{Name: name, Properties: properties})
	}
	rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		// We assume all blocks must have all their states registered beforehand. Vanilla blocks will have
		// this done through registering of all states present in the block_states.nbt file.
		panic(fmt.Sprintf("block state returned is not registered (%v {%#v})", name, properties))
	}
	if _, ok := blocks[rid].(unknownBlock); !ok {
		panic(fmt.Sprintf("block with name and properties %v {%#v} already registered", name, properties))
	}
	blocks[rid] = b
	if c, ok := b.(CustomBlock); ok {
		if _, ok := customBlocks[name]; !ok {
			customBlocks[name] = c
		}
	}
}

// finaliseBlockRegistry is called after blocks have finished registering and the palette can be sorted and
// hashed, which also calls finaliseBlock for each block that has been registered up to this point.
// noinspection GoUnusedFunction
//
//lint:ignore U1000 Function is used through compiler directives.
func finaliseBlockRegistry() {
	if bitSize > 0 {
		return
	}
	bitSize = bits.Len64(uint64(len(blocks)))
	sort.SliceStable(blocks, func(i, j int) bool {
		nameOne, _ := blocks[i].EncodeBlock()
		nameTwo, _ := blocks[j].EncodeBlock()
		return nameOne != nameTwo && fnv1.HashString64(nameOne) < fnv1.HashString64(nameTwo)
	})
	for rid, b := range blocks {
		finaliseBlock(uint32(rid), b)
		if _, hash := b.Hash(); hash != math.MaxUint64 {
			// b is not an unknownBlock.
			h := int64(BlockHash(b))
			if other, ok := hashes.Get(h); ok {
				panic(fmt.Sprintf("block %#v with hash %v already registered by %#v", b, h, blocks[other]))
			}
			hashes.Put(h, int64(rid))
		}
	}
}

// finaliseBlock stores the necessary information for the provided block to be quickly accessed at runtime.
func finaliseBlock(rid uint32, b Block) {
	name, properties := b.EncodeBlock()
	i := stateHash{name: name, properties: hashProperties(properties)}
	if name == "minecraft:air" {
		airRID = rid
	}
	stateRuntimeIDs[i] = rid

	if diffuser, ok := b.(lightDiffuser); ok {
		chunk.FilteringBlocks[rid] = diffuser.LightDiffusionLevel()
	}
	if emitter, ok := b.(lightEmitter); ok {
		chunk.LightBlocks[rid] = emitter.LightEmissionLevel()
	}
	if _, ok := b.(NBTer); ok {
		nbtBlocks[rid] = true
	}
	if _, ok := b.(RandomTicker); ok {
		randomTickBlocks[rid] = true
	}
	if _, ok := b.(Liquid); ok {
		liquidBlocks[rid] = true
	}
	if _, ok := b.(LiquidDisplacer); ok {
		liquidDisplacingBlocks[rid] = true
	}
}

// BlockHash returns a unique identifier of the block including the block states. This function is used internally
// to convert a block to a single integer which can be used in map lookups. The hash produced therefore does not
// need to match anything in the game, but it must be unique among all registered blocks.
// The tool in `/cmd/blockhash` may be used to automatically generate block hashes of blocks in a package.
func BlockHash(b Block) uint64 {
	base, hash := b.Hash()
	return base | (hash << bitSize)
}

// BlockRuntimeID attempts to return a runtime ID of a block previously registered using RegisterBlock().
// If the runtime ID cannot be found because the Block wasn't registered, BlockRuntimeID will panic.
func BlockRuntimeID(b Block) uint32 {
	if b == nil {
		return airRID
	}
	if _, h := b.Hash(); h != math.MaxUint64 {
		// b is not an unknownBlock.
		if rid, ok := hashes.Get(int64(BlockHash(b))); ok {
			return uint32(rid)
		}
		panic(fmt.Sprintf("cannot find block by non-0 hash of block %#v", b))
	}
	return slowBlockRuntimeID(b)
}

// slowBlockRuntimeID finds the runtime ID of a Block by hashing the properties produced by calling the
// Block.EncodeBlock method and looking it up in the stateRuntimeIDs map.
func slowBlockRuntimeID(b Block) uint32 {
	name, properties := b.EncodeBlock()

	rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		panic(fmt.Sprintf("cannot find block by (name + properties): %#v", b))
	}
	return rid
}

// BlockByRuntimeID attempts to return a Block by its runtime ID. If not found, the bool returned is
// false. If found, the block is non-nil and the bool true.
func BlockByRuntimeID(rid uint32) (Block, bool) {
	if rid >= uint32(len(blocks)) {
		return air(), false
	}
	return blocks[rid], true
}

func blockByRuntimeIDOrAir(rid uint32) Block {
	bl, _ := BlockByRuntimeID(rid)
	return bl
}

// BlockByName attempts to return a Block by its name and properties. If not found, the bool returned is
// false.
func BlockByName(name string, properties map[string]any) (Block, bool) {
	rid, ok := stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		return nil, false
	}
	return blocks[rid], true
}

// Blocks returns a slice of all registered blocks.
func Blocks() []Block {
	return slices.Clone(blocks)
}

// CustomBlocks returns a map of all custom blocks registered with their names as keys.
func CustomBlocks() map[string]CustomBlock {
	return customBlocks
}

// air returns an air block.
func air() Block {
	return blocks[airRID]
}

// RandomTicker represents a block that executes an action when it is ticked randomly. Every 20th of a second,
// one random block in each sub chunk are picked to receive a random tick.
type RandomTicker interface {
	// RandomTick handles a random tick of the block at the position passed. Additionally, a rand.RandSource
	// instance is passed which may be used to generate values randomly without locking.
	RandomTick(pos cube.Pos, tx *Tx, r *rand.Rand)
}

// ScheduledTicker represents a block that executes an action when it has a block update scheduled, such as
// when a block adjacent to it is broken.
type ScheduledTicker interface {
	// ScheduledTick handles a scheduled tick initiated by an event in one of the neighbouring blocks, such as
	// when a block is placed or broken. Additionally, a rand.RandSource instance is passed which may be used to
	// generate values randomly without locking.
	ScheduledTick(pos cube.Pos, tx *Tx, r *rand.Rand)
}

// TickerBlock is an implementation of NBTer with an additional Tick method that is called on every world
// tick for loaded blocks that implement this interface.
type TickerBlock interface {
	NBTer
	Tick(currentTick int64, pos cube.Pos, tx *Tx)
}

// NeighbourUpdateTicker represents a block that is updated when a block adjacent to it is updated, either
// through placement or being broken.
type NeighbourUpdateTicker interface {
	// NeighbourUpdateTick handles a neighbouring block being updated. The position of that block and the
	// position of this block is passed.
	NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *Tx)
}

// NBTer represents either an item or a block which may decode NBT data and encode to NBT data. Typically,
// this is done to store additional data.
type NBTer interface {
	// DecodeNBT returns the (new) item, block or Entity, depending on which of those the NBTer was, with the NBT data
	// decoded into it.
	DecodeNBT(data map[string]any) any
	// EncodeNBT encodes the Entity into a map which can then be encoded as NBT to be written.
	EncodeNBT() map[string]any
}

// LiquidDisplacer represents a block that is able to displace a liquid to a different world layer, without
// fully removing the liquid.
type LiquidDisplacer interface {
	// CanDisplace specifies if the block is able to displace the liquid passed.
	CanDisplace(b Liquid) bool
	// SideClosed checks if a position on the side of the block placed in the world at a specific position is
	// closed. When this returns true (for example, when the side is below the position and the block is a
	// slab), liquid inside the displacer won't flow from pos into side.
	SideClosed(pos, side cube.Pos, tx *Tx) bool
}

// lightEmitter is identical to a block.LightEmitter.
type lightEmitter interface {
	LightEmissionLevel() uint8
}

// lightDiffuser is identical to a block.LightDiffuser.
type lightDiffuser interface {
	LightDiffusionLevel() uint8
}

// replaceableBlock represents a block that may be replaced by another block automatically. An example is
// grass, which may be replaced by clicking it with another block.
type replaceableBlock interface {
	// ReplaceableBy returns a bool which indicates if the block is replaceable by another block.
	ReplaceableBy(b Block) bool
}

// replaceable checks if the block at the position passed is replaceable with the block passed.
func replaceable(w *World, c *Column, pos cube.Pos, with Block) bool {
	if r, ok := w.blockInChunk(c, pos).(replaceableBlock); ok {
		return r.ReplaceableBy(with)
	}
	return false
}

// BlockAction represents an action that may be performed by a block. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type BlockAction interface {
	BlockAction()
}
