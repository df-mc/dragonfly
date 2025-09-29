package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Cake is an edible block.
type Cake struct {
	transparent
	sourceWaterDisplacer

	// Bites is the amount of bites taken out of the cake.
	Bites int
}

func (c Cake) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (c Cake) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}

	if _, air := tx.Block(pos.Side(cube.FaceDown)).(Air); air {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

func (c Cake) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, air := tx.Block(pos.Side(cube.FaceDown)).(Air); air {
		breakBlock(c, pos, tx)
	}
}

func (c Cake) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if i, ok := u.(interface {
		Saturate(food int, saturation float64)
	}); ok {
		i.Saturate(2, 0.4)
		tx.PlaySound(u.Position().Add(mgl64.Vec3{0, 1.5}), sound.Burp{})
		c.Bites++
		if c.Bites > 6 {
			tx.SetBlock(pos, nil, nil)
			return true
		}
		tx.SetBlock(pos, c, nil)
		return true
	}
	return false
}

func (c Cake) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, neverHarvestable, nothingEffective, simpleDrops())
}

func (c Cake) EncodeItem() (name string, meta int16) {
	return "minecraft:cake", 0
}

func (c Cake) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:cake", map[string]any{"bite_counter": int32(c.Bites)}
}

func (c Cake) Model() world.BlockModel {
	return model.Cake{Bites: c.Bites}
}

func allCake() (cake []world.Block) {
	for bites := 0; bites < 7; bites++ {
		cake = append(cake, Cake{Bites: bites})
	}
	return
}
