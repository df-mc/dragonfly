package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
	"time"
)

// Chest is a container block which may be used to store items. Chests may also be paired to create a bigger
// single container.
// The empty value of Chest is not valid. It must be created using block.NewChest().
type Chest struct {
	chest
	transparent
	bass

	// Facing is the direction that the chest is facing.
	Facing cube.Direction
	// CustomName is the custom name of the chest. This name is displayed when the chest is opened, and may
	// include colour codes.
	CustomName string

	pairPos [2]int32
	paired  bool

	doubleInventory *inventory.Inventory
	inventory       *inventory.Inventory
	viewerMu        *sync.RWMutex
	viewers         map[ContainerViewer]struct{}
}

// NewChest creates a new initialised chest. The inventory is properly initialised.
func NewChest() Chest {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)
	return Chest{
		inventory: inventory.New(27, func(slot int, item item.Stack) {
			m.RLock()
			defer m.RUnlock()
			for viewer := range v {
				viewer.ViewSlotChange(slot, item)
			}
		}),
		viewerMu: m,
		viewers:  v,
	}
}

// Inventory returns the inventory of the chest. The size of the inventory will be 27 or 54, depending on
// whether the chest is single or double.
func (c Chest) Inventory(w *world.World, pos cube.Pos) *inventory.Inventory {
	if p, _, paired := c.pair(w, pos); paired {
		return p.doubleInventory
	}
	return c.inventory
}

// PreBreak ...
func (c Chest) PreBreak(pos cube.Pos, w *world.World, _ item.User) world.Block {
	if p, pairPos, paired := c.pair(w, pos); paired {
		p.paired = false
		w.SetBlock(pairPos, p, nil)
	}
	c.paired = false
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
	viewers := w.Viewers(pos.Vec3())
	for _, v := range viewers {
		v.ViewBlockAction(pos, OpenAction{})
	}
	if _, pairPos, paired := c.pair(w, pos); paired {
		for _, v := range viewers {
			v.ViewBlockAction(pairPos, OpenAction{})
		}
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestOpen{})
}

// close closes the chest, displaying the animation and playing a sound.
func (c Chest) close(w *world.World, pos cube.Pos) {
	viewers := w.Viewers(pos.Vec3())
	for _, v := range viewers {
		v.ViewBlockAction(pos, CloseAction{})
	}
	if _, pairPos, paired := c.pair(w, pos); paired {
		for _, v := range viewers {
			v.ViewBlockAction(pairPos, CloseAction{})
		}
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestClose{})
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c Chest) AddViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	defer c.viewerMu.Unlock()
	viewing := len(c.viewers)
	c.viewers[v] = struct{}{}
	if p, _, paired := c.pair(w, pos); paired {
		p.viewerMu.Lock()
		defer p.viewerMu.Unlock()
		viewing += len(p.viewers)
		p.viewers[v] = struct{}{}
	}
	if viewing == 0 {
		c.open(w, pos)
	}
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to
// it.
func (c Chest) RemoveViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	defer c.viewerMu.Unlock()
	delete(c.viewers, v)
	remaining := len(c.viewers)
	if p, _, paired := c.pair(w, pos); paired {
		p.viewerMu.Lock()
		defer p.viewerMu.Unlock()
		delete(p.viewers, v)
		remaining += len(p.viewers)
	}
	if remaining == 0 {
		c.close(w, pos)
	}
}

// Activate ...
func (c Chest) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		if d, ok := w.Block(pos.Side(cube.FaceUp)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
			opener.OpenBlockContainer(pos)
		}
		return true
	}
	return false
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
	for _, sidePos := range []cube.Pos{pos.Side(c.Facing.RotateLeft().Face()), pos.Side(c.Facing.RotateRight().Face())} {
		if otherC, ok := w.Block(sidePos).(Chest); ok && c.Facing == otherC.Facing && !otherC.paired {
			//noinspection GoAssignmentToReceiver
			c = c.pairWith(sidePos)
			otherC = otherC.pairWith(pos)
			w.SetBlock(sidePos, otherC, nil)
			break
		}
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c Chest) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, simpleDrops(item.NewStack(c, 1)))
}

