package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
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
	Wood WoodType
	// Facing is the direction the trapdoor is facing.
	Facing cube.Direction
	// Open is whether the trapdoor is open.
	Open bool
	// Top is whether the trapdoor occupies the top or bottom part of a block.
	Top bool
}

// FlammabilityInfo ...
func (t WoodTrapdoor) FlammabilityInfo() FlammabilityInfo {
	if !t.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(0, 0, true)
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
func (t WoodTrapdoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) bool {
	t.Open = !t.Open
	w.SetBlock(pos, t, nil)
	w.PlaySound(pos.Vec3Centre(), sound.Door{})
	return true
}

// BreakInfo ...
func (t WoodTrapdoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, axeEffective, oneOf(t))
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
func (t WoodTrapdoor) EncodeItem() (name string, meta int16) {
	if t.Wood == OakWood() {
		return "minecraft:trapdoor", 0
	}
	return "minecraft:" + t.Wood.String() + "_trapdoor", 0
}

// EncodeBlock ...
func (t WoodTrapdoor) EncodeBlock() (name string, properties map[string]any) {
	if t.Wood == OakWood() {
		return "minecraft:trapdoor", map[string]any{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
	}
	return "minecraft:" + t.Wood.String() + "_trapdoor", map[string]any{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
}

// allTrapdoors returns a list of all trapdoor types
func allTrapdoors() (trapdoors []world.Block) {
	for _, w := range WoodTypes() {
		for i := cube.Direction(0); i <= 3; i++ {
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: false, Top: false})
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: false, Top: true})
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: true, Top: true})
			trapdoors = append(trapdoors, WoodTrapdoor{Wood: w, Facing: i, Open: true, Top: false})
		}
	}
	return
}
