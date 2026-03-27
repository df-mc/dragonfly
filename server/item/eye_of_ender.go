package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	worldportal "github.com/df-mc/dragonfly/server/world/portal"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// EyeOfEnder is an item used to activate end portal frames.
type EyeOfEnder struct{}

// MaxCount ...
func (EyeOfEnder) MaxCount() int {
	return 16
}

// Use throws an eye of ender signal.
func (EyeOfEnder) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().EyeOfEnder
	if create == nil {
		return false
	}

	opts := world.EntitySpawnOpts{
		Position: eyePosition(user),
		Velocity: user.Rotation().Vec3().Mul(1.25),
	}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// UseOnBlock inserts the eye into an end portal frame.
func (EyeOfEnder) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, user User, ctx *UseContext) bool {
	name, properties := tx.Block(pos).EncodeBlock()
	if name != "minecraft:end_portal_frame" && name != "end_portal_frame" {
		return false
	}
	if properties == nil {
		return false
	}
	if eye, ok := properties["end_portal_eye_bit"].(bool); !ok || eye {
		return false
	}

	props := make(map[string]any, len(properties))
	for key, value := range properties {
		props[key] = value
	}
	props["end_portal_eye_bit"] = true
	frameWithEye, ok := world.BlockByName("minecraft:end_portal_frame", props)
	if !ok {
		return false
	}
	tx.SetBlock(pos, frameWithEye, nil)
	worldportal.TryActivateEndPortal(tx, pos)

	if gm, ok := user.(interface{ GameMode() world.GameMode }); !ok || !gm.GameMode().CreativeInventory() {
		ctx.SubtractFromCount(1)
	}
	return true
}

// EncodeItem ...
func (EyeOfEnder) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_eye", 0
}