// FuelInfo ...
func (Chest) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// FlammabilityInfo ...
func (c Chest) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// DecodeNBT ...
func (c Chest) DecodeNBT(data map[string]any) any {
	facing := c.Facing
	//noinspection GoAssignmentToReceiver
	c = NewChest()
	c.Facing = facing
	c.CustomName = nbtconv.Map[string](data, "CustomName")
	c.pairPos[0], c.paired = nbtconv.TryMap[int32](data, "pairx")
	c.pairPos[1], c.paired = nbtconv.TryMap[int32](data, "pairz")
	nbtconv.InvFromNBT(c.inventory, nbtconv.Map[[]any](data, "Items"))
	return c
}

// EncodeNBT ...
func (c Chest) EncodeNBT() map[string]any {
	if c.inventory == nil {
		facing, customName := c.Facing, c.CustomName
		//noinspection GoAssignmentToReceiver
		c = NewChest()
		c.Facing, c.CustomName = facing, customName
	}
	m := map[string]any{
		"Items": nbtconv.InvToNBT(c.inventory),
		"id":    "Chest",
	}
	if c.CustomName != "" {
		m["CustomName"] = c.CustomName
	}
	if c.paired {
		m["pairx"] = c.pairPos[0]
		m["pairz"] = c.pairPos[1]
	}
	return m
}

// preferPair returns true if this chest is the first side of the double-chest.
func (c Chest) preferPair(pos cube.Pos) bool {
	i := int(c.pairPos[0]) + (int(c.pairPos[1]) << 15)
	j := pos.X() + (pos.Z() << 15)
	return i > j
}

// pairWith pairs this chest with the given chest position.
func (c Chest) pairWith(pos cube.Pos) Chest {
	c.pairPos, c.paired = [2]int32{int32(pos.X()), int32(pos.Z())}, true
	return c
}

// pair returns the paired chest of this chest.
func (c Chest) pair(w *world.World, pos cube.Pos) (Chest, cube.Pos, bool) {
	if !c.paired {
		return Chest{}, cube.Pos{}, false
	}
	pairPos := cube.Pos{int(c.pairPos[0]), pos.Y(), int(c.pairPos[1])}
	p := w.Block(pairPos).(Chest)
	if c.doubleInventory == nil || p.doubleInventory == nil {
		first, last := c.inventory, p.inventory
		if c.preferPair(pos) {
			first, last = last, first
		}

		size := first.Size() + last.Size()
		offset := size / 2

		merged := inventory.New(size, func(slot int, item item.Stack) {
			if slot < offset {
				_ = first.SetItemSilently(slot, item)
			} else {
				_ = last.SetItemSilently(slot-offset, item)
			}

			c.viewerMu.RLock()
			defer c.viewerMu.RUnlock()
			for viewer := range c.viewers {
				viewer.ViewSlotChange(slot, item)
			}
		})
		for i := 0; i < merged.Size(); i++ {
			if i < offset {
				it, _ := first.Item(i)
				_ = merged.SetItemSilently(i, it)
				continue
			}
			it, _ := last.Item(i - offset)
			_ = merged.SetItemSilently(i, it)
		}

		c.doubleInventory, p.doubleInventory = merged, merged
		w.SetBlock(pairPos, p, nil)
		w.SetBlock(pos, c, nil)
	}
	return p, pairPos, true
}

// EncodeItem ...
func (Chest) EncodeItem() (name string, meta int16) {
	return "minecraft:chest", 0
}

// EncodeBlock ...
func (c Chest) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:chest", map[string]any{"facing_direction": 2 + int32(c.Facing)}
}

// allChests ...
func allChests() (chests []world.Block) {
	for _, direction := range cube.Directions() {
		chests = append(chests, Chest{Facing: direction})
	}
	return
}
