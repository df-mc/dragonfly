package world

import (
	"fmt"
	"maps"
	"math"
	"math/bits"
	"slices"
	"sort"
	"strings"

	"github.com/df-mc/dragonfly/server/internal/intintmap"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/segmentio/fasthash/fnv1"
)

// DefaultBlockRegistry is the default (vanilla) block registry used by Dragonfly when no custom registry is provided.
//
// The registry is populated during package init (block states + block implementations) and is finalized on first use
// by `server.Config.New()`, `world.Config.New()`, or `mcdb.Config.Open()`. Callers that need custom blocks should
// create a new registry using `NewBlockRegistry()` and pass it through config instead of mutating this value.
var DefaultBlockRegistry BlockRegistry = &BlockRegistryImpl{
	blockProperties: make(map[string]map[string]any),
	stateRuntimeIDs: make(map[stateHash]uint32),
	customBlocks:    make(map[string]CustomBlock),
}

// BlockRegistry converts between runtime IDs and block states/implementations.
//
// A BlockRegistry has a build/finalize lifecycle:
//   - During setup, blocks and block states may be registered using RegisterBlockState/RegisterBlock.
//   - After calling Finalize, the registry becomes immutable and is ready for use in chunk encoding/decoding and
//     network serialization. RegisterBlock/RegisterBlockState will panic after finalization.
//
// The interface is split because the chunk package cannot import world.Block.
type BlockRegistry interface {
	chunk.BlockRegistry

	// BlockByRuntimeID looks up a Block by runtime ID. If the runtime ID is unknown/out of range, ok is false.
	BlockByRuntimeID(rid uint32) (Block, bool)
	// BlockByRuntimeIDOrAir looks up a Block by runtime ID. If not found, an air block is returned.
	BlockByRuntimeIDOrAir(rid uint32) Block
	// BlockRuntimeID looks up the runtime ID of a previously registered Block.
	BlockRuntimeID(block Block) (rid uint32)
	// RegisterBlock registers a Block implementation for a previously registered block state.
	RegisterBlock(block Block)
	// RegisterBlockState registers a block state that blocks may encode to.
	RegisterBlockState(blockState BlockState)
	// CustomBlocks returns custom blocks registered in this registry, keyed by identifier.
	CustomBlocks() map[string]CustomBlock
	// BlockByName looks up a Block by full identifier and properties.
	BlockByName(name string, properties map[string]any) (Block, bool)
	// Blocks returns all blocks registered in the registry, indexed by runtime ID.
	Blocks() []Block
	// Air returns the air block registered in the registry.
	Air() Block
	// Clone returns an independent copy of the registry. If the source registry is finalized, the clone is also
	// finalized. If the source is not finalized, the clone remains mutable.
	Clone() BlockRegistry
	// Finalize finalizes the registry, building derived lookup tables required for runtime usage. Finalize is
	// idempotent.
	Finalize()
	// BitSize returns the number of bits used by BlockHash (depends on the number of registered blocks).
	BitSize() int
	// BlockHash returns a unique identifier of the block including the block states. The hash is internal to Dragonfly
	// and is used for fast map lookups; it does not need to match any in-game identifiers.
	BlockHash(b Block) uint64
}

const (
	blockFlagNBT uint16 = 1 << iota
	blockFlagRandomTick
	blockFlagLiquid
	blockFlagLiquidDisplacing
)

type blockInfo uint16

func (b *blockInfo) set(flag uint16) {
	*b ^= blockInfo(flag)
}

func (b blockInfo) get(flag uint16) bool {
	return uint16(b)&flag != 0
}

func (b *blockInfo) setLight(light uint8) {
	*b ^= blockInfo(light) << 8
}

func (b *blockInfo) setLightFilter(light uint8) {
	*b ^= blockInfo(light) << 12
}

func (b blockInfo) getLight() uint8 {
	return uint8((b >> 8) & 0xF)
}

func (b blockInfo) getLightFilter() uint8 {
	return uint8((b >> 12) & 0xF)
}

type BlockRegistryImpl struct {
	finalized         bool
	bitSize           int
	hashes            *intintmap.Map
	networkhashToRids map[uint32]uint32

	// stateRuntimeIDs holds a map for looking up the runtime ID of a block by the stateHash it produces.
	stateRuntimeIDs map[stateHash]uint32

	blockProperties map[string]map[string]any
	// blocks holds a list of all registered Blocks indexed by their runtime ID. Blocks that were not explicitly
	// registered are of the type unknownBlock.
	blocks []Block
	// customBlocks maps a custom block's identifier to a slice of custom blocks.
	customBlocks map[string]CustomBlock

	blockInfos []blockInfo

	airRID uint32
}

