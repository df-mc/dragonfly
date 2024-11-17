package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
)

// CopperTrapdoor is a block that can be used as an openable 1x1 barrier.
type CopperTrapdoor struct {
	transparent
	bass
	sourceWaterDisplacer

	// Oxidation is the level of oxidation of the copper trapdoor.
	Oxidation OxidationType
	// Waxed bool is whether the copper trapdoor has been waxed with honeycomb.
	Waxed bool
	// Facing is the direction the trapdoor is facing.
	Facing cube.Direction
	// Open is whether the trapdoor is open.
	Open bool
	// Top is whether the trapdoor occupies the top or bottom part of a block.
	Top bool
}

// Model ...
func (t CopperTrapdoor) Model() world.BlockModel {
	return model.Trapdoor{Facing: t.Facing, Top: t.Top, Open: t.Open}
}

// UseOnBlock handles the directional placing of trapdoors and makes sure they are properly placed upside down
// when needed.
func (t CopperTrapdoor) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, t)
	if !used {
		return false
	}
	t.Facing = user.Rotation().Direction().Opposite()
	t.Top = (clickPos.Y() > 0.5 && face != cube.FaceUp) || face == cube.FaceDown

	place(w, pos, t, user, ctx)
	return placed(ctx)
}

// Wax waxes the copper trapdoor to stop it from oxidising further.
func (t CopperTrapdoor) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if t.Waxed {
		return t, false
	}
	t.Waxed = true
	return t, true
}

func (t CopperTrapdoor) CanOxidate() bool {
	return !t.Waxed
}

func (t CopperTrapdoor) OxidationLevel() OxidationType {
	return t.Oxidation
}

func (t CopperTrapdoor) WithOxidationLevel(o OxidationType) Oxidizable {
	t.Oxidation = o
	return t
}

func (t CopperTrapdoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User, _ *item.UseContext) bool {
	t.Open = !t.Open
	w.SetBlock(pos, t, nil)
	if t.Open {
		w.PlaySound(pos.Vec3Centre(), sound.TrapdoorOpen{Block: t})
		return true
	}
	w.PlaySound(pos.Vec3Centre(), sound.TrapdoorClose{Block: t})
	return true
}

func (t CopperTrapdoor) SneakingActivate(pos cube.Pos, _ cube.Face, w *world.World, user item.User, _ *item.UseContext) bool {
	var ok bool
	t.Oxidation, t.Waxed, ok = activateOxidizable(pos, w, user, t.Oxidation, t.Waxed)
	if ok {
		w.SetBlock(pos, t, nil)
		return true
	}
	return false
}

func (t CopperTrapdoor) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	attemptOxidation(pos, w, r, t)
}

// BreakInfo ...
func (t CopperTrapdoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(t))
}

// SideClosed ...
func (t CopperTrapdoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (t CopperTrapdoor) EncodeItem() (name string, meta int16) {
	name = "copper_trapdoor"
	if t.Oxidation != NormalOxidation() {
		name = t.Oxidation.String() + "_" + name
	}
	if t.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, 0
}

// EncodeBlock ...
func (t CopperTrapdoor) EncodeBlock() (name string, properties map[string]any) {
	name = "copper_trapdoor"
	if t.Oxidation != NormalOxidation() {
		name = t.Oxidation.String() + "_" + name
	}
	if t.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, map[string]any{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
}

// allCopperTrapdoors returns a list of all copper trapdoor types
func allCopperTrapdoors() (trapdoors []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			for i := cube.Direction(0); i <= 3; i++ {
				trapdoors = append(trapdoors, CopperTrapdoor{Oxidation: o, Waxed: waxed, Facing: i, Open: false, Top: false})
				trapdoors = append(trapdoors, CopperTrapdoor{Oxidation: o, Waxed: waxed, Facing: i, Open: false, Top: true})
				trapdoors = append(trapdoors, CopperTrapdoor{Oxidation: o, Waxed: waxed, Facing: i, Open: true, Top: true})
				trapdoors = append(trapdoors, CopperTrapdoor{Oxidation: o, Waxed: waxed, Facing: i, Open: true, Top: false})
			}
		}
	}
	f(false)
	f(true)
	return
}
