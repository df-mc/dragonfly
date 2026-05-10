package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndPortalFrame is the indestructible block that forms the twelve-block ring of an End portal. Inserting an Eye of
// Ender flips the Eye bit and triggers a portal completion check.
type EndPortalFrame struct {
	solid
	bassDrum

	// Eye reports whether an Eye of Ender has been inserted into this frame.
	Eye bool
	// Facing is the cardinal direction this frame faces. Each frame in a valid ring faces inward, toward the centre
	// of the 3x3 interior.
	Facing cube.Direction
}

// EncodeItem ...
func (EndPortalFrame) EncodeItem() (name string, meta int16) {
	return "minecraft:end_portal_frame", 0
}

// EncodeBlock ...
func (f EndPortalFrame) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_portal_frame", map[string]any{
		"end_portal_eye_bit":           f.Eye,
		"minecraft:cardinal_direction": f.Facing.String(),
	}
}

// UseOnBlock places the frame opposite the placer's facing direction, matching the vanilla Bedrock convention
// (cardinal_direction = opposite of the placing player's facing). The new frame is always empty (Eye = false).
func (f EndPortalFrame) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, f)
	if !used {
		return false
	}
	f.Facing = user.Rotation().Direction().Opposite()
	f.Eye = false
	place(tx, pos, f, user, ctx)
	return placed(ctx)
}

// EndPortalFrameState exposes the frame's eye and facing state to the world/portal package without forcing it to
// import server/block, which would create an import cycle.
func (f EndPortalFrame) EndPortalFrameState() (eye bool, facing cube.Direction) {
	return f.Eye, f.Facing
}

// InsertEndPortalEye returns a copy of this frame with the eye inserted and ok=true. If the frame already holds an
// eye, it returns the original frame and ok=false.
func (f EndPortalFrame) InsertEndPortalEye() (world.Block, bool) {
	if f.Eye {
		return f, false
	}
	f.Eye = true
	return f, true
}

// allEndPortalFrames returns every state combination of EndPortalFrame for registration.
func allEndPortalFrames() []world.Block {
	frames := make([]world.Block, 0, len(cube.Directions())*2)
	for _, dir := range cube.Directions() {
		for _, eye := range []bool{false, true} {
			frames = append(frames, EndPortalFrame{Facing: dir, Eye: eye})
		}
	}
	return frames
}
