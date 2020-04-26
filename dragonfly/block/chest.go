package block

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/action"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/internal/nbtconv"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/inventory"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl32"
	"strings"
	"sync"
)

// Chest is a container block which may be used to store items. Chests may also be paired to create a bigger
// single container.
// The empty value of Chest is not valid. It must be created using item.NewChest().
type Chest struct {
	// Facing is the direction that the chest is facing.
	Facing world.Face
	// CustomName is the custom name of the chest. This name is displayed when the chest is opened, and may
	// include colour codes.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   *[]ContainerViewer
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

// open opens the chest, displaying the animation and playing a sound.
func (c Chest) open(w *world.World, pos world.BlockPos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, action.Open{})
	}
	w.PlaySound(pos.Vec3().Add(mgl32.Vec3{0.5, 0.5, 0.5}), sound.ChestOpen{})
}

// close closes the chest, displaying the animation and playing a sound.
func (c Chest) close(w *world.World, pos world.BlockPos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, action.Close{})
	}
	w.PlaySound(pos.Vec3().Add(mgl32.Vec3{0.5, 0.5, 0.5}), sound.ChestClose{})
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c Chest) AddViewer(v ContainerViewer, w *world.World, pos world.BlockPos) {
	c.viewerMu.Lock()
	if len(*c.viewers) == 0 {
		c.open(w, pos)
	}
	*c.viewers = append(*c.viewers, v)
	c.viewerMu.Unlock()
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to
// it.
func (c Chest) RemoveViewer(v ContainerViewer, w *world.World, pos world.BlockPos) {
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
func (c Chest) Activate(pos world.BlockPos, _ world.Face, _ *world.World, e world.Entity) {
	if opener, ok := e.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
	}
}

// UseOnBlock ...
func (c Chest) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl32.Vec3, s *item.Stack, w *world.World, user item.User) {
	if replaceable(w, pos.Side(face), c) {
		if c.inventory == nil {
			//noinspection GoAssignmentToReceiver
			c = NewChest()
		}
		c.Facing = user.Facing().Opposite()
		w.PlaceBlock(pos.Side(face), c)
		*s = (*s).Grow(-1)
	}
}

// BreakInfo ...
func (c Chest) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2.5,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(append(c.inventory.Contents(), item.NewStack(c, 1))...),
	}
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
func (Chest) EncodeItem() (id int32, meta int16) {
	return 54, 0
}

// EncodeBlock ...
func (c Chest) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:chest", map[string]interface{}{"facing_direction": int32(c.Facing)}
}
