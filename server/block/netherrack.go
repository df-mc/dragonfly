package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

// Netherrack is a block found in The Nether.
type Netherrack struct {
	solid
	bassDrum
}

// SoilFor ...
func (n Netherrack) SoilFor(block world.Block) bool {
	flower, ok := block.(Flower)
	return ok && flower.Type == WitherRose()
}

// BoneMeal ...
func (n Netherrack) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	var crimson, warped bool
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			for z := -1; z <= 1; z++ {
				// Vanilla searches the whole cube around the netherrack apart from the block right below it.
				if x == 0 && z == 0 && y <= 0 {
					continue
				}
				if nylium, ok := tx.Block(pos.Add(cube.Pos{x, y, z})).(Nylium); ok {
					warped = warped || nylium.Warped
					crimson = crimson || !nylium.Warped
				}
			}
		}
	}
	if !crimson && !warped {
		return item.BoneMealResultNone
	}
	nylium := Nylium{Warped: warped}
	if crimson && warped {
		nylium.Warped = rand.IntN(2) == 0
	}
	tx.AddParticle(pos.Vec3(), particle.BoneMeal{})
	tx.SetBlock(pos, nylium, nil)
	return item.BoneMealResultSmall
}

// AddsBoneMealParticle ...
func (Netherrack) AddsBoneMealParticle() {}

// BreakInfo ...
func (n Netherrack) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, pickaxeHarvestable, pickaxeEffective, oneOf(n))
}

// SmeltInfo ...
func (Netherrack) SmeltInfo() item.SmeltInfo {
	return newSmeltInfo(item.NewStack(item.NetherBrick{}, 1), 0.1)
}

// EncodeItem ...
func (Netherrack) EncodeItem() (name string, meta int16) {
	return "minecraft:netherrack", 0
}

// EncodeBlock ...
func (Netherrack) EncodeBlock() (string, map[string]any) {
	return "minecraft:netherrack", nil
}
