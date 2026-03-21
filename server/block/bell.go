package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// Bell is a transparent block that can be placed on the floor, ceiling, or walls and rang by interacting with it.
// Village and raid behaviour is intentionally not implemented.
type Bell struct {
	transparent
	sourceWaterDisplacer

	// Attach represents the attachment type of the Bell.
	Attach BellAttachment
	// Facing represents the horizontal direction of the Bell.
	Facing cube.Direction
	// Toggle is true while the Bell is ringing.
	Toggle bool

	// ringTicks is the number of ticks the Bell has been ringing for.
	ringTicks int
	// ringDirection is the direction the Bell is ringing towards.
	ringDirection cube.Face
}

// Model ...
func (b Bell) Model() world.BlockModel {
	return model.Bell{Attachment: model.BellAttachment(b.Attach.Uint8()), Facing: b.Facing}
}

// BreakInfo ...
func (b Bell) BreakInfo() BreakInfo {
	return newBreakInfo(5, alwaysHarvestable, pickaxeEffective, oneOf(b)).withBlastResistance(5)
}

// Activate ...
func (b Bell) Activate(pos cube.Pos, face cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	return b.ring(pos, tx, face)
}

// ProjectileHit ...
func (b Bell) ProjectileHit(pos cube.Pos, tx *world.Tx, _ world.Entity, face cube.Face) {
	b.ring(pos, tx, face)
}

// UseOnBlock ...
func (b Bell) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	switch face {
	case cube.FaceUp:
		if !bellSupportSolid(tx, pos, cube.FaceDown) {
			return false
		}
		b.Attach = StandingBellAttachment()
		b.Facing = user.Rotation().Direction().Opposite()
	case cube.FaceDown:
		if !bellSupportSolid(tx, pos, cube.FaceUp) {
			return false
		}
		b.Attach = HangingBellAttachment()
		b.Facing = user.Rotation().Direction().Opposite()
	default:
		if !bellSupportSolid(tx, pos, face.Opposite()) {
			return false
		}
		b.Attach = SideBellAttachment()
		b.Facing = face.Direction()
		if bellSupportSolid(tx, pos, face) {
			b.Attach = MultipleBellAttachment()
		}
	}
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (b Bell) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	switch b.Attach {
	case StandingBellAttachment():
		if !bellSupportSolid(tx, pos, cube.FaceDown) {
			breakBlock(b, pos, tx)
		}
	case HangingBellAttachment():
		if !bellSupportSolid(tx, pos, cube.FaceUp) {
			breakBlock(b, pos, tx)
		}
	case SideBellAttachment():
		if !bellSupportSolid(tx, pos, b.Facing.Face().Opposite()) {
			breakBlock(b, pos, tx)
		}
	case MultipleBellAttachment():
		positive, negative := b.multipleSupportFaces()
		positiveSolid := bellSupportSolid(tx, pos, positive)
		negativeSolid := bellSupportSolid(tx, pos, negative)

		switch {
		case positiveSolid && negativeSolid:
		case positiveSolid:
			b.Attach = SideBellAttachment()
			b.Facing = positive.Opposite().Direction()
			tx.SetBlock(pos, b, &world.SetOpts{DisableBlockUpdates: true, DisableLiquidDisplacement: true})
		case negativeSolid:
			b.Attach = SideBellAttachment()
			tx.SetBlock(pos, b, &world.SetOpts{DisableBlockUpdates: true, DisableLiquidDisplacement: true})
		default:
			breakBlock(b, pos, tx)
		}
	}
}

// Tick ...
func (b Bell) Tick(_ int64, pos cube.Pos, tx *world.Tx) {
	if !b.Toggle {
		return
	}
	b.ringTicks++
	if b.ringTicks < 50 {
		tx.SetBlock(pos, b, &world.SetOpts{DisableBlockUpdates: true, DisableLiquidDisplacement: true})
		return
	}
	b.Toggle, b.ringTicks = false, 0
	tx.SetBlock(pos, b, &world.SetOpts{DisableBlockUpdates: true, DisableLiquidDisplacement: true})
}

// EncodeBlock ...
func (b Bell) EncodeBlock() (string, map[string]any) {
	return "minecraft:bell", map[string]any{
		"attachment": b.Attach.String(),
		"direction":  int32(horizontalDirection(b.Facing)),
		"toggle_bit": boolByte(b.Toggle),
	}
}

// EncodeItem ...
func (Bell) EncodeItem() (name string, meta int16) {
	return "minecraft:bell", 0
}

// EncodeNBT ...
func (b Bell) EncodeNBT() map[string]any {
	return map[string]any{
		"id":        "Bell",
		"Ringing":   boolByte(b.Toggle),
		"Ticks":     int32(b.ringTicks),
		"Direction": int32(bellFace(int32(b.ringDirection), b.Facing.Face())),
	}
}

// DecodeNBT ...
func (b Bell) DecodeNBT(data map[string]any) any {
	b.Toggle = nbtconv.Bool(data, "Ringing")
	b.ringTicks = int(nbtconv.Int32(data, "Ticks"))
	b.ringDirection = bellFace(nbtconv.Int32(data, "Direction"), b.Facing.Face())
	return b
}

// Explode ...
func (b Bell) Explode(_ mgl64.Vec3, pos cube.Pos, tx *world.Tx, c ExplosionConfig) {
	b.ring(pos, tx, b.Facing.Face())

	if breakHandler := b.BreakInfo().BreakHandler; breakHandler != nil {
		breakHandler(pos, tx, nil)
	}
	tx.SetBlock(pos, nil, nil)

	if c.ItemDropChance < 0 {
		return
	}
	if c.ItemDropChance < 1 && c.ItemDropChance <= rand.Float64() {
		return
	}
	for _, drop := range b.BreakInfo().Drops(item.ToolNone{}, nil) {
		dropItem(tx, drop, pos.Vec3Centre())
	}
}

// SideClosed ...
func (Bell) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// ring starts or refreshes the Bell ring state.
func (b Bell) ring(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	b.Toggle, b.ringTicks = true, 0
	b.ringDirection = bellFace(int32(face), b.Facing.Face())
	tx.SetBlock(pos, b, &world.SetOpts{DisableBlockUpdates: true, DisableLiquidDisplacement: true})
	tx.PlaySound(pos.Vec3Centre(), sound.BellRing{})
	return true
}

// multipleSupportFaces returns the faces checked for a Bell attached between two side supports.
func (b Bell) multipleSupportFaces() (cube.Face, cube.Face) {
	face := b.Facing.Face()
	return face, face.Opposite()
}

// bellSupportSolid reports whether the block on the given side of the Bell can support it.
func bellSupportSolid(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	supportPos := pos.Side(face)
	return tx.Block(supportPos).Model().FaceSolid(supportPos, face.Opposite(), tx)
}

// bellFace returns a cube.Face from the given int32, or the fallback if the int32 is not a valid horizontal face.
func bellFace(v int32, fallback cube.Face) cube.Face {
	f := cube.Face(v)
	if f < cube.FaceDown || f > cube.FaceEast || f.Axis() == cube.Y {
		return fallback
	}
	return f
}

// allBells ...
func allBells() (bells []world.Block) {
	for _, a := range BellAttachments() {
		for _, d := range cube.Directions() {
			bells = append(bells, Bell{Attach: a, Facing: d})
			bells = append(bells, Bell{Attach: a, Facing: d, Toggle: true})
		}
	}
	return
}