func (br *BlockRegistryImpl) BitSize() int {
	if !br.finalized {
		panic("BlockRegistry.BitSize called on non finalized BlockRegistry")
	}
	return br.bitSize
}

func (br *BlockRegistryImpl) BlockCount() int {
	if !br.finalized {
		panic("BlockRegistry.BlockCount called on non finalized BlockRegistry")
	}
	return len(br.blockInfos)
}

func (br *BlockRegistryImpl) RandomTickBlock(rid uint32) bool {
	if !br.finalized {
		panic("BlockRegistry.RandomTickBlock called on non finalized BlockRegistry")
	}
	return br.blockInfos[rid].get(blockFlagRandomTick)
}

func (br *BlockRegistryImpl) FilteringBlock(rid uint32) uint8 {
	if !br.finalized {
		panic("BlockRegistry.FilteringBlock called on non finalized BlockRegistry")
	}
	return br.blockInfos[rid].getLightFilter()
}

func (br *BlockRegistryImpl) LightBlock(rid uint32) uint8 {
	if !br.finalized {
		panic("BlockRegistry.LightBlock called on non finalized BlockRegistry")
	}
	return br.blockInfos[rid].getLight()
}

func (br *BlockRegistryImpl) NBTBlock(rid uint32) bool {
	if !br.finalized {
		panic("BlockRegistry.NBTBlock called on non finalized BlockRegistry")
	}
	return br.blockInfos[rid].get(blockFlagNBT)
}

func (br *BlockRegistryImpl) LiquidDisplacingBlock(rid uint32) bool {
	if !br.finalized {
		panic("BlockRegistry.LiquidDisplacingBlock called on non finalized BlockRegistry")
	}
	return br.blockInfos[rid].get(blockFlagLiquidDisplacing)
}

func (br *BlockRegistryImpl) LiquidBlock(rid uint32) bool {
	if !br.finalized {
		panic("BlockRegistry.LiquidBlock called on non finalized BlockRegistry")
	}
	return br.blockInfos[rid].get(blockFlagLiquid)
}

func (br *BlockRegistryImpl) Blocks() []Block {
	return slices.Clone(br.blocks)
}

func (br *BlockRegistryImpl) HashToRuntimeID(hash uint32) (rid uint32, ok bool) {
	if !br.finalized {
		panic("BlockRegistry.HashToRuntimeID called on non finalized BlockRegistry")
	}
	rid, ok = br.networkhashToRids[hash]
	return
}

func (br *BlockRegistryImpl) Clone() BlockRegistry {
	br2 := &BlockRegistryImpl{
		blockProperties: make(map[string]map[string]any, len(br.blockProperties)),
		stateRuntimeIDs: make(map[stateHash]uint32, len(br.stateRuntimeIDs)),
		customBlocks:    make(map[string]CustomBlock, len(br.customBlocks)),
		finalized:       br.finalized,
		bitSize:         br.bitSize,
		airRID:          br.airRID,
	}

	for k, v := range br.blockProperties {
		br2.blockProperties[k] = maps.Clone(v)
	}
	maps.Copy(br2.stateRuntimeIDs, br.stateRuntimeIDs)
	maps.Copy(br2.customBlocks, br.customBlocks)

	br2.blocks = make([]Block, len(br.blocks))
	copy(br2.blocks, br.blocks)
	br2.blockInfos = append([]blockInfo(nil), br.blockInfos...)

	if br.finalized {
		br2.hashes = intintmap.New(len(br.blocks), 0.999)
		for rid, b := range br2.blocks {
			if _, hash := b.Hash(); hash == math.MaxUint64 {
				continue
			}
			br2.hashes.Put(int64(br2.BlockHash(b)), int64(rid))
		}
		br2.networkhashToRids = make(map[uint32]uint32, len(br.networkhashToRids))
		maps.Copy(br2.networkhashToRids, br.networkhashToRids)
	}

	return br2
}

