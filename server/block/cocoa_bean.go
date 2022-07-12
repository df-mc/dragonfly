package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// CocoaBean is a crop block found in jungle biomes.
type CocoaBean struct {
	transparent

	// Facing is the direction from the cocoa bean to the log.
	Facing cube.Direction
	// Age is the stage of the cocoa bean's growth. 2 is fully grown.
	Age int
}

// BoneMeal ...
func (c CocoaBean) BoneMeal(pos cube.Pos, w *world.World) bool {
	if c.Age == 2 {
		return false
	}
	c.Age++
	w.SetBlock(pos, c, nil)
	return true
}

// HasLiquidDrops ...
func (c CocoaBean) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (c CocoaBean) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	var woodType WoodType
	switch b := w.Block(pos.Side(c.Facing.Face())).(type) {
	case Log:
		woodType = b.Wood
	case Wood:
		woodType = b.Wood
	}
	if woodType != JungleWood() {
		w.SetBlock(pos, nil, nil)
	}
}

// UseOnBlock ...
func (c CocoaBean) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}

	if face == cube.FaceUp || face == cube.FaceDown {
		return false
	}

	var woodType WoodType
	oppositePos := pos.Side(face.Opposite())
	if log, ok := w.Block(oppositePos).(Log); ok {
		woodType = log.Wood
	} else if wood, ok := w.Block(oppositePos).(Wood); ok {
		woodType = wood.Wood
	}
	if woodType == JungleWood() {
		c.Facing = face.Opposite().Direction()
		ctx.IgnoreBBox = true

		place(w, pos, c, user, ctx)
		return placed(ctx)
	}

	return false
}

// RandomTick ...
func (c CocoaBean) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if c.Age < 2 && r.Intn(5) == 0 {
		c.Age++
		w.SetBlock(pos, c, nil)
	}
}

// BreakInfo ...
func (c CocoaBean) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, axeEffective, func(item.Tool, []item.Enchantment) []item.Stack {
		if c.Age == 2 {
			return []item.Stack{item.NewStack(c, rand.Intn(2)+2)}
		}
		return []item.Stack{item.NewStack(c, 1)}
	})
}

// EncodeItem ...
func (c CocoaBean) EncodeItem() (name string, meta int16) {
	return "minecraft:cocoa_beans", 0
}

// EncodeBlock ...
func (c CocoaBean) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:cocoa", map[string]any{"age": int32(c.Age), "direction": int32(horizontalDirection(c.Facing))}
}

// Model ...
func (c CocoaBean) Model() world.BlockModel {
	return model.CocoaBean{Facing: c.Facing, Age: c.Age}
}

// allCocoaBeans ...
func allCocoaBeans() (cocoa []world.Block) {
	for i := cube.Direction(0); i <= 3; i++ {
		cocoa = append(cocoa, CocoaBean{Facing: i, Age: 0})
		cocoa = append(cocoa, CocoaBean{Facing: i, Age: 1})
		cocoa = append(cocoa, CocoaBean{Facing: i, Age: 2})
	}
	return
}
