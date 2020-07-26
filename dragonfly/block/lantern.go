package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Lantern is a light emitting block.
type Lantern struct {
	noNBT
	transparent

	// Hanging determines if a lantern is hanging off a block.
	Hanging bool
	// Soul determines whether it is a normal lantern or soul lantern.
	Soul bool
}

func (l Lantern) Model() world.BlockModel {
	return model.Lantern{}
}

// NeighbourUpdateTick ...
func (l Lantern) NeighbourUpdateTick(pos, changedNeighbour world.BlockPos, w *world.World) {
	if l.Hanging {
		if _, air := w.Block(pos.Side(world.FaceUp)).(Air); air {
			w.BreakBlock(pos)
		}
	} else if _, air := w.Block(pos.Side(world.FaceDown)).(Air); air {
		w.BreakBlock(pos)
	}
}

// LightEmissionLevel ...
func (l Lantern) LightEmissionLevel() uint8 {
	if l.Soul {
		return 10
	}
	return 15
}

// UseOnBlock ...
func (l Lantern) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, l)
	if !used {
		return false
	}
	l.Hanging = face == world.FaceDown

	place(w, pos, l, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (l Lantern) HasLiquidDrops() bool {
	return true
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
func (l Lantern) EncodeItem() (id int32, meta int16) {
	if l.Soul {
		return -269, 0
	}
	return -208, 0
}

// EncodeBlock ...
func (l Lantern) EncodeBlock() (name string, properties map[string]interface{}) {
	if l.Soul {
		return "minecraft:soul_Lantern", map[string]interface{}{"hanging": l.Hanging}
	}
	return "minecraft:lantern", map[string]interface{}{"hanging": l.Hanging}
}

// Hash ...
func (l Lantern) Hash() uint64 {
	return hashLantern | (uint64(boolByte(l.Hanging)) << 32) | (uint64(boolByte(l.Soul)) << 33)
}
