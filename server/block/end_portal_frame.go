package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndPortalFrame is the indestructible block that forms the twelve-block ring of an End portal.
type EndPortalFrame struct {
	bassDrum

	// Eye is true if an Eye of Ender has been inserted into the frame.
	Eye bool
	// Facing is the direction the frame faces. Each frame in a valid ring faces the centre of the 3x3 interior.
	Facing cube.Direction
}

// Model ...
func (f EndPortalFrame) Model() world.BlockModel {
	return model.EndPortalFrame{Eye: f.Eye}
}

// LightEmissionLevel returns 1.
func (EndPortalFrame) LightEmissionLevel() uint8 {
	return 1
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

// UseOnBlock ...
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

// EndPortalFrameState returns the frame's eye and facing state.
func (f EndPortalFrame) EndPortalFrameState() (eye bool, facing cube.Direction) {
	return f.Eye, f.Facing
}

// InsertEndPortalEye returns a copy of the frame with an eye inserted, or false if it already held one.
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
