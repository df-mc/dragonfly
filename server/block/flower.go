package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// Flower is a non-solid plant that occur in a variety of shapes and colours. They are primarily used for decoration
// and crafted into dyes.
type Flower struct {
	empty
	transparent

	// Type is the type of flower.
	Type FlowerType
}

// EntityInside ...
func (f Flower) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if f.Type == WitherRose() {
		if living, ok := e.(interface {
			AddEffect(effect.Effect)
		}); ok {
			living.AddEffect(effect.New(effect.Wither, 1, 2*time.Second))
		}
	}
}

// BoneMeal ...
func (f Flower) BoneMeal(pos cube.Pos, tx *world.Tx) (success bool) {
	if f.Type == WitherRose() {
		return
	}

	for i := 0; i < 8; i++ {
		p := pos.Add(cube.Pos{rand.IntN(7) - 3, rand.IntN(3) - 1, rand.IntN(7) - 3})
		if _, ok := tx.Block(p).(Air); !ok {
			continue
		}
		if _, ok := tx.Block(p.Side(cube.FaceDown)).(Grass); !ok {
			continue
		}
		flowerType := f.Type
		if rand.Float64() < 0.1 {
			if f.Type == Dandelion() {
				flowerType = Poppy()
			} else if f.Type == Poppy() {
				flowerType = Dandelion()
			}
		}
		tx.SetBlock(p, Flower{Type: flowerType}, nil)
		success = true
	}
	return
}

// NeighbourUpdateTick ...
func (f Flower) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(f, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(f, pos, tx)
	}
}

// UseOnBlock ...
func (f Flower) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, f)
	if !used || !supportsVegetation(f, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, f, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (Flower) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (f Flower) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (f Flower) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(f))
}

// CompostChance ...
func (Flower) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (f Flower) EncodeItem() (name string, meta int16) {
	return "minecraft:" + f.Type.String(), 0
}

// EncodeBlock ...
func (f Flower) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + f.Type.String(), nil
}

// allFlowers ...
func allFlowers() (b []world.Block) {
	for _, f := range FlowerTypes() {
		b = append(b, Flower{Type: f})
	}
	return
}
