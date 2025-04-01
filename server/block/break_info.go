package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"math"
	"math/rand/v2"
	"slices"
	"time"
)

// Breakable represents a block that may be broken by a player in survival mode. Blocks not include are blocks
// such as bedrock.
type Breakable interface {
	// BreakInfo returns information of the block related to the breaking of it.
	BreakInfo() BreakInfo
}

// BreakDuration returns the base duration that breaking the block passed takes when being broken using the
// item passed.
func BreakDuration(b world.Block, i item.Stack) time.Duration {
	breakable, ok := b.(Breakable)
	if !ok {
		return math.MaxInt64
	}
	t, ok := i.Item().(item.Tool)
	if !ok {
		t = item.ToolNone{}
	}
	info := breakable.BreakInfo()

	breakTime := info.Hardness * 5
	if info.Harvestable(t) {
		breakTime = info.Hardness * 1.5
	}
	if info.Effective(t) {
		eff := t.BaseMiningEfficiency(b)
		if e, ok := i.Enchantment(enchantment.Efficiency); ok {
			eff += enchantment.Efficiency.Addend(e.Level())
		}
		breakTime /= eff
	}
	// TODO: Account for haste etc here.
	timeInTicksAccurate := math.Round(breakTime/0.05) * 0.05

	return (time.Duration(math.Round(timeInTicksAccurate*20)) * time.Second) / 20
}

// BreaksInstantly checks if the block passed can be broken instantly using the item stack passed to break
// it.
func BreaksInstantly(b world.Block, i item.Stack) bool {
	breakable, ok := b.(Breakable)
	if !ok {
		return false
	}
	hardness := breakable.BreakInfo().Hardness
	if hardness == 0 {
		return true
	}
	t, ok := i.Item().(item.Tool)
	if !ok || !breakable.BreakInfo().Effective(t) {
		return false
	}

	// TODO: Account for haste etc here.
	efficiencyVal := 0.0
	if e, ok := i.Enchantment(enchantment.Efficiency); ok {
		efficiencyVal += enchantment.Efficiency.Addend(e.Level())
	}
	hasteVal := 0.0
	return (t.BaseMiningEfficiency(b)+efficiencyVal)*hasteVal >= hardness*30
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
	if breakable, ok := b.(Breakable); ok {
		breakHandler := breakable.BreakInfo().BreakHandler
		if breakHandler != nil {
			breakHandler(pos, tx, nil)
		}
	}
	tx.SetBlock(pos, nil, nil)
	tx.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})
}
