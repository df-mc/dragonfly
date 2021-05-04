package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Lantern is a light emitting block.
type Lantern struct {
	transparent

	// Hanging determines if a lantern is hanging off a block.
	Hanging bool
	// Type of fire lighting the lantern.
	Type FireType
}

// Model ...
func (l Lantern) Model() world.BlockModel {
	return model.Lantern{Hanging: l.Hanging}
}

// NeighbourUpdateTick ...
func (l Lantern) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if l.Hanging {
		up := pos.Side(cube.FaceUp)
		if !w.Block(up).Model().FaceSolid(up, cube.FaceDown, w) {
			w.BreakBlockWithoutParticles(pos)
		}
	} else {
		down := pos.Side(cube.FaceDown)
		if !w.Block(down).Model().FaceSolid(down, cube.FaceUp, w) {
			w.BreakBlockWithoutParticles(pos)
		}
	}
}

// LightEmissionLevel ...
func (l Lantern) LightEmissionLevel() uint8 {
	return l.Type.LightLevel()
}

// UseOnBlock ...
func (l Lantern) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, l)
	if !used {
		return false
	}
	if face == cube.FaceDown {
		upPos := pos.Side(cube.FaceUp)
		if !w.Block(upPos).Model().FaceSolid(upPos, cube.FaceDown, w) {
			face = cube.FaceUp
		}
	}
	if face != cube.FaceDown {
		downPos := pos.Side(cube.FaceDown)
		if !w.Block(downPos).Model().FaceSolid(downPos, cube.FaceUp, w) {
			return false
		}
	}
	l.Hanging = face == cube.FaceDown

	place(w, pos, l, user, ctx)
	return placed(ctx)
}

// CanDisplace ...
func (l Lantern) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (l Lantern) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (l Lantern) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3.5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(l, 1)),
	}
}

// EncodeItem ...
func (l Lantern) EncodeItem() (name string, meta int16) {
	switch l.Type {
	case NormalFire():
		return "minecraft:lantern", 0
	case SoulFire():
		return "minecraft:soul_lantern", 0
	}
	panic("invalid fire type")
}

// EncodeBlock ...
func (l Lantern) EncodeBlock() (name string, properties map[string]interface{}) {
	switch l.Type {
	case NormalFire():
		return "minecraft:lantern", map[string]interface{}{"hanging": l.Hanging}
	case SoulFire():
		return "minecraft:soul_lantern", map[string]interface{}{"hanging": l.Hanging}
	}
	panic("invalid fire type")
}

// allLanterns ...
func allLanterns() (lanterns []world.Block) {
	for _, f := range FireTypes() {
		lanterns = append(lanterns, Lantern{Hanging: false, Type: f})
		lanterns = append(lanterns, Lantern{Hanging: true, Type: f})
	}
	return
}
