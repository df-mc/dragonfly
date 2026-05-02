package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Piston is a block capable of pushing blocks when powered by redstone.
type Piston struct {
	solid
	transparent

	// Facing is the direction the piston faces.
	Facing cube.Face
	// Extended is true if the piston is currently extended.
	Extended bool
}

// BreakInfo ...
func (p Piston) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, oneOf(p))
}

// UseOnBlock ...
func (p Piston) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, p)
	if !used {
		return
	}
	p.Facing = user.Rotation().Direction().Face().Opposite()

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (p Piston) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
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
func (p Piston) extend(pos cube.Pos, tx *world.Tx) bool {
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
	tx.SetBlock(pushPos, PistonArmCollision{Facing: p.Facing, Sticky: false}, nil)
	return true
}

// retract handles the retraction of the piston.
func (p Piston) retract(pos cube.Pos, tx *world.Tx) {
	armPos := pos.Side(p.Facing)
	tx.SetBlock(armPos, Air{}, nil)
}

// isImmovable returns true if the block cannot be pushed.
func (p Piston) isImmovable(b world.Block) bool {
	switch b.(type) {
	case Bedrock, Obsidian, InvisibleBedrock, Barrier, Water, Lava:
		return true
	}
	return false
}

// isPowered checks if the piston is powered by an adjacent redstone block.
func (p Piston) isPowered(pos cube.Pos, tx *world.Tx) bool {
	for _, face := range cube.Faces() {
		if _, ok := tx.Block(pos.Side(face)).(RedstoneBlock); ok {
			return true
		}
	}
	return false
}

const hashPiston = 12347 // Temporary constant

// Hash ...
func (p Piston) Hash() (uint64, uint64) {
	return hashPiston, uint64(p.Facing) | uint64(boolByte(p.Extended))<<3
}

// EncodeItem ...
func (p Piston) EncodeItem() (name string, meta int16) {
	return "minecraft:piston", 0
}

// EncodeBlock ...
func (p Piston) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston", map[string]any{"facing_direction": int32(p.Facing), "extended_bit": p.Extended}
}

// allPistons ...
func allPistons() (pistons []world.Block) {
	for _, face := range cube.Faces() {
		for _, extended := range []bool{false, true} {
			pistons = append(pistons, Piston{Facing: face, Extended: extended})
		}
	}
	return
}
