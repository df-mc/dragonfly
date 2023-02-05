package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Dropper is a low-capacity storage block that can eject its contents into the world or into other containers when
// given a redstone signal.
type Dropper struct {
	solid

	// Facing is the direction the dropper is facing.
	Facing cube.Face
	// Powered is whether the dropper is powered or not.
	Powered bool
	// CustomName is the custom name of the dropper. This name is displayed when the dropper is opened, and may include
	// colour codes.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// NewDropper creates a new initialised dropper. The inventory is properly initialised.
func NewDropper() Dropper {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)
	return Dropper{
		inventory: inventory.New(9, func(slot int, _, item item.Stack) {
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

// BreakInfo ...
func (d Dropper) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(d))
}

// Inventory returns the inventory of the dropper.
func (d Dropper) Inventory() *inventory.Inventory {
	return d.inventory
}

// WithName returns the dropper after applying a specific name to the block.
func (d Dropper) WithName(a ...any) world.Item {
	d.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return d
}

// AddViewer adds a viewer to the dropper, so that it is updated whenever the inventory of the dropper is changed.
func (d Dropper) AddViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	d.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the dropper, so that slot updates in the inventory are no longer sent to it.
func (d Dropper) RemoveViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	delete(d.viewers, v)
}

// Activate ...
func (Dropper) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User, _ *item.UseContext) bool {
	if o, ok := u.(ContainerOpener); ok {
		o.OpenBlockContainer(pos)
		return true
	}
	return false
}

// UseOnBlock ...
func (d Dropper) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	//noinspection GoAssignmentToReceiver
	d = NewDropper()
	d.Facing = calculateAnySidedFace(user, pos, true)

	place(w, pos, d, user, ctx)
	return placed(ctx)
}

// RedstoneUpdate ...
func (d Dropper) RedstoneUpdate(pos cube.Pos, w *world.World) {
	powered := receivedRedstonePower(pos, w)
	if powered == d.Powered {
		return
	}

	d.Powered = powered
	w.SetBlock(pos, d, nil)
	if d.Powered {
		w.ScheduleBlockUpdate(pos, time.Millisecond*200)
	}
}

// ScheduledTick ...
func (d Dropper) ScheduledTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	slot, ok := d.randomSlotFromInventory(r)
	if !ok {
		w.PlaySound(pos.Vec3Centre(), sound.DispenseFail{})
		return
	}

	it, _ := d.Inventory().Item(slot)
	if c, ok := w.Block(pos.Side(d.Facing)).(Container); ok {
		if _, err := c.Inventory().AddItem(it.Grow(-it.Count() + 1)); err != nil {
			return
		}
		_ = d.Inventory().SetItem(slot, it.Grow(-1))
		return
	}

	_ = d.Inventory().SetItem(slot, it.Grow(-1))

	dist := r.Float64()/10 + 0.2
	sourcePos := pos.Vec3Centre().Add(cube.Pos{}.Side(d.Facing).Vec3().Mul(0.7))

	xOffset, zOffset := 0.0, 0.0
	if axis := d.Facing.Axis(); axis == cube.X {
		xOffset = 1.0
	} else if axis == cube.Z {
		zOffset = 1.0
	}

	xMultiplier, zMultiplier := -1.0, -1.0
	if d.Facing.Positive() {
		xMultiplier, zMultiplier = 1.0, 1.0
	}

	w.PlaySound(sourcePos, sound.Dispense{})
	w.AddParticle(sourcePos, particle.Dispense{})
	w.AddEntity(w.EntityRegistry().Config().Item(
		it.Grow(-it.Count()+1),
		sourcePos,
		mgl64.Vec3{
			(r.Float64()*2-1)*6*0.0075 + xOffset*xMultiplier*dist,
			(r.Float64()*2-1)*6*0.0075 + 0.2,
			(r.Float64()*2-1)*6*0.0075 + zOffset*zMultiplier*dist,
		},
	))
}

// randomSlotFromInventory returns a random slot from the inventory of the dropper. If the inventory is empty, the
// second return value is false.
func (d Dropper) randomSlotFromInventory(r *rand.Rand) (int, bool) {
	slots := make([]int, 0, d.inventory.Size())
	for slot, it := range d.inventory.Slots() {
		if !it.Empty() {
			slots = append(slots, slot)
		}
	}
	if len(slots) == 0 {
		return 0, false
	}
	return slots[r.Intn(len(slots))], true
}

// EncodeItem ...
func (Dropper) EncodeItem() (name string, meta int16) {
	return "minecraft:dropper", 0
}

// EncodeBlock ...
func (d Dropper) EncodeBlock() (string, map[string]any) {
	return "minecraft:dropper", map[string]any{
		"facing_direction": int32(d.Facing),
		"triggered_bit":    d.Powered,
	}
}

// EncodeNBT ...
func (d Dropper) EncodeNBT() map[string]any {
	if d.inventory == nil {
		facing, powered, customName := d.Facing, d.Powered, d.CustomName
		//noinspection GoAssignmentToReceiver
		d = NewDropper()
		d.Facing, d.Powered, d.CustomName = facing, powered, customName
	}
	m := map[string]any{
		"Items": nbtconv.InvToNBT(d.inventory),
		"id":    "Dropper",
	}
	if d.CustomName != "" {
		m["CustomName"] = d.CustomName
	}
	return m
}

// DecodeNBT ...
func (d Dropper) DecodeNBT(data map[string]any) any {
	facing, powered := d.Facing, d.Powered
	//noinspection GoAssignmentToReceiver
	d = NewDropper()
	d.Facing = facing
	d.Powered = powered
	d.CustomName = nbtconv.String(data, "CustomName")
	nbtconv.InvFromNBT(d.inventory, nbtconv.Slice(data, "Items"))
	return d
}

// allDroppers ...
func allDroppers() (droppers []world.Block) {
	for _, f := range cube.Faces() {
		for _, p := range []bool{false, true} {
			droppers = append(droppers, Dropper{Facing: f, Powered: p})
		}
	}
	return droppers
}
