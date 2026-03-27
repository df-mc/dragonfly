package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndPortalFrame is the frame block used to create End portals.
type EndPortalFrame struct {
	transparent

	// Facing is the direction the frame points toward.
	Facing cube.Direction
	// Eye specifies if an eye of ender has been inserted.
	Eye bool
}

// Model returns the end portal frame model.
func (f EndPortalFrame) Model() world.BlockModel {
	return model.EndPortalFrame{Eye: f.Eye}
}

// LightEmissionLevel ...
func (EndPortalFrame) LightEmissionLevel() uint8 {
	return 1
}

// UseOnBlock places an end portal frame oriented toward the player.
func (f EndPortalFrame) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, f)
	if !used {
		return false
	}
	if user != nil {
		f.Facing = user.Rotation().Direction().Opposite()
	} else {
		f.Facing = cube.South
	}
	place(tx, pos, f, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (EndPortalFrame) EncodeItem() (name string, meta int16) {
	return "minecraft:end_portal_frame", 0
}

// EncodeBlock ...
func (f EndPortalFrame) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_portal_frame", map[string]any{
		"minecraft:cardinal_direction": f.Facing.String(),
		"end_portal_eye_bit":           f.Eye,
	}
}

func allEndPortalFrames() (b []world.Block) {
	for _, facing := range []cube.Direction{cube.North, cube.East, cube.South, cube.West} {
		b = append(b, EndPortalFrame{Facing: facing})
		b = append(b, EndPortalFrame{Facing: facing, Eye: true})
	}
	return
}
