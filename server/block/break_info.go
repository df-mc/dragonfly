package block

import (
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

// Breakable represents a block that may be broken by a player in survival mode. Blocks not include are blocks
// such as bedrock.
type Breakable interface {
	// BreakInfo returns information of the block related to the breaking of it. Callers that execute the BreakHandler
	// must call BreakInfo on the concrete block value being broken, as handlers may need block state after the world
	// position has been cleared.
	BreakInfo() BreakInfo
}

// BreakContext carries the environmental and status-effect state that influences how quickly a block is
// broken. The zero value represents a player standing on the ground, out of water, without any relevant
// status effects or enchantments.
type BreakContext struct {
	// HasteLevel is the level of the Haste effect (0 if absent). Level 1 corresponds to Haste I.
	HasteLevel int
	// ConduitPowerLevel is the level of the Conduit Power effect (0 if absent). It grants a mining speed
	// boost equivalent to Haste; the two do not stack (the higher of the two is used).
	ConduitPowerLevel int
	// MiningFatigueLevel is the level of the Mining Fatigue effect (0 if absent).
	MiningFatigueLevel int
	// Underwater is true if the player's head is submerged in water, which slows mining fivefold unless
	// negated by AquaAffinity.
	Underwater bool
	// AquaAffinity is true if the player wears a helmet enchanted with Aqua Affinity, negating the
	// underwater mining penalty.
	AquaAffinity bool
	// AirBorne is true if the player is not on the ground, which slows mining by 5x.
	AirBorne bool
}

// BreakDuration returns the duration that breaking the block passed takes when being broken using the item
// passed, accounting for the status effects and environment described by ctx.
// See https://minecraft.wiki/w/Breaking#Calculation.
func BreakDuration(b world.Block, i item.Stack, ctx BreakContext) time.Duration {
	breakable, ok := b.(Breakable)
	if !ok {
		return math.MaxInt64
	}
	info := breakable.BreakInfo()
	if info.Hardness <= 0 {
		return 0
	}
	t, ok := i.Item().(item.Tool)
	if !ok {
		t = item.ToolNone{}
	}

	canHarvest := info.Harvestable(t)
	speed := 1.0
	if info.Effective(t) {
		speed = t.BaseMiningEfficiency(b)
		if !canHarvest {
			// A tool of the correct type but wrong tier (e.g. a wooden pickaxe on diamond ore) grants no
			// speed bonus in Bedrock Edition.
			speed = 1
		} else if e, ok := i.Enchantment(enchantment.Efficiency); ok {
			speed += enchantment.Efficiency.Addend(e.Level())
		}
	}

	// Haste and Conduit Power do not stack; the higher of the two is used. They boost both the mining speed
	// and the final destroy progress per tick.
	positive := max(ctx.HasteLevel, ctx.ConduitPowerLevel)
	if positive > 0 {
		speed *= 0.2*float64(positive) + 1
	}
	if ctx.MiningFatigueLevel > 0 {
		speed *= math.Pow(0.3, float64(ctx.MiningFatigueLevel))
	}
	if ctx.Underwater && !ctx.AquaAffinity {
		speed /= 5
	}
	if ctx.AirBorne {
		speed /= 5
	}

	damage := speed / info.Hardness
	if canHarvest {
		damage /= 30
	} else {
		damage /= 100
	}
	if positive > 0 {
		damage *= math.Pow(1.2, float64(positive))
	}
	if ctx.MiningFatigueLevel > 0 {
		damage *= math.Pow(0.7, float64(ctx.MiningFatigueLevel))
	}
	if damage >= 1 {
		// The block breaks within a single tick.
		return 0
	}
	return time.Duration(math.Ceil(1/damage)) * time.Second / 20
}

// BreaksInstantly checks if the block passed can be broken instantly using the item stack passed to break
// it, without any status effects.
func BreaksInstantly(b world.Block, i item.Stack) bool {
	return BreakDuration(b, i, BreakContext{}) == 0
}

// BreakInfo is a struct returned by every block. It holds information on block breaking related data, such as
// the tool type and tier required to break it.
type BreakInfo struct {
	// Hardness is the hardness of the block, which influences the speed with which the block may be mined.
	Hardness float64
	// Harvestable is a function called to check if the block is harvestable using the tool passed. If the
	// item used to break the block is not a tool, a tool.ToolNone is passed.
	Harvestable func(t item.Tool) bool
	// Effective is a function called to check if the block can be mined more effectively with the tool passed
	// than with an empty hand.
	Effective func(t item.Tool) bool
	// Drops is a function called to get the drops of the block if it is broken using the item passed.
	Drops func(t item.Tool, enchantments []item.Enchantment) []item.Stack
	// BreakHandler is called after the block has broken.
	BreakHandler func(pos cube.Pos, w *world.Tx, u item.User)
	// XPDrops is the range of XP a block can drop when broken.
	XPDrops XPDropRange
	// BlastResistance is the blast resistance of the block, which influences the block's ability to withstand an
	// explosive blast.
	BlastResistance float64
}

// newBreakInfo creates a BreakInfo struct with the properties passed. The XPDrops field is 0 by default. The blast
// resistance is set to the block's hardness*5 by default.
func newBreakInfo(hardness float64, harvestable func(item.Tool) bool, effective func(item.Tool) bool, drops func(item.Tool, []item.Enchantment) []item.Stack) BreakInfo {
	return BreakInfo{
		Hardness:        hardness,
		BlastResistance: hardness * 5,
		Harvestable:     harvestable,
		Effective:       effective,
		Drops:           drops,
	}
}

// withXPDropRange sets the XPDropRange field of the BreakInfo struct to the passed value.
func (b BreakInfo) withXPDropRange(min, max int) BreakInfo {
	b.XPDrops = XPDropRange{min, max}
	return b
}

// withBlastResistance sets the BlastResistance field of the BreakInfo struct to the passed value.
func (b BreakInfo) withBlastResistance(res float64) BreakInfo {
	b.BlastResistance = res
	return b
}

// withBreakHandler sets the BreakHandler field of the BreakInfo struct to the passed value.
func (b BreakInfo) withBreakHandler(handler func(pos cube.Pos, w *world.Tx, u item.User)) BreakInfo {
	b.BreakHandler = handler
	return b
}

// XPDropRange holds the min & max XP drop amounts of blocks.
type XPDropRange [2]int

// RandomValue returns a random XP value that falls within the drop range.
func (r XPDropRange) RandomValue() int {
	diff := r[1] - r[0]
	// Add one because it's a [r[0], r[1]] interval.
	return rand.IntN(diff+1) + r[0]
}

// pickaxeEffective is a convenience function for blocks that are effectively mined with a pickaxe.
var pickaxeEffective = func(t item.Tool) bool {
	return t.ToolType() == item.TypePickaxe
}

// axeEffective is a convenience function for blocks that are effectively mined with an axe.
var axeEffective = func(t item.Tool) bool {
	return t.ToolType() == item.TypeAxe
}

// shearsEffective is a convenience function for blocks that are effectively mined with shears.
var shearsEffective = func(t item.Tool) bool {
	return t.ToolType() == item.TypeShears
}

// swordEffective is a convenience function for blocks that are effectively mined with a sword.
var swordEffective = func(t item.Tool) bool {
	return t.ToolType() == item.TypeSword
}

// shovelEffective is a convenience function for blocks that are effectively mined with a shovel.
var shovelEffective = func(t item.Tool) bool {
	return t.ToolType() == item.TypeShovel
}

// hoeEffective is a convenience function for blocks that are effectively mined with a hoe.
var hoeEffective = func(t item.Tool) bool {
	return t.ToolType() == item.TypeHoe
}

// nothingEffective is a convenience function for blocks that cannot be mined efficiently with any tool.
var nothingEffective = func(item.Tool) bool {
	return false
}

// alwaysHarvestable is a convenience function for blocks that are harvestable using any item.
var alwaysHarvestable = func(t item.Tool) bool {
	return true
}

// neverHarvestable is a convenience function for blocks that are not harvestable by any item.
var neverHarvestable = func(t item.Tool) bool {
	return false
}

// pickaxeHarvestable is a convenience function for blocks that are harvestable using any kind of pickaxe.
var pickaxeHarvestable = pickaxeEffective

// simpleDrops returns a drops function that returns the items passed.
func simpleDrops(s ...item.Stack) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(item.Tool, []item.Enchantment) []item.Stack {
		return s
	}
}

