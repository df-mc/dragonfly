package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/tool"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"math"
	"time"
)

// BreakDuration returns the base duration that breaking the block passed takes when being broken using the
// item passed.
func BreakDuration(b world.Block, i item.Stack) time.Duration {
	breakable, ok := b.(Breakable)
	if !ok {
		return math.MaxInt64
	}
	t, ok := i.Item().(tool.Tool)
	if !ok {
		t = tool.None{}
	}
	info := breakable.BreakInfo()

	breakTime := info.Hardness * 5
	if info.Harvestable(t) {
		breakTime = info.Hardness * 1.5
	}
	if info.Effective(t) {
		breakTime /= t.BaseMiningEfficiency()
	}
	timeInTicksAccurate := math.Round(breakTime/0.05) * 0.05

	return (time.Duration(math.Round(timeInTicksAccurate*20)) * time.Second) / 20
}

// Breakable represents a block that may be broken by a player in survival mode. Blocks not include are blocks
// such as bedrock.
type Breakable interface {
	// BreakInfo returns information of the block related to the breaking of it.
	BreakInfo() BreakInfo
}

// BreakInfo is a struct returned by every block. It holds information on block breaking related data, such as
// the tool type and tier required to break it.
type BreakInfo struct {
	// Hardness is the hardness of the block, which influences the speed with which the block may be mined.
	Hardness float64
	// Harvestable is a function called to check if the block is harvestable using the tool passed. If the
	// item used to break the block is not a tool, a tool.None is passed.
	Harvestable func(t tool.Tool) bool
	// Effective is a function called to check if the block can be mined more effectively with the tool passed
	// than with an empty hand.
	Effective func(t tool.Tool) bool
	// Drops is a function called to get the drops of the block if it is broken using the tool passed. If the
	// item used to break the block is not a tool, a tool.None is passed.
	Drops func(t tool.Tool) []item.Stack
}

// pickaxeEffective is a convenience function for blocks that are effectively mined with a pickaxe.
var pickaxeEffective = func(t tool.Tool) bool {
	return t.ToolType() == tool.TypePickaxe
}

// axeEffective is a convenience function for blocks that are effectively mined with an axe.
var axeEffective = func(t tool.Tool) bool {
	return t.ToolType() == tool.TypeAxe
}

// shovelEffective is a convenience function for blocks that are effectively mined with an axe.
var shovelEffective = func(t tool.Tool) bool {
	return t.ToolType() == tool.TypeShovel
}

// alwaysHarvestable is a convenience function for blocks that are harvestable using any item.
var alwaysHarvestable = func(t tool.Tool) bool {
	return true
}

// pickaxeHarvestable is a convenience function for blocks that are harvestable using any kind of pickaxe.
var pickaxeHarvestable = pickaxeEffective

// simpleDrops returns a drops function that returns the items passed.
func simpleDrops(s ...item.Stack) func(t tool.Tool) []item.Stack {
	return func(t tool.Tool) []item.Stack {
		return s
	}
}
