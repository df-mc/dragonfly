package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
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
func (f Flower) EntityInside(_ cube.Pos, _ *world.World, e world.Entity) {
	if f.Type == WitherRose() {
		if living, ok := e.(interface {
			AddEffect(effect.Effect)
		}); ok {
			living.AddEffect(effect.New(effect.Wither{}, 1, 2*time.Second))
		}
	}
}

// BoneMeal ...
func (f Flower) BoneMeal(pos cube.Pos, w *world.World) (success bool) {
	if f.Type == WitherRose() {
		return
	}

	for i := 0; i < 8; i++ {
		p := pos.Add(cube.Pos{rand.Intn(7) - 3, rand.Intn(3) - 1, rand.Intn(7) - 3})
		if _, ok := w.Block(p).(Air); !ok {
			continue
		}
		if _, ok := w.Block(p.Side(cube.FaceDown)).(Grass); !ok {
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
		w.SetBlock(p, Flower{Type: flowerType}, nil)
		success = true
	}
	return
}

// NeighbourUpdateTick ...
func (f Flower) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !supportsVegetation(f, w.Block(pos.Side(cube.FaceDown))) {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: f})
	}
}

// UseOnBlock ...
func (f Flower) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, f)
	if !used {
		return false
	}
	if !supportsVegetation(f, w.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(w, pos, f, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (Flower) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (Flower) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 100, false)
}

// BreakInfo ...
func (f Flower) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(f))
}

// EncodeItem ...
func (f Flower) EncodeItem() (name string, meta int16) {
	if f.Type == Dandelion() {
		return "minecraft:yellow_flower", 0
	} else if f.Type == WitherRose() {
		return "minecraft:wither_rose", 0
	}
	return "minecraft:red_flower", int16(f.Type.Uint8() - 1)
}

// EncodeBlock ...
func (f Flower) EncodeBlock() (string, map[string]any) {
	if f.Type == Dandelion() {
		return "minecraft:yellow_flower", nil
	} else if f.Type == WitherRose() {
		return "minecraft:wither_rose", nil
	}
	return "minecraft:red_flower", map[string]any{"flower_type": f.Type.String()}
}

// allFlowers ...
func allFlowers() (b []world.Block) {
	for _, f := range FlowerTypes() {
		b = append(b, Flower{Type: f})
	}
	return
}