// NewBlockRegistry returns a mutable registry seeded with all vanilla block states and block implementations.
// Callers may RegisterBlockState/RegisterBlock and must call Finalize() before using the registry for world/chunk
// serialization. The returned registry is independent from DefaultBlockRegistry.
func NewBlockRegistry() BlockRegistry {
	// Clone() produces a fully populated registry. We then mark it mutable again so that additional block states and
	// blocks can be registered before finalizing.
	br := DefaultBlockRegistry.Clone().(*BlockRegistryImpl)
	br.finalized = false
	br.bitSize = 0
	br.hashes = nil
	br.networkhashToRids = nil
	br.blockInfos = nil
	return br
}

// RegisterBlock registers the Block passed. The EncodeBlock method will be used to encode and decode the
// block passed. RegisterBlock panics if the block properties returned were not valid, existing properties.
func (br *BlockRegistryImpl) RegisterBlock(b Block) {
	if br.finalized {
		panic("BlockRegistry.RegisterBlock called on finalized BlockRegistry")
	}
	if br.bitSize > 0 {
		panic(fmt.Errorf("tried to register a block after the block registry was finalised"))
	}
	name, properties := b.EncodeBlock()
	if _, ok := b.(CustomBlock); ok {
		br.RegisterBlockState(BlockState{Name: name, Properties: properties})
	}
	rid, ok := br.stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		// We assume all blocks must have all their states registered beforehand. Vanilla blocks will have
		// this done through registering of all states present in the block_states.nbt file.
		panic(fmt.Sprintf("block state returned is not registered (%v {%#v})", name, properties))
	}
	if _, ok := br.blocks[rid].(unknownBlock); !ok {
		panic(fmt.Sprintf("block with name and properties %v {%#v} already registered", name, properties))
	}
	br.blocks[rid] = b
	if c, ok := b.(CustomBlock); ok {
		if _, ok := br.customBlocks[name]; !ok {
			br.customBlocks[name] = c
		}
	}
}

// RegisterBlockState registers a blockStates to the states slice. The function panics if the properties the
// blockState hold are invalid or if the blockState was already registered.
func (br *BlockRegistryImpl) RegisterBlockState(s BlockState) {
	if br.finalized {
		panic("BlockRegistry.RegisterBlockState called on finalized BlockRegistry")
	}
	h := stateHash{name: s.Name, properties: hashProperties(s.Properties)}
	if _, ok := br.stateRuntimeIDs[h]; ok {
		panic(fmt.Sprintf("cannot register the same state twice (%+v)", s))
	}
	if _, ok := br.blockProperties[s.Name]; !ok {
		br.blockProperties[s.Name] = s.Properties
	}
	rid := uint32(len(br.blocks))
	br.blocks = append(br.blocks, unknownBlock{BlockState: s})
	br.stateRuntimeIDs[h] = rid
}

func (br *BlockRegistryImpl) Finalize() {
	if br.finalized {
		// Finalize is intentionally idempotent so code can defensively call it when a registry is provided through
		// config. RegisterBlock/RegisterBlockState still panic after finalization.
		return
	}
	br.finalized = true

	br.bitSize = bits.Len64(uint64(len(br.blocks)))
	sort.SliceStable(br.blocks, func(i, j int) bool {
		var nameOne string
		if b1, ok := br.blocks[i].(unknownBlock); ok {
			nameOne = b1.Name
		} else {
			nameOne, _ = br.blocks[i].EncodeBlock()
		}
		var nameTwo string
		if b2, ok := br.blocks[j].(unknownBlock); ok {
			nameTwo = b2.Name
		} else {
			nameTwo, _ = br.blocks[j].EncodeBlock()
		}
		return fnv1.HashString64(nameOne) < fnv1.HashString64(nameTwo)
	})

	br.blockInfos = make([]blockInfo, len(br.blocks))
	br.hashes = intintmap.New(len(br.blocks), 0.999)
	br.networkhashToRids = make(map[uint32]uint32, len(br.blocks))
	br.stateRuntimeIDs = make(map[stateHash]uint32, len(br.blocks))

	for idx, b := range br.blocks {
		rid := uint32(idx)
		name, properties := b.EncodeBlock()
		h := stateHash{name: name, properties: hashProperties(properties)}
		if name == "minecraft:air" {
			br.airRID = rid
		}
		if _, ok := br.stateRuntimeIDs[h]; ok {
			panic(fmt.Sprintf("cannot register the same state twice (%s %+v)", name, properties))
		}
		br.stateRuntimeIDs[h] = rid

		var info blockInfo
		if diffuser, ok := b.(lightDiffuser); ok {
			info.setLightFilter(diffuser.LightDiffusionLevel())
		}
		if emitter, ok := b.(lightEmitter); ok {
			info.setLight(emitter.LightEmissionLevel())
		}
		if _, ok := b.(NBTer); ok {
			info.set(blockFlagNBT)
		}
		if _, ok := b.(RandomTicker); ok {
			info.set(blockFlagRandomTick)
		}
		if _, ok := b.(Liquid); ok {
			info.set(blockFlagLiquid)
		}
		if _, ok := b.(LiquidDisplacer); ok {
			info.set(blockFlagLiquidDisplacing)
		}
		br.blockInfos[rid] = info

		if _, hash := b.Hash(); hash != math.MaxUint64 {
			h := int64(br.BlockHash(b))
			if other, ok := br.hashes.Get(h); ok {
				panic(fmt.Sprintf("block %#v with hash %v already registered by %#v", b, h, br.blocks[other]))
			}
			br.hashes.Put(h, int64(rid))
		}
		br.networkhashToRids[networkBlockHash(name, properties)] = rid
	}
}

