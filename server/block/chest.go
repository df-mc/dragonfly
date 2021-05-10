package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/action"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"strings"
	"sync"
)

// Chest is a container block which may be used to store items. Chests may also be paired to create a bigger
// single container.
// The empty value of Chest is not valid. It must be created using item.NewChest().
// TODO: Redo inventory stuff in here. The inventory should be moved to a different place in world.World so
//  that this block can be hashed properly.
type Chest struct {
	chest
	transparent
	bass

	// Facing is the direction that the chest is facing.
	Facing cube.Direction
	// CustomName is the custom name of the chest. This name is displayed when the chest is opened, and may
	// include colour codes.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   *[]ContainerViewer
}

// FlammabilityInfo ...
func (c Chest) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// NewChest creates a new initialised chest. The inventory is properly initialised.
func NewChest() Chest {
	m := new(sync.RWMutex)
	v := new([]ContainerViewer)
	return Chest{
		inventory: inventory.New(27, func(slot int, item item.Stack) {
			m.RLock()
			for _, viewer := range *v {
				viewer.ViewSlotChange(slot, item)
			}
			m.RUnlock()
		}),
		viewerMu: m,
		viewers:  v,
	}
}

// Inventory returns the inventory of the chest. The size of the inventory will be 27 or 54, depending on
// whether the chest is single or double.
func (c Chest) Inventory() *inventory.Inventory {
	return c.inventory
}

// WithName returns the chest after applying a specific name to the block.
func (c Chest) WithName(a ...interface{}) world.Item {
	c.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return c
}

// CanDisplace ...
func (Chest) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (Chest) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// open opens the chest, displaying the animation and playing a sound.
func (c Chest) open(w *world.World, pos cube.Pos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, action.Open{})
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestOpen{})
}

// close closes the chest, displaying the animation and playing a sound.
func (c Chest) close(w *world.World, pos cube.Pos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, action.Close{})
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestClose{})
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c Chest) AddViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	if len(*c.viewers) == 0 {
		c.open(w, pos)
	}
	*c.viewers = append(*c.viewers, v)
	c.viewerMu.Unlock()
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to
// it.
func (c Chest) RemoveViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	if len(*c.viewers) == 0 {
		c.viewerMu.Unlock()
		return
	}
	newViewers := make([]ContainerViewer, 0, len(*c.viewers)-1)
	for _, viewer := range *c.viewers {
		if viewer != v {
			newViewers = append(newViewers, viewer)
		}
	}
	*c.viewers = newViewers
	if len(*c.viewers) == 0 {
		c.close(w, pos)
	}
	c.viewerMu.Unlock()
}

// Activate ...
func (c Chest) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
	}
}

// UseOnBlock ...
func (c Chest) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, c)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	c = NewChest()
	c.Facing = user.Facing().Opposite()

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c Chest) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, simpleDrops(append(c.inventory.Contents(), item.NewStack(c, 1))...))
}

// Drops returns the drops of the chest. This includes all items held in the inventory and the chest itself.
func (c Chest) Drops() []item.Stack {
	return append(c.inventory.Contents(), item.NewStack(c, 1))
}

// DecodeNBT ...
func (c Chest) DecodeNBT(data map[string]interface{}) interface{} {
	facing := c.Facing
	//noinspection GoAssignmentToReceiver
	c = NewChest()
	c.Facing = facing
	c.CustomName = readString(data, "CustomName")
	nbtconv.InvFromNBT(c.inventory, readSlice(data, "Items"))
	return c
}

// EncodeNBT ...
func (c Chest) EncodeNBT() map[string]interface{} {
	if c.inventory == nil {
		facing, customName := c.Facing, c.CustomName
		//noinspection GoAssignmentToReceiver
		c = NewChest()
		c.Facing, c.CustomName = facing, customName
	}
	m := map[string]interface{}{
		"Items": nbtconv.InvToNBT(c.inventory),
		"id":    "Chest",
	}
	if c.CustomName != "" {
		m["CustomName"] = c.CustomName
	}
	return m
}

// EncodeItem ...
func (Chest) EncodeItem() (name string, meta int16) {
	return "minecraft:chest", 0
}

// EncodeBlock ...
func (c Chest) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:chest", map[string]interface{}{"facing_direction": 2 + int32(c.Facing)}
}

// allChests ...
func allChests() (chests []world.Block) {
	for _, direction := range cube.Directions() {
		chests = append(chests, Chest{Facing: direction})
	}
	return
}
