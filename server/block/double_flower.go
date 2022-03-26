package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// DoubleFlower is a two block high flower consisting of an upper and lower part.
type DoubleFlower struct {
	transparent
	empty

	// UpperPart is set if the plant is the upper part.
	UpperPart bool
	// Type is the type of the double plant.
	Type DoubleFlowerType
}

// FlammabilityInfo ...
func (d DoubleFlower) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

// BoneMeal ...
func (d DoubleFlower) BoneMeal(pos cube.Pos, w *world.World) bool {
	itemEntity := entity.NewItem(item.NewStack(d, 1), pos.Vec3Centre())
	itemEntity.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
	w.AddEntity(itemEntity)
	return true
}

// NeighbourUpdateTick ...
func (d DoubleFlower) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d.UpperPart {
		if bottom, ok := w.Block(pos.Side(cube.FaceDown)).(DoubleFlower); !ok || bottom.Type != d.Type || bottom.UpperPart {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
		}
		return
	}
	if upper, ok := w.Block(pos.Side(cube.FaceUp)).(DoubleFlower); !ok || upper.Type != d.Type || !upper.UpperPart {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
		return
	}
	if !supportsVegetation(d, w.Block(pos.Side(cube.FaceDown))) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	}
}

// UseOnBlock ...
func (d DoubleFlower) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	if !replaceableWith(w, pos.Side(cube.FaceUp), d) {
		return false
	}
	if !supportsVegetation(d, w.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(w, pos, d, user, ctx)
	place(w, pos.Side(cube.FaceUp), DoubleFlower{Type: d.Type, UpperPart: true}, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (d DoubleFlower) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(d))
}

// HasLiquidDrops ...
func (d DoubleFlower) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (d DoubleFlower) EncodeItem() (name string, meta int16) {
	return "minecraft:double_plant", int16(d.Type.Uint8())
}

// EncodeBlock ...
func (d DoubleFlower) EncodeBlock() (string, map[string]any) {
	return "minecraft:double_plant", map[string]any{"double_plant_type": d.Type.String(), "upper_block_bit": d.UpperPart}
}

// allDoubleFlowers ...
func allDoubleFlowers() (b []world.Block) {
	for _, d := range DoubleFlowerTypes() {
		b = append(b, DoubleFlower{Type: d, UpperPart: true})
		b = append(b, DoubleFlower{Type: d, UpperPart: false})
	}
	return
}
