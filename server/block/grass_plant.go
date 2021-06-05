package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/grass"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// GrassPlant is a transparent plant block which can be used to obtain seeds and as decoration.
type GrassPlant struct {
	replaceable
	transparent
	empty

	// Type is the type of grass that the plant represents.
	Type grass.Grass
}

// FlammabilityInfo ...
func (g GrassPlant) FlammabilityInfo() FlammabilityInfo {
	if g.Type == grass.NetherSprouts() {
		return newFlammabilityInfo(60, 0, true)
	}
	return newFlammabilityInfo(60, 100, true)
}

// BreakInfo ...
func (g GrassPlant) BreakInfo() BreakInfo {
	// TODO: Silk touch.
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(t tool.Tool) []item.Stack {
		if g.Type == grass.NetherSprouts() {
			return []item.Stack{item.NewStack(g, 1)}
		}
		if rand.Float32() > 0.57 {
			return []item.Stack{item.NewStack(WheatSeeds{}, 1)}
		}
		return []item.Stack{}
	})
}

// BoneMeal attempts to affect the block using a bone meal item.
func (g GrassPlant) BoneMeal(pos cube.Pos, w *world.World) bool {
	switch g.Type {
	case grass.SmallGrass():
		upper := DoublePlant{Type: TallGrass(), UpperPart: true}
		if replaceableWith(w, pos.Side(cube.FaceUp), upper) {
			w.SetBlock(pos, DoublePlant{Type: TallGrass()})
			w.SetBlock(pos.Side(cube.FaceUp), upper)
			return true
		}
	case grass.Fern():
		upper := DoublePlant{Type: LargeFern(), UpperPart: true}
		if replaceableWith(w, pos.Side(cube.FaceUp), upper) {
			w.SetBlock(pos, DoublePlant{Type: LargeFern()})
			w.SetBlock(pos.Side(cube.FaceUp), upper)
			return true
		}
	}
	return false
}

// NeighbourUpdateTick ...
func (g GrassPlant) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Grass); !ok {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(Dirt); !ok {
			w.BreakBlock(pos)
		}
	}
}

// HasLiquidDrops ...
func (g GrassPlant) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (g GrassPlant) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, g)
	if !used {
		return false
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Grass); !ok {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(Dirt); !ok {
			return false
		}
	}

	place(w, pos, g, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (g GrassPlant) EncodeItem() (name string, meta int16) {
	switch g.Type {
	case grass.SmallGrass():
		return "minecraft:tallgrass", 1
	case grass.Fern():
		return "minecraft:tallgrass", 2
	case grass.NetherSprouts():
		return "minecraft:nether_sprouts", 0
	}
	panic("should never happen")
}

// EncodeBlock ...
func (g GrassPlant) EncodeBlock() (name string, properties map[string]interface{}) {
	switch g.Type {
	case grass.SmallGrass():
		return "minecraft:tallgrass", map[string]interface{}{"tall_grass_type": "tall"}
	case grass.Fern():
		return "minecraft:tallgrass", map[string]interface{}{"tall_grass_type": "fern"}
	case grass.NetherSprouts():
		return "minecraft:nether_sprouts", map[string]interface{}{}
	}
	panic("should never happen")
}

// allGrassPlants ...
func allGrassPlants() (grasses []world.Block) {
	for _, g := range grass.All() {
		grasses = append(grasses, GrassPlant{Type: g})
	}
	return
}