// AirRuntimeID returns the runtime ID of the air block.
func (br *BlockRegistryImpl) AirRuntimeID() uint32 {
	return br.airRID
}

// RuntimeIDToState returns the name and state properties of a block by its runtime ID.
func (br *BlockRegistryImpl) RuntimeIDToState(runtimeID uint32) (name string, properties map[string]any, found bool) {
	if !br.finalized {
		panic("BlockRegistry.RuntimeIDToState called on non finalized BlockRegistry")
	}
	if runtimeID >= uint32(len(br.blocks)) {
		return "", nil, false
	}
	name, properties = br.blocks[runtimeID].EncodeBlock()
	return name, properties, true
}

// StateToRuntimeID returns the runtime ID of a block by its name and state properties.
func (br *BlockRegistryImpl) StateToRuntimeID(name string, properties map[string]any) (runtimeID uint32, found bool) {
	if !br.finalized {
		panic("BlockRegistry.StateToRuntimeID called on non finalized BlockRegistry")
	}
	if rid, ok := br.stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]; ok {
		return rid, true
	}
	rid, ok := br.stateRuntimeIDs[stateHash{name: name, properties: hashProperties(br.blockProperties[name])}]
	return rid, ok
}

// BlockHash returns a unique identifier of the block including the block states. This function is used internally
// to convert a block to a single integer which can be used in map lookups. The hash produced therefore does not
// need to match anything in the game, but it must be unique among all registered blocks.
// The tool in `/cmd/blockhash` may be used to automatically generate block hashes of blocks in a package.
func (br *BlockRegistryImpl) BlockHash(b Block) uint64 {
	base, hash := b.Hash()
	return base | (hash << uint64(br.bitSize))
}

// BlockRuntimeID attempts to return a runtime ID of a block previously registered using RegisterBlock().
// If the runtime ID cannot be found because the Block wasn't registered, BlockRuntimeID will panic.
func (br *BlockRegistryImpl) BlockRuntimeID(b Block) uint32 {
	if !br.finalized {
		panic("BlockRegistry.BlockRuntimeID called on non finalized BlockRegistry")
	}
	if b == nil {
		return br.airRID
	}
	if _, hash := b.Hash(); hash != math.MaxUint64 {
		// b is not an unknownBlock.
		if rid, ok := br.hashes.Get(int64(br.BlockHash(b))); ok {
			return uint32(rid)
		}
		panic(fmt.Sprintf("cannot find block by non-0 hash of block %#v", b))
	}
	return br.slowBlockRuntimeID(b)
}

func (br *BlockRegistryImpl) BlockByRuntimeIDOrAir(rid uint32) Block {
	bl, _ := br.BlockByRuntimeID(rid)
	return bl
}

// slowBlockRuntimeID finds the runtime ID of a Block by hashing the properties produced by calling the
// Block.EncodeBlock method and looking it up in the stateRuntimeIDs map.
func (br *BlockRegistryImpl) slowBlockRuntimeID(b Block) uint32 {
	name, properties := b.EncodeBlock()
	rid, ok := br.stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		panic(fmt.Sprintf("cannot find block by (name + properties): %#v", b))
	}
	return rid
}

