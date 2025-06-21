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

// IronTrapDoor is a solid, transparent block that can be used as an openable 1×1 barrier
// can only be opened by using redstone.
type IronTrapDoor struct {
	transparent
	bass
	sourceWaterDisplacer

	// Facing is the direction the trapdoor is facing.
	Facing cube.Direction
	// Open is whether the trapdoor is open.
	Open bool
	// Top is whether the trapdoor occupies the top or bottom part of a block.
	Top bool
}

// Model ...
func (t IronTrapDoor) Model() world.BlockModel {
	return model.Trapdoor{Facing: t.Facing, Top: t.Top, Open: t.Open}
}

// UseOnBlock handles the directional placing of trapdoors and makes sure they are properly placed upside down
// when needed.
func (t IronTrapDoor) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, t)
	if !used {
		return false
	}
	t.Facing = user.Rotation().Direction().Opposite()
	t.Top = (clickPos.Y() > 0.5 && face != cube.FaceUp) || face == cube.FaceDown

	place(tx, pos, t, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (t IronTrapDoor) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(t))
}

// SideClosed ...
func (t IronTrapDoor) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// RedstoneUpdate ...
func (t IronTrapDoor) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	if t.Open == receivedRedstonePower(pos, tx) {
		return
	}

	t.Open = receivedRedstonePower(pos, tx)
	tx.SetBlock(pos, t, nil)

	if t.Open {
		tx.PlaySound(pos.Vec3Centre(), sound.TrapdoorOpen{Block: t})
	} else {
		tx.PlaySound(pos.Vec3Centre(), sound.TrapdoorClose{Block: t})
	}
}

// EncodeItem ...
func (t IronTrapDoor) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_trapdoor", 0
}

// EncodeBlock ...
func (t IronTrapDoor) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:iron_trapdoor", map[string]any{"direction": int32(math.Abs(float64(t.Facing) - 3)), "open_bit": t.Open, "upside_down_bit": t.Top}
}

// allIronTrapdoors returns a list of all trapdoor types
func allIronTrapdoors() (trapdoors []world.Block) {
	for i := cube.Direction(0); i <= 3; i++ {
		trapdoors = append(trapdoors, IronTrapDoor{Facing: i, Open: false, Top: false})
		trapdoors = append(trapdoors, IronTrapDoor{Facing: i, Open: false, Top: true})
		trapdoors = append(trapdoors, IronTrapDoor{Facing: i, Open: true, Top: true})
		trapdoors = append(trapdoors, IronTrapDoor{Facing: i, Open: true, Top: false})
	}
	return
}
