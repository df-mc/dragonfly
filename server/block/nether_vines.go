package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NetherVines are climbable non-solid Nether plants.
type NetherVines struct {
	transparent
	replaceable
	empty

	// Twisting specifies if the vine is the upward-growing twisting variant.
	Twisting bool
	// Age is the current vine age.
	Age int
}

// EntityInside ...
func (NetherVines) EntityInside(_ cube.Pos, _ *world.Tx, e world.Entity) {
	if fallEntity, ok := e.(fallDistanceEntity); ok {
		fallEntity.ResetFallDistance()
	}
}

// NeighbourUpdateTick ...
func (v NetherVines) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if v.Twisting {
		below := tx.Block(pos.Side(cube.FaceDown))
		if _, ok := below.(NetherVines); ok {
			return
		}
		if !supportsTwistingVines(below) {
			breakBlock(v, pos, tx)
		}
		return
	}

	above := tx.Block(pos.Side(cube.FaceUp))
	if _, ok := above.(NetherVines); ok {
		return
	}
	if !supportsWeepingVines(above) {
		breakBlock(v, pos, tx)
	}
}

// UseOnBlock ...
func (v NetherVines) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, v)
	if !used {
		return false
	}
	if v.Twisting {
		if !supportsTwistingVines(tx.Block(pos.Side(cube.FaceDown))) {
			return false
		}
	} else if !supportsWeepingVines(tx.Block(pos.Side(cube.FaceUp))) {
		return false
	}
	place(tx, pos, v, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (NetherVines) HasLiquidDrops() bool {
	return false
}

// FlammabilityInfo ...
func (NetherVines) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(15, 100, true)
}

// BreakInfo ...
func (v NetherVines) BreakInfo() BreakInfo {
	return newBreakInfo(0, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears
	}, nothingEffective, oneOf(v))
}

// CompostChance ...
func (NetherVines) CompostChance() float64 {
	return 0.5
}

// EncodeItem ...
func (v NetherVines) EncodeItem() (name string, meta int16) {
	if v.Twisting {
		return "minecraft:twisting_vines", 0
	}
	return "minecraft:weeping_vines", 0
}

// EncodeBlock ...
func (v NetherVines) EncodeBlock() (string, map[string]any) {
	if v.Twisting {
		return "minecraft:twisting_vines", map[string]any{"twisting_vines_age": int32(v.Age)}
	}
	return "minecraft:weeping_vines", map[string]any{"weeping_vines_age": int32(v.Age)}
}

func supportsNetherFlora(b world.Block) bool {
	switch b.(type) {
	case Nylium:
		return true
	default:
		return false
	}
}

func supportsNetherRoots(b world.Block) bool {
	switch b.(type) {
	case Nylium, SoulSoil:
		return true
	default:
		return false
	}
}

func supportsTwistingVines(b world.Block) bool {
	switch b.(type) {
	case Netherrack, Nylium, NetherWartBlock, Blackstone:
		return true
	default:
		return false
	}
}

func supportsWeepingVines(b world.Block) bool {
	switch b.(type) {
	case Netherrack, NetherWartBlock, Wood, Log:
		return true
	default:
		return false
	}
}
