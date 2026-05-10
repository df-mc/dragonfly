package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
)

// EnderEye is the item used to fill End portal frames. In vanilla, throwing it locates the nearest stronghold; that
// projectile behaviour is not implemented in dragonfly.
type EnderEye struct{}

// EncodeItem ...
func (EnderEye) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_eye", 0
}

// MaxCount ...
func (EnderEye) MaxCount() int {
	return 64
}

// endPortalFrame is the local view of a block that can hold an Eye of Ender. block.EndPortalFrame implements it.
// InsertEndPortalEye returns the updated block and ok=true on success, or ok=false if the frame already had an eye.
type endPortalFrame interface {
	InsertEndPortalEye() (world.Block, bool)
}

// UseOnBlock fills the targeted End portal frame with an Eye of Ender. If the placement completes a twelve-frame ring,
// the End portal blocks are spawned in the interior. Has no effect on non-frame blocks or already-filled frames.
func (EnderEye) UseOnBlock(pos cube.Pos, _ cube.Face, _ mgl64.Vec3, tx *world.Tx, _ User, ctx *UseContext) bool {
	f, ok := tx.Block(pos).(endPortalFrame)
	if !ok {
		return false
	}
	updated, inserted := f.InsertEndPortalEye()
	if !inserted {
		return false
	}
	tx.SetBlock(pos, updated, nil)
	portal.ActivateEndPortal(tx, pos)
	ctx.SubtractFromCount(1)
	return true
}
