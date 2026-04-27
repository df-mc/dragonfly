package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// String is an item obtained from spiders and cobwebs. When placed, it creates a tripwire that
// detects entities passing through it.
type String struct {
	empty
	transparent

	// Attached is true if the tripwire is connected to valid tripwire hooks on both sides.
	Attached bool
	// Disarmed is true if the tripwire was cut using shears, preventing it from activating.
	Disarmed bool
	// Powered is true if the tripwire is currently activated by an entity passing through it.
	Powered bool
	// Suspended is true if the tripwire is not resting on a solid surface.
	Suspended bool
}

// UseOnBlock places the string as a tripwire on the target surface.
func (s String) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
    pos, _, canPlace := firstReplaceable(tx, pos, face, s)
    if !canPlace {
        return false
    }
    below := pos.Side(cube.FaceDown)
    s.Suspended = !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx)
    place(tx, pos, s, user, ctx)
    return placed(ctx)
}

func (s String) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	below := pos.Side(cube.FaceDown)
	suspended := !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx)
	if suspended != s.Suspended {
		s.Suspended = suspended
		tx.SetBlock(pos, s, nil)
	}
}

// HasLiquidDrops ...
func (s String) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (s String) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(String{}))
}

// EncodeItem ...
func (s String) EncodeItem() (name string, meta int16) {
	return "minecraft:string", 0
}

// EncodeBlock ...
func (s String) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:trip_wire", map[string]any{
		"attached_bit":  boolByte(s.Attached),
		"disarmed_bit":  boolByte(s.Disarmed),
		"powered_bit":   boolByte(s.Powered),
		"suspended_bit": boolByte(s.Suspended),
	}
}

// allString ...
func allString() (blocks []world.Block) {
	for meta := 0; meta < 16; meta++ {
		blocks = append(blocks, String{
			Powered:   meta&0x1 != 0,
			Suspended: meta&0x2 != 0,
			Attached:  meta&0x4 != 0,
			Disarmed:  meta&0x8 != 0,
		})
	}
	return blocks
}
