package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/slices"
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

	paired  bool
	pairPos cube.Pos

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
			defer m.RUnlock()
			for _, viewer := range *v {
				viewer.ViewSlotChange(slot, item)
			}
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

// PreBreak ...
func (c Chest) PreBreak(pos cube.Pos, w *world.World, _ item.User) world.Block {
	if c.paired {
		c.unpair()
	}
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
		if c.paired {
			v.ViewBlockAction(c.pairPos, OpenAction{})
		}
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestOpen{})
}

// close closes the chest, displaying the animation and playing a sound.
func (c Chest) close(w *world.World, pos cube.Pos) {
	viewers := w.Viewers(pos.Vec3())
	for _, v := range viewers {
		v.ViewBlockAction(pos, CloseAction{})
		if c.paired {
			v.ViewBlockAction(c.pairPos, CloseAction{})
		}
	}
	w.PlaySound(pos.Vec3Centre(), sound.ChestClose{})
}

// AddViewer adds a viewer to the chest, so that it is updated whenever the inventory of the chest is changed.
func (c Chest) AddViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	defer c.viewerMu.Unlock()
	viewing := len(*c.viewers)
	*c.viewers = append(*c.viewers, v)
	if viewing == 0 {
		c.open(w, pos)
	}
}

// RemoveViewer removes a viewer from the chest, so that slot updates in the inventory are no longer sent to
// it.
func (c Chest) RemoveViewer(v ContainerViewer, w *world.World, pos cube.Pos) {
	c.viewerMu.Lock()
	defer c.viewerMu.Unlock()
	i := sliceutil.Index(*c.viewers, v)
	if i == -1 {
		return
	}
	*c.viewers = slices.Delete(*c.viewers, i, i+1)
	if len(*c.viewers) == 0 {
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
	for _, dir := range []cube.Direction{c.Facing.RotateLeft(), c.Facing.RotateRight()} {
		sidePos := pos.Side(dir.Face())
		if ch, pair, ok := c.pair(w, pos, sidePos); ok {
			place(w, pos, ch, user, ctx)
			place(w, sidePos, pair, user, ctx)
			return placed(ctx)
		}
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c Chest) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(c))
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
func (c Chest) DecodeNBT(pos cube.Pos, _ *world.World, data map[string]any) any {
	facing := c.Facing
	//noinspection GoAssignmentToReceiver
	c = NewChest()
	c.Facing = facing
	c.CustomName = nbtconv.Map[string](data, "CustomName")

	pairX, ok := nbtconv.TryMap[int32](data, "pairx")
	pairZ, ok2 := nbtconv.TryMap[int32](data, "pairz")
	if ok && ok2 {
		c.paired = true
		c.pairPos = cube.Pos{int(pairX), pos.Y(), int(pairZ)}
	}
	nbtconv.InvFromNBT(c.inventory, nbtconv.Map[[]any](data, "Items"))
	return c
}

// EncodeNBT ...
func (c Chest) EncodeNBT(cube.Pos, *world.World) map[string]any {
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
		m["pairx"] = int32(c.pairPos[0])
		m["pairz"] = int32(c.pairPos[2])
	}
	return m
}

// pair pairs this chest with the given chest position.
func (c Chest) pair(w *world.World, pos, pairPos cube.Pos) (ch, pair Chest, ok bool) {
	pair, ok = w.Block(pairPos).(Chest)
	if !ok || c.Facing != pair.Facing || pair.paired {
		return c, pair, false
	}
	c.pairPos, c.paired = pairPos, true
	pair.pairPos, pair.paired = pos, true

	left, right := c.inventory, pair.inventory
	if pos.Side(c.Facing.RotateLeft().Face()) == c.pairPos {
		left, right = right, left
	}

	m := new(sync.RWMutex)
	v := new([]ContainerViewer)
	c.viewerMu, pair.viewerMu = m, m
	c.viewers, pair.viewers = v, v
	double := inventory.New(54, func(slot int, item item.Stack) {
		m.RLock()
		defer m.RUnlock()
		for _, viewer := range *v {
			viewer.ViewSlotChange(slot, item)
		}
	})
	for i, it := range append(left.Slots(), right.Slots()...) {
		_ = double.SetItem(i, it)
	}

	c.inventory, pair.inventory = double, double
	return c, pair, true
}

// unpair ...
// TODO: Proper unpairing logic.
func (c Chest) unpair() (ch, pair Chest, ok bool) {
	return Chest{}, Chest{}, false
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
