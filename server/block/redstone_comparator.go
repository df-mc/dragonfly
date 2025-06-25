package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneComparator is a redstone component used to maintain, compare, or subtract signal strength, or to measure
// certain block states (primarily the fullness of containers).
type RedstoneComparator struct {
	transparent

	// Facing is the direction from the torch to the block.
	Facing cube.Direction
	// Subtract is true if the comparator is in subtract mode.
	Subtract bool
	// Powered is true if the repeater is powered by a redstone signal.
	Powered bool
	// Power is the current power level of the redstone comparator. It ranges from 0 to 15.
	Power int
}

// HasLiquidDrops ...
func (RedstoneComparator) HasLiquidDrops() bool {
	return true
}

// Model ...
func (RedstoneComparator) Model() world.BlockModel {
	return model.Diode{}
}

// EncodeItem ...
func (RedstoneComparator) EncodeItem() (name string, meta int16) {
	return "minecraft:comparator", 0
}

// EncodeBlock ...
func (r RedstoneComparator) EncodeBlock() (string, map[string]any) {
	name := "minecraft:unpowered_comparator"
	if r.Powered {
		name = "minecraft:powered_comparator"
	}
	return name, map[string]any{
		"minecraft:cardinal_direction": r.Facing.String(),
		"output_lit_bit":               boolByte(r.Powered),
		"output_subtract_bit":          boolByte(r.Subtract),
	}
}

// UseOnBlock ...
func (r RedstoneComparator) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used {
		return false
	}
	if d, ok := tx.Block(pos.Side(cube.FaceDown)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		return false
	}
	r.Facing = user.Rotation().Direction().Opposite()

	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (r RedstoneComparator) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if d, ok := tx.Block(pos.Side(cube.FaceDown)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		breakBlock(r, pos, tx)
	}
}

// Activate ...
func (r RedstoneComparator) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	r.Subtract = !r.Subtract
	tx.SetBlock(pos, r, nil)
	return false
}

// RedstoneSource ...
func (r RedstoneComparator) RedstoneSource() bool {
	return r.Powered
}

// WeakPower ...
func (r RedstoneComparator) WeakPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	return 0
}

// StrongPower ...
func (r RedstoneComparator) StrongPower(pos cube.Pos, face cube.Face, tx *world.Tx, accountForDust bool) int {
	return r.WeakPower(pos, face, tx, accountForDust)
}

// EncodeNBT ...
func (r RedstoneComparator) EncodeNBT() map[string]any {
	return map[string]any{"OutputSignal": int32(r.Power)}
}

// DecodeNBT ...
func (r RedstoneComparator) DecodeNBT(data map[string]any) any {
	r.Power = int(nbtconv.Int32(data, "OutputSignal"))
	return r
}

// allRedstoneComparators ...
func allRedstoneComparators() (comparators []world.Block) {
	for _, d := range cube.Directions() {
		for _, s := range []bool{false, true} {
			for _, p := range []bool{false, true} {
				comparators = append(comparators, RedstoneComparator{Facing: d, Subtract: s, Powered: p})
			}
		}
	}
	return
}
