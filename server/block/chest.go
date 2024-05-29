package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"strings"
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
	sourceWaterDisplacer

	// Facing is the direction that the chest is facing.
	Facing cube.Direction
	// CustomName is the custom name of the chest. This name is displayed when the chest is opened, and may
	// include colour codes.
	CustomName string

	paired       bool
	pairX, pairZ int
	pairInv      *inventory.Inventory

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// NewChest creates a new initialised chest. The inventory is properly initialised.
func NewChest() Chest {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)

	c := Chest{
		viewerMu: m,
		viewers:  v,
	}

	c.inventory = inventory.New(27, func(slot int, _, item item.Stack) {
		c.viewerMu.RLock()
		defer c.viewerMu.RUnlock()
		for viewer := range c.viewers {
			viewer.ViewSlotChange(slot, item)
		}
	})
	return c
}

// Inventory returns the inventory of the chest. The size of the inventory will be 27 or 54, depending on
// whether the chest is single or double.
func (c Chest) Inventory() *inventory.Inventory {
	if c.paired {
		return c.pairInv
	}
	return c.inventory
}

// WithName returns the chest after applying a specific name to the block.
func (c Chest) WithName(a ...any) world.Item {
	c.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return c
}

// SideClosed ...
func (Chest) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// open opens the chest, displaying the animation and playing a sound.
func (c Chest) open(w *world.World, pos cube.Pos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		if c.paired {
			v.ViewBlockAction(c.PairPos(pos), OpenAction{})
		}
		v.ViewBlockAction(pos, OpenAction{})
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestOpen{})
}

// close closes the chest, displaying the animation and playing a sound.
func (c Chest) close(w *world.World, pos cube.Pos) {
	for _, v := range w.Viewers(pos.Vec3()) {
		if c.paired {
			v.ViewBlockAction(c.PairPos(pos), CloseAction{})
		}
		v.ViewBlockAction(pos, CloseAction{})
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestClose{})
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c Chest) AddViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	defer c.viewerMu.Unlock()
	if len(c.viewers) == 0 {
		c.open(w, pos)
	}
	c.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to
// it.
func (c Chest) RemoveViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	defer c.viewerMu.Unlock()
	if len(c.viewers) == 0 {
		return
	}
	delete(c.viewers, v)
	if len(c.viewers) == 0 {
		c.close(w, pos)
	}
}

// Activate ...
func (c Chest) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		if c.paired && c.pairInv == nil {
			if ch, pair, ok := c.pair(w, pos, c.PairPos(pos)); ok {
				w.SetBlock(pos, ch, nil)
				w.SetBlock(c.PairPos(pos), pair, nil)
			}
		}
		if d, ok := w.Block(pos.Side(cube.FaceUp)).(LightDiffuser); ok && d.LightDiffusionLevel() <= 2 {
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
	c.Facing = user.Rotation().Direction().Opposite()

	for _, dir := range []cube.Direction{c.Facing.RotateLeft(), c.Facing.RotateRight()} {
		sidePos := pos.Side(dir.Face())
		if ch, pair, ok := c.pair(w, pos, sidePos); ok {
			place(w, pos, ch, user, ctx)
			w.SetBlock(ch.PairPos(pos), pair, nil)
			return placed(ctx)
		}
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c Chest) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(c)).withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
		if c.paired {
			pairPos := c.PairPos(pos)
			if _, pair, ok := c.unpair(w, pos); ok {
				w.SetBlock(pairPos, pair, nil)
			}
		}
	})
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
	c.CustomName = nbtconv.String(data, "CustomName")

	pairX, ok := data["pairx"]
	pairZ, ok2 := data["pairz"]
	if ok && ok2 {
		c.paired = true
		// TODO: type assertion checks
		c.pairX, c.pairZ = int(pairX.(int32)), int(pairZ.(int32))
		c.pairInv = inventory.New(54, func(slot int, _, item item.Stack) {
			c.viewerMu.RLock()
			defer c.viewerMu.RUnlock()
			for viewer := range c.viewers {
				viewer.ViewSlotChange(slot, item)
			}
		})

		nbtconv.InvFromNBT(c.pairInv, nbtconv.Slice[any](data, "Items"))
	}

		nbtconv.InvFromNBT(c.inventory, nbtconv.Slice[any](data, "Items"))
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
		m["pairx"] = int32(c.pairX)
		m["pairz"] = int32(c.pairZ)
	}
	return m
}

// pair pairs this chest with the given chest position.
func (c Chest) pair(w *world.World, pos, pairPos cube.Pos) (ch, pair Chest, ok bool) {
	pair, ok = w.Block(pairPos).(Chest)
	if !ok || c.Facing != pair.Facing || pair.paired && (pair.pairX != pos[0] || pair.pairZ != pos[2]) {
		return c, pair, false
	}
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{})
	left, right := c.inventory, pair.inventory
	if pos.Side(c.Facing.RotateRight().Face()) == pairPos {
		left, right = right, left
	}
	double := left.Merge(right, func(slot int, _, item item.Stack) {
		if slot < 27 {
			_ = left.SetItem(slot, item)
		} else {
			_ = right.SetItem(slot-27, item)
		}
		m.RLock()
		defer m.RUnlock()
		for viewer := range v {
			viewer.ViewSlotChange(slot, item)
		}
	})

	c.pairX, c.pairZ, c.paired = pairPos[0], pairPos[2], true
	pair.pairX, pair.pairZ, pair.paired = pos[0], pos[2], true
	c.viewerMu, pair.viewerMu = m, m
	c.viewers, pair.viewers = v, v
	c.pairInv, pair.pairInv = double, double
	return c, pair, true
}

// unpair unpairs this chest from the chest it is currently paired with.
func (c Chest) unpair(w *world.World, pos cube.Pos) (ch, pair Chest, ok bool) {
	if !c.paired {
		return c, Chest{}, false
	}

	pair, ok = w.Block(c.PairPos(pos)).(Chest)
	if !ok || c.Facing != pair.Facing || pair.paired && (pair.pairX != pos[0] || pair.pairZ != pos[2]) {
		return c, pair, false
	}

	if len(c.viewers) != 0 {
		c.close(w, pos)
	}

	c.paired, pair.paired = false, false
	c.viewerMu, pair.viewerMu = new(sync.RWMutex), new(sync.RWMutex)
	c.viewers, pair.viewers = make(map[ContainerViewer]struct{}, 1), make(map[ContainerViewer]struct{}, 1)
	c.pairInv, pair.pairInv = nil, nil
	return c, pair, true
}

// PairPos returns the position of the chest that this chest is paired with.
func (c Chest) PairPos(pos cube.Pos) cube.Pos {
	return cube.Pos{c.pairX, pos[1], c.pairZ}
}

// Paired returns whether the chest is paired with another chest.
func (c Chest) Paired() bool {
	return c.paired
}

// EncodeItem ...
func (Chest) EncodeItem() (name string, meta int16) {
	return "minecraft:chest", 0
}

// EncodeBlock ...
func (c Chest) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:chest", map[string]any{"minecraft:cardinal_direction": c.Facing.String()}
}

// allChests ...
func allChests() (chests []world.Block) {
	for _, direction := range cube.Directions() {
		chests = append(chests, Chest{Facing: direction})
	}
	return
}
