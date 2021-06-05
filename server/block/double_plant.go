package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// DoublePlant is a two block high plant consisting of an upper and lower part.
type DoublePlant struct {
	transparent
	empty

	// UpperPart is set if the plant is the upper part.
	UpperPart bool
	// Type is the type of the double plant.
	Type DoublePlantType
}

// ReplaceableBy ...
func (d DoublePlant) ReplaceableBy(world.Block) bool {
	return d.Type == TallGrass() || d.Type == LargeFern()
}

// BoneMeal ...
func (d DoublePlant) BoneMeal(pos cube.Pos, w *world.World) bool {
	switch d.Type {
	case TallGrass(), LargeFern():
		return false
	default:
		itemEntity := entity.NewItem(item.NewStack(d, 1), pos.Vec3Centre())
		itemEntity.SetVelocity(mgl64.Vec3{rand.Float64()*0.2 - 0.1, 0.2, rand.Float64()*0.2 - 0.1})
		w.AddEntity(itemEntity)
		return true
	}
}

// NeighbourUpdateTick ...
func (d DoublePlant) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d.UpperPart {
		if bottom, ok := w.Block(pos.Side(cube.FaceDown)).(DoublePlant); !ok || bottom.Type != d.Type || bottom.UpperPart {
			w.BreakBlock(pos)
		}
		return
	}
	if upper, ok := w.Block(pos.Side(cube.FaceUp)).(DoublePlant); !ok || upper.Type != d.Type || !upper.UpperPart {
		w.BreakBlock(pos)
		return
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Grass); !ok {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(Dirt); !ok {
			w.BreakBlock(pos)
		}
	}
}

// UseOnBlock ...
func (d DoublePlant) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	if !replaceableWith(w, pos.Side(cube.FaceUp), d) {
		return false
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Grass); !ok {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(Dirt); !ok {
			return false
		}
	}

	place(w, pos, d, user, ctx)
	place(w, pos.Side(cube.FaceUp), DoublePlant{Type: d.Type, UpperPart: true}, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (d DoublePlant) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(t tool.Tool) []item.Stack {
		switch d.Type {
		case TallGrass(), LargeFern():
			if t.ToolType() == tool.TypeShears { //TODO: Silk Touch
				return []item.Stack{item.NewStack(d, 1)}
			}
			if rand.Float32() > 0.57 {
				return []item.Stack{item.NewStack(WheatSeeds{}, 1)}
			}
			return []item.Stack{}
		default:
			return []item.Stack{item.NewStack(d, 1)}
		}
	})
}

// HasLiquidDrops ...
func (d DoublePlant) HasLiquidDrops() bool {
	return true
}

// EncodeItem ...
func (d DoublePlant) EncodeItem() (name string, meta int16) {
	return "minecraft:double_plant", int16(d.Type.Uint8())
}

// EncodeBlock ...
func (d DoublePlant) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:double_plant", map[string]interface{}{"double_flower_type": d.Type.String(), "upper_block_bit": d.UpperPart}
}

// allDoublePlants ...
func allDoublePlants() (b []world.Block) {
	for _, d := range DoublePlantTypes() {
		b = append(b, DoublePlant{Type: d, UpperPart: true})
		b = append(b, DoublePlant{Type: d, UpperPart: false})
	}
	return
}
