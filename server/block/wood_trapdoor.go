package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/block/wood"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// WoodTrapdoor is a block that can be used as an openable 1x1 barrier.
type WoodTrapdoor struct {
	transparent
	bass

	// Wood is the type of wood of the trapdoor. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Facing is the direction the trapdoor is facing.
	Facing cube.Direction
	// Open is whether or not the trapdoor is open.
	Open bool
	// Top is whether the trapdoor occupies the top or bottom part of a block.
	Top bool
}

// FlammabilityInfo ...
func (t WoodTrapdoor) FlammabilityInfo() FlammabilityInfo {
	if !t.Wood.Flammable() {
		return FlammabilityInfo{}
	}
	return FlammabilityInfo{LavaFlammable: true}
}

// Model ...
func (t WoodTrapdoor) Model() world.BlockModel {
	return model.Trapdoor{Facing: t.Facing, Top: t.Top, Open: t.Open}
}

// UseOnBlock handles the directional placing of trapdoors and makes sure they are properly placed upside down
// when needed.
func (t WoodTrapdoor) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, t)
	if !used {
		return false
	}
	t.Facing = user.Facing().Opposite()
	t.Top = (clickPos.Y() > 0.5 && face != cube.FaceUp) || face == cube.FaceDown

	place(w, pos, t, user, ctx)
	return placed(ctx)
}

// Activate ...
func (t WoodTrapdoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) {
	t.Open = !t.Open
	w.PlaceBlock(pos, t)
	w.PlaySound(pos.Vec3Centre(), sound.Door{})
}

// BreakInfo ...
func (t WoodTrapdoor) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(t, 1)),
	}
}

// CanDisplace ...
func (t WoodTrapdoor) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (t WoodTrapdoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (t WoodTrapdoor) EncodeItem() (id int32, name string, meta int16) {
	switch t.Wood {
	case wood.Oak():
		return 96, "minecraft:trapdoor", 0
	case wood.Spruce():
		return -149, "minecraft:spruce_trapdoor", 0
	case wood.Birch():
		return -146, "minecraft:birch_trapdoor", 0
	case wood.Jungle():
		return -148, "minecraft:jungle_trapdoor", 0
	case wood.Acacia():
		return -145, "minecraft:acacia_trapdoor", 0
	case wood.DarkOak():
		return -147, "minecraft:dark_oak_trapdoor", 0
	case wood.Crimson():
		return -246, "minecraft:crimson_trapdoor", 0
	case wood.Warped():
		return -247, "minecraft:warped_trapdoor", 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (t WoodTrapdoor) EncodeBlock() (name string, properties map[string]interface{}) {
	switch t.Wood {
	case wood.Oak():
		return "minecraft:trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	default:
		return "minecraft:" + t.Wood.String() + "_trapdoor", map[string]interface{}{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	}
}

// allTrapdoors returns a list of all trapdoor types
func allTrapdoors() (trapdoors []world.Block) {
	for _, w := range wood.All() {
		for i := cube.Direction(0); i <= 3; i++ {
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: false, Top: false})
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: false, Top: true})
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: true, Top: true})
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: true, Top: false})
		}
	}
	return
}
