package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// EnderEye is the item used to fill End portal frames. The stronghold-locating throw is not implemented.
type EnderEye struct{}

// EncodeItem ...
func (EnderEye) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_eye", 0
}

// MaxCount ...
func (EnderEye) MaxCount() int {
	return 64
}

// endPortalFrame is implemented by block.EndPortalFrame, which cannot be imported here directly.
type endPortalFrame interface {
	InsertEndPortalEye() (world.Block, bool)
}

// UseOnBlock fills the targeted End portal frame with an Eye of Ender, activating the portal if this completes the
// twelve-frame ring.
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
	tx.PlaySound(pos.Vec3Centre(), sound.EnderEyePlaced{})
	if portal.ActivateEndPortal(tx, pos) {
		tx.PlaySound(pos.Vec3Centre(), sound.EndPortalCreated{})
	}
	ctx.SubtractFromCount(1)
	return true
}