// oneOf returns a drops function that returns one of each of the item types passed.
func oneOf(i ...world.Item) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(item.Tool, []item.Enchantment) []item.Stack {
		var s []item.Stack
		for _, it := range i {
			s = append(s, item.NewStack(it, 1))
		}
		return s
	}
}

// hasSilkTouch checks if an item has the silk touch enchantment.
func hasSilkTouch(enchantments []item.Enchantment) bool {
	return slices.IndexFunc(enchantments, func(i item.Enchantment) bool {
		return i.Type() == enchantment.SilkTouch
	}) != -1
}

// silkTouchOneOf returns a drop function that returns 1x of the silk touch drop when silk touch exists, or 1x of the
// normal drop when it does not.
func silkTouchOneOf(normal, silkTouch world.Item) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(silkTouch, 1)}
		}
		return []item.Stack{item.NewStack(normal, 1)}
	}
}

// silkTouchDrop returns a drop function that returns the silk touch drop when silk touch exists, or the
// normal drop when it does not.
func silkTouchDrop(normal, silkTouch item.Stack) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{silkTouch}
		}
		return []item.Stack{normal}
	}
}

// silkTouchOnlyDrop returns a drop function that returns the drop when silk touch exists.
func silkTouchOnlyDrop(it world.Item) func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(it, 1)}
		}
		return nil
	}
}

