package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Lantern is a light emitting block
type Lantern struct {
	noNBT

	// Hanging determines if a lantern is hanging off a block
	Hanging bool
	// Soul determines whether it is a normal lantern or soul lantern
	Soul bool
}

func (n Lantern) LightEmissionLevel() uint8 {
	if n.Soul {
		return 10
	}
	return 15
}

func (n Lantern) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, n)
	if !used {
		return false
	}
	n.Hanging = face == world.FaceDown

	place(w, pos, n, user, ctx)
	return placed(ctx)
}

func (n Lantern) HasLiquidDrops() bool {
	return true
}

func (n Lantern) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3.5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(n, 1)),
	}
}

func (n Lantern) EncodeItem() (id int32, meta int16) {
	if n.Soul {
		return -269, 0
	}
	return -208, 0
}

func (n Lantern) EncodeBlock() (name string, properties map[string]interface{}) {
	if n.Soul {
		return "minecraft:soul_Lantern", map[string]interface{}{"hanging": n.Hanging}
	}
	return "minecraft:lantern", map[string]interface{}{"hanging": n.Hanging}
}

func (n Lantern) Hash() uint64 {
	return hashLantern | (uint64(boolByte(n.Hanging)) << 32) | (uint64(boolByte(n.Soul)) << 33)
}
