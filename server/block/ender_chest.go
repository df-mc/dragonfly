package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"sync/atomic"
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
	sourceWaterDisplacer

	// Facing is the direction that the ender chest is facing.
	Facing cube.Direction

	viewers *atomic.Int64
}

// NewEnderChest creates a new initialised ender chest.
func NewEnderChest() EnderChest {
	return EnderChest{viewers: &atomic.Int64{}}
}

// BreakInfo ...
func (c EnderChest) BreakInfo() BreakInfo {
	return newBreakInfo(22.5, pickaxeHarvestable, pickaxeEffective, silkTouchDrop(item.NewStack(Obsidian{}, 8), item.NewStack(NewEnderChest(), 1))).withBlastResistance(3000)
}

// LightEmissionLevel ...
func (c EnderChest) LightEmissionLevel() uint8 {
	return 7
}

// SideClosed ...
func (EnderChest) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// UseOnBlock ...
func (c EnderChest) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	c = NewEnderChest()
	c.Facing = user.Rotation().Direction().Opposite()

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// Activate ...
func (c EnderChest) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(enderChestOwner); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

// AddViewer ...
func (c EnderChest) AddViewer(tx *world.Tx, pos cube.Pos) {
	if c.viewers.Add(1) == 1 {
		c.open(tx, pos)
	}
}

// RemoveViewer ...
func (c EnderChest) RemoveViewer(tx *world.Tx, pos cube.Pos) {
	if c.viewers.Load() == 0 {
		return
	}
	if c.viewers.Add(-1) == 0 {
		c.close(tx, pos)
	}
}

// open opens the ender chest, displaying the animation and playing a sound.
func (c EnderChest) open(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, OpenAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.EnderChestOpen{})
}

// close closes the ender chest, displaying the animation and playing a sound.
func (c EnderChest) close(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, CloseAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.EnderChestClose{})
}

// EncodeNBT ...
func (c EnderChest) EncodeNBT() map[string]interface{} {
	return map[string]interface{}{"id": "EnderChest"}
}

// DecodeNBT ...
func (c EnderChest) DecodeNBT(map[string]interface{}) interface{} {
	ec := NewEnderChest()
	ec.Facing = c.Facing
	return ec
}

// EncodeItem ...
func (EnderChest) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_chest", 0
}

// EncodeBlock ...
func (c EnderChest) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:ender_chest", map[string]any{"minecraft:cardinal_direction": c.Facing.String()}
}

// allEnderChests ...
func allEnderChests() (chests []world.Block) {
	for _, direction := range cube.Directions() {
		chests = append(chests, EnderChest{Facing: direction})
	}
	return
}
