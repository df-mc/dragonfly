package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// StickyPiston is a block capable of pushing and pulling blocks.
type StickyPiston struct {
	solid
	transparent

	// Facing is the direction the piston faces.
	Facing cube.Face
	// Extended is true if the piston is currently extended.
	Extended bool
}

// BreakInfo ...
func (p StickyPiston) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, oneOf(p))
}

// UseOnBlock ...
func (p StickyPiston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, p)
	if !used {
		return
	}
	p.Facing = cube.OrientationFromRotation(user.Rotation()).Facing().Opposite()

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (p StickyPiston) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if p.Extended {
		if !p.isPowered(pos, tx) {
			p.Extended = false
			tx.SetBlock(pos, p, nil)
			p.retract(pos, tx)
		}
		return
	}
	if p.isPowered(pos, tx) {
		if p.extend(pos, tx) {
			p.Extended = true
			tx.SetBlock(pos, p, nil)
		}
	}
}

// extend handles the extension of the piston, pushing up to 12 blocks.
func (p StickyPiston) extend(pos cube.Pos, tx *world.Tx) bool {
	dir := p.Facing
	pushPos := pos.Side(dir)

	blocks := []cube.Pos{}
	curr := pushPos
	for len(blocks) < 12 {
		b := tx.Block(curr)
		if _, ok := b.(Air); ok {
			break
		}
		if p.isImmovable(b) {
			return false
		}
		blocks = append(blocks, curr)
		curr = curr.Side(dir)
	}

	if len(blocks) >= 12 {
		b := tx.Block(curr)
		if len(b.Model().BBox(curr, tx)) > 0 {
			return false
		}
	}

	// Move blocks in reverse order to avoid overwriting.
	for i := len(blocks) - 1; i >= 0; i-- {
		b := tx.Block(blocks[i])
		tx.SetBlock(blocks[i].Side(dir), b, nil)
	}
	tx.SetBlock(pushPos, PistonArmCollision{Facing: p.Facing, Sticky: true}, nil)
	return true
}

// retract handles the retraction of the piston.
func (p StickyPiston) retract(pos cube.Pos, tx *world.Tx) {
	armPos := pos.Side(p.Facing)
	tx.SetBlock(armPos, Air{}, nil)

	pullPos := armPos.Side(p.Facing)
	b := tx.Block(pullPos)
	if !p.isImmovable(b) && len(b.Model().BBox(pullPos, tx)) > 0 {
		tx.SetBlock(armPos, b, nil)
		tx.SetBlock(pullPos, Air{}, nil)
	}
}

// isImmovable returns true if the block cannot be pushed.
func (p StickyPiston) isImmovable(b world.Block) bool {
	switch b.(type) {
	case Bedrock, Obsidian, InvisibleBedrock, Barrier, Water, Lava:
		return true
	}
	return false
}

// isPowered checks if the piston is powered by an adjacent redstone block.
func (p StickyPiston) isPowered(pos cube.Pos, tx *world.Tx) bool {
	for _, face := range cube.Faces() {
		if _, ok := tx.Block(pos.Side(face)).(RedstoneBlock); ok {
			return true
		}
	}
	return false
}

const hashStickyPiston = 12348 // Temporary constant

// Hash ...
func (p StickyPiston) Hash() (uint64, uint64) {
	return hashStickyPiston, uint64(p.Facing) | uint64(boolByte(p.Extended))<<3
}

// Model ...
func (p StickyPiston) Model() world.BlockModel {
	return world.FullBlockModel{}
}

// EncodeItem ...
func (p StickyPiston) EncodeItem() (name string, meta int16) {
	return "minecraft:sticky_piston", 0
}

// EncodeBlock ...
func (p StickyPiston) EncodeBlock() (string, map[string]any) {
	return "minecraft:sticky_piston", map[string]any{"facing_direction": int32(p.Facing), "extended_bit": p.Extended}
}

// allStickyPistons ...
func allStickyPistons() (pistons []world.Block) {
	for _, face := range cube.Faces() {
		for _, extended := range []bool{false, true} {
			pistons = append(pistons, StickyPiston{Facing: face, Extended: extended})
		}
	}
	return
}
