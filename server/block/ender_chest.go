package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"go.uber.org/atomic"
)

// enderChestOwner represents an entity that has an ender chest inventory.
type enderChestOwner interface {
	ContainerOpener
	EnderChestInventory() *inventory.Inventory
}

// EnderChest is a type of chest whose contents are exclusive to each player, and can be accessed from anywhere.
// The empty value of EnderChest is not valid. It must be created using block.NewEnderChest().
type EnderChest struct {
	chest
	transparent
	bass

	// Facing is the direction that the ender chest is facing.
	Facing cube.Direction

	viewers *atomic.Int64
}

// NewEnderChest creates a new initialised ender chest.
func NewEnderChest() EnderChest {
	return EnderChest{viewers: atomic.NewInt64(0)}
}

// BreakInfo ...
func (c EnderChest) BreakInfo() BreakInfo {
	return newBreakInfo(22.5, pickaxeHarvestable, pickaxeEffective, silkTouchDrop(item.NewStack(Obsidian{}, 8), item.NewStack(NewEnderChest(), 1))).withBlastResistance(3000)
}

// LightEmissionLevel ...
func (c EnderChest) LightEmissionLevel() uint8 {
	return 7
}

// CanDisplace ...
func (EnderChest) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (EnderChest) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// UseOnBlock ...
func (c EnderChest) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	c = NewEnderChest()
	c.Facing = user.Facing().Opposite()

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// Activate ...
func (c EnderChest) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(enderChestOwner); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// AddViewer ...
func (c EnderChest) AddViewer(w *world.World, pos cube.Pos) {
	if c.viewers.Inc() == 1 {
		c.open(w, pos)
	}
}

// RemoveViewer ...
func (c EnderChest) RemoveViewer(w *world.World, pos cube.Pos) {
	if c.viewers.Load() == 0 {
		return
	}
	if c.viewers.Dec() == 0 {
		c.close(w, pos)
	}
}

// open opens the ender chest, displaying the animation and playing a sound.
func (c EnderChest) open(w *world.World, pos cube.Pos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, OpenAction{})
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestOpen{})
}

// close closes the ender chest, displaying the animation and playing a sound.
func (c EnderChest) close(w *world.World, pos cube.Pos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, CloseAction{})
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestClose{})
}

// EncodeNBT ...
func (c EnderChest) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{"id": "EnderChest"}
}

// DecodeNBT ...
func (c EnderChest) DecodeNBT(map[string]interface{}) interface{} {
	return NewEnderChest()
}

// EncodeItem ...
func (EnderChest) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_chest", 0
}

// EncodeBlock ...
func (c EnderChest) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:ender_chest", map[string]interface{}{"facing_direction": 2 + int32(c.Facing)}
}

// allEnderChests ...
func allEnderChests() (chests []world.Block) {
	for _, direction := range cube.Directions() {
		chests = append(chests, EnderChest{Facing: direction})
	}
	return
}