// fortuneLevel returns the level of the fortune enchantment in enchantments, or 0 if it isn't present.
func fortuneLevel(enchantments []item.Enchantment) int {
	index := slices.IndexFunc(enchantments, func(i item.Enchantment) bool {
		return i.Type() == enchantment.Fortune
	})
	if index == -1 {
		return 0
	}
	return enchantments[index].Level()
}

// fortuneOreCount computes the drop count for an ore after applying the Fortune ore multiplier to a given base
// drop count. The Fortune enchantment has a 2/(level + 2) chance of applying an integer bonus multiplier between
// 2x up to (level + 1)x to the drop count.
func fortuneOreCount(base int, enchantments []item.Enchantment) int {
	fortune := fortuneLevel(enchantments)
	if fortune == 0 || rand.IntN(fortune+2) < 2 {
		return base
	}
	multiplier := rand.IntN(fortune) + 2
	return base * multiplier
}

// fortuneDiscreteCount computes the drop count for a block with a discrete uniform distribution. A drop count is
// chosen with equal likelihood between min and max. Every level of Fortune will increase the max by one. The final
// drop count is then limited by the cap count.
func fortuneDiscreteCount(minCount, maxCount, capCount int, enchantments []item.Enchantment) int {
	fortune := fortuneLevel(enchantments)
	maxWithFortune := maxCount + fortune
	return min(capCount, rand.IntN(maxWithFortune-minCount+1)+minCount)
}

// fortuneBinomial computes the binomial distribution B(n=attempts, p=8/15) for crop seed drops.
func fortuneBinomial(attempts int) int {
	count := 0
	for range attempts {
		if rand.IntN(15) < 8 {
			count++
		}
	}
	return count
}