// BlockByRuntimeID attempts to return a Block by its runtime ID. If not found, the bool returned is
// false. If found, the block is non-nil and the bool true.
func (br *BlockRegistryImpl) BlockByRuntimeID(rid uint32) (Block, bool) {
	if !br.finalized {
		panic("BlockRegistry.BlockByRuntimeID called on non finalized BlockRegistry")
	}
	if rid >= uint32(len(br.blocks)) {
		return br.Air(), false
	}
	return br.blocks[rid], true
}

// BlockByNetworkID attempts to return a Block by its static network ID. If not found, the bool returned is
// false. If found, the block is non-nil and the bool true.
func (br *BlockRegistryImpl) BlockByNetworkID(rid uint32) (Block, bool) {
	if !br.finalized {
		panic("BlockRegistry.BlockByNetworkID called on non finalized BlockRegistry")
	}
	if rid >= uint32(len(br.blocks)) {
		return br.Air(), false
	}
	return br.blocks[rid], true
}

// BlockByName attempts to return a Block by its name and properties. If not found, the bool returned is
// false.
func (br *BlockRegistryImpl) BlockByName(name string, properties map[string]any) (Block, bool) {
	if !br.finalized {
		panic("BlockRegistry.BlockByName called on non finalized BlockRegistry")
	}
	rid, ok := br.stateRuntimeIDs[stateHash{name: name, properties: hashProperties(properties)}]
	if !ok {
		return nil, false
	}
	return br.blocks[rid], true
}

// CustomBlocks returns a map of all custom blocks registered with their names as keys.
func (br *BlockRegistryImpl) CustomBlocks() map[string]CustomBlock {
	return maps.Clone(br.customBlocks)
}

// Air returns an air block.
func (br *BlockRegistryImpl) Air() Block {
	b, _ := br.BlockByRuntimeID(br.airRID)
	return b
}

var traitLookup = map[string][]any{
	"minecraft:facing_direction": {
		"north", "east", "south", "west", "down", "up",
	},
	"minecraft:cardinal_direction": {
		"north", "east", "south", "west",
	},
	"minecraft:vertical_half": {
		"top", "bottom",
	},
	"minecraft:block_face": {
		"north", "east", "south", "west", "down", "up",
	},
}

func AddCustomBlocks(reg BlockRegistry, entries []protocol.BlockEntry) error {
	for _, entry := range entries {
		ns, _ := splitNamespace(entry.Name)
		if ns == "minecraft" {
			continue
		}

		var propertyNames []string
		var propertyValues [][]any

		props, ok := entry.Properties["properties"].([]any)
		if ok {
			for _, v := range props {
				v := v.(map[string]any)
				name := v["name"].(string)
				enum := v["enum"].([]any)
				propertyNames = append(propertyNames, name)
				propertyValues = append(propertyValues, enum)
			}
		}

		traits, ok := entry.Properties["traits"].([]any)
		if ok {
			for _, trait := range traits {
				trait := trait.(map[string]any)
				enabled_states := trait["enabled_states"].(map[string]any)
				for k, enabled := range enabled_states {
					if !strings.ContainsRune(k, ':') {
						k = "minecraft:" + k
					}
					enabled := enabled.(uint8)
					if enabled == 0 {
						continue
					}
					v, ok := traitLookup[k]
					if !ok {
						return fmt.Errorf("unresolved trait %s", k)
					}

					propertyNames = append(propertyNames, k)
					propertyValues = append(propertyValues, v)
				}
			}
		}

		permutations := cartesianProduct(propertyValues)

		for _, values := range permutations {
			m := make(map[string]any)
			for i, value := range values {
				name := propertyNames[i]
				m[name] = value
			}
			reg.RegisterBlockState(BlockState{
				Name:       entry.Name,
				Properties: m,
			})
		}
	}

	return nil
}

func cartesianProduct(inputs [][]any) [][]any {
	if len(inputs) == 0 {
		return [][]any{{}}
	}
	res := make([][]any, 0)
	current := make([]any, len(inputs))
	var recur func(int)
	recur = func(idx int) {
		if idx == len(inputs) {
			comb := make([]any, len(current))
			copy(comb, current)
			res = append(res, comb)
			return
		}
		for _, v := range inputs[idx] {
			current[idx] = v
			recur(idx + 1)
		}
	}
	recur(0)
	return res
}
