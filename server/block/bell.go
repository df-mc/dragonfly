package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Bell is a transparent, animated block entity that produces a sound when used. Unlike most utility blocks, bells
// cannot be crafted.
type Bell struct {
	transparent

	// Attach represents the attachment type of the Bell.
	Attach BellAttachment
	// Facing represents the direction the Bell is facing.
	Facing cube.Direction
}

// Model ...
func (b Bell) Model() world.BlockModel {
	// TODO: Use the actual bell model.
	return model.Solid{}
}

// BreakInfo ...
func (b Bell) BreakInfo() BreakInfo {
	return newBreakInfo(1, pickaxeHarvestable, pickaxeEffective, oneOf(b)).withBlastResistance(15)
}

// UseOnBlock ...
func (b Bell) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}
	b.Facing = user.Facing().Opposite()
	if face == cube.FaceUp {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).Model().(model.Solid); !ok {
			return false
		}
	} else if face == cube.FaceDown {
		if _, ok := w.Block(pos.Side(cube.FaceUp)).Model().(model.Solid); !ok {
			return false
		}
		b.Attach = HangingBellAttachment()
	} else {
		if _, ok := w.Block(pos.Side(face.Opposite())).Model().(model.Solid); !ok {
			return false
		}
		b.Facing = face.Direction()
		b.Attach = WallBellAttachment()
		if _, ok := w.Block(pos.Side(face)).Model().(model.Solid); ok {
			b.Attach = WallsBellAttachment()
		}
	}
	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (b Bell) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	var supportFaces []cube.Face
	switch b.Attach {
	case HangingBellAttachment():
		supportFaces = append(supportFaces, cube.FaceUp)
	case StandingBellAttachment():
		supportFaces = append(supportFaces, cube.FaceDown)
	case WallBellAttachment(), WallsBellAttachment():
		supportFaces = append(supportFaces, b.Facing.Face().Opposite())
		if b.Attach == WallsBellAttachment() {
			supportFaces = append(supportFaces, b.Facing.Face())
		}
	}
	for _, supportFace := range supportFaces {
		if _, ok := w.Block(pos.Side(supportFace)).Model().(model.Solid); !ok {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})
			dropItem(w, item.NewStack(b, 1), pos.Vec3Centre())
			break
		}
	}
}

// Activate ...
func (b Bell) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	s, f := u.Facing().Opposite().Face(), b.Facing.Face()
	switch b.Attach {
	case HangingBellAttachment():
		b.Ring(pos, s, w)
		return true
	case StandingBellAttachment():
		if s.Axis() == f.Axis() {
			b.Ring(pos, s, w)
			return true
		}
	case WallBellAttachment(), WallsBellAttachment():
		if s == f.RotateLeft() || s == f.RotateRight() {
			b.Ring(pos, s, w)
			return true
		}
	}
	return false
}

// ProjectileHit ...
func (b Bell) ProjectileHit(w *world.World, _ world.Entity, pos cube.Pos, face cube.Face) {
	b.Ring(pos, face, w)
}

// Ring rings the bell on the face passed.
func (b Bell) Ring(pos cube.Pos, face cube.Face, w *world.World) {
	w.PlaySound(pos.Vec3Centre(), sound.BellRing{})
	for _, v := range w.Viewers(pos.Vec3Centre()) {
		v.ViewBlockAction(pos, BellRing{Face: face})
	}
}

// EncodeNBT encodes the Bell's block entity ID. There are other properties, but we can skip those.
func (b Bell) EncodeNBT() map[string]any {
	return map[string]any{"id": "Bell"}
}

// DecodeNBT ...
func (b Bell) DecodeNBT(map[string]any) any {
	return b
}

// EncodeItem ...
func (b Bell) EncodeItem() (name string, meta int16) {
	return "minecraft:bell", 0
}

// EncodeBlock ...
func (b Bell) EncodeBlock() (string, map[string]any) {
	return "minecraft:bell", map[string]any{
		"toggle_bit": uint8(0), // Useless property, updated on ring in vanilla.
		"attachment": b.Attach.String(),
		"direction":  int32(horizontalDirection(b.Facing)),
	}
}

// allBells ...
func allBells() (bells []world.Block) {
	for _, a := range BellAttachments() {
		for _, d := range cube.Directions() {
			bells = append(bells, Bell{Attach: a, Facing: d})
		}
	}
	return
}