// oreDrops returns a drop function for ores that drop a single item, such as diamond. Silk touch tools will
// cause the ore block itself to always drop. Otherwise, a single item is dropped. The Fortune enchantment has a
// 2/(level + 2) chance of applying an integer bonus multiplier between 2x up to (level + 1)x to the drop count.
func oreDrops(drop, block world.Item) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(block, 1)}
		}
		return []item.Stack{item.NewStack(drop, fortuneOreCount(1, enchantments))}
	}
}

// multiOreDrops returns a drop function for ores that drop multiple items, such as copper. Silk touch tools will
// cause the ore block itself to always drop. Otherwise, a drop count is chosen with equal likelihood between min
// and max. The Fortune enchantment has a 2/(level + 2) chance of applying an integer bonus multiplier between 2x
// up to (level + 1)x to the drop count.
func multiOreDrops(drop, block world.Item, minCount, maxCount int) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(block, 1)}
		}
		baseCount := rand.IntN(maxCount-minCount+1) + minCount
		return []item.Stack{item.NewStack(drop, fortuneOreCount(baseCount, enchantments))}
	}
}

// discreteDrops returns a drop function for blocks with discrete uniform random drops, such as glowstone or melon
// blocks. Silk touch tools will cause the block itself to always drop. Otherwise, a drop count is chosen with equal
// likelihood between min and max. Every level of Fortune will increase the max by one. The final drop count is then
// limited by the cap count.
func discreteDrops(drop, block world.Item, minCount, maxCount, capCount int) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(block, 1)}
		}
		return []item.Stack{item.NewStack(drop, fortuneDiscreteCount(minCount, maxCount, capCount, enchantments))}
	}
}

// grassDrops returns a drop function for grass/fern blocks. Shears or silk touch tools will cause the grass block
// itself to always drop. Otherwise, there is a 12.5% chance of dropping a wheat seed. Every level of Fortune will
// increase the max drop count by 2, with each possible drop count being equally likely.
func grassDrops(grass world.Item) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears || hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(grass, 1)}
		}
		if rand.Float32() < 0.125 {
			count := 1
			if fortune := fortuneLevel(enchantments); fortune > 0 {
				count += rand.IntN(fortune*2 + 1)
			}
			return []item.Stack{item.NewStack(WheatSeeds{}, count)}
		}
		return nil
	}
}

// cropSeedDrops returns a drop function for wheat/beetroot seeds.
// Uses binomial distribution B(3+fortune, 8/15), seeds may not drop.
func cropSeedDrops(seed, crop world.Item, growth int) func(item.Tool, []item.Enchantment) []item.Stack {
	return func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if growth < 7 {
			return []item.Stack{item.NewStack(seed, 1)}
		}
		seedCount := fortuneBinomial(3 + fortuneLevel(enchantments))
		if seedCount == 0 {
			return []item.Stack{item.NewStack(crop, 1)}
		}
		return []item.Stack{item.NewStack(crop, 1), item.NewStack(seed, seedCount)}
	}
}

// breakBlock removes a block, shows breaking particles and drops the drops of
// the block as items.
func breakBlock(b world.Block, pos cube.Pos, tx *world.Tx) {
	breakBlockNoDrops(b, pos, tx)
	if breakable, ok := b.(Breakable); ok {
		for _, drop := range breakable.BreakInfo().Drops(item.ToolNone{}, nil) {
			dropItem(tx, drop, pos.Vec3Centre())
		}
	}
}

func breakBlockNoDrops(b world.Block, pos cube.Pos, tx *world.Tx) {
	// Clear the block first so neighbour-sensitive break handlers observe the post-break world state.
	tx.SetBlock(pos, nil, nil)
	if breakable, ok := b.(Breakable); ok {
		breakHandler := breakable.BreakInfo().BreakHandler
		if breakHandler != nil {
			breakHandler(pos, tx, nil)
		}
	}
	tx.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})
}
