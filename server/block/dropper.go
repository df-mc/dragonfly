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
	"math/rand/v2"
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
	d := Dropper{
		viewerMu: new(sync.RWMutex),
		viewers:  make(map[ContainerViewer]struct{}),
	}

	d.inventory = inventory.New(9, func(slot int, before, after item.Stack) {
		d.viewerMu.RLock()
		defer d.viewerMu.RUnlock()
		for viewer := range d.viewers {
			viewer.ViewSlotChange(slot, after)
		}
	})
	return d
}

// BreakInfo ...
func (d Dropper) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		for _, i := range d.Inventory(tx, pos).Clear() {
			dropItem(tx, i, pos.Vec3())
		}
	})
}

// WithName returns the dropper after applying a specific name to the block.
func (d Dropper) WithName(a ...any) world.Item {
	d.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return d
}

// Inventory returns the inventory of the dropper. The size of the inventory will be 9.
func (d Dropper) Inventory(*world.Tx, cube.Pos) *inventory.Inventory {
	return d.inventory
}

// AddViewer adds a viewer to the dropper, so that it is updated whenever the inventory of the dropper is changed.
func (d Dropper) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	d.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the dropper, so that slot updates in the inventory are no longer sent to
// it.
func (d Dropper) RemoveViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	if len(d.viewers) == 0 {
		return
	}
	delete(d.viewers, v)
}

// Activate ...
func (d Dropper) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

// UseOnBlock ...
func (d Dropper) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, d)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	d = NewDropper()
	d.Facing = calculateFace(user, pos, true)

	place(tx, pos, d, user, ctx)
	return placed(ctx)
}

// RedstoneUpdate ...
func (d Dropper) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	powered := receivedRedstonePower(pos, tx)
	if powered == d.Powered {
		return
	}

	d.Powered = powered
	tx.SetBlock(pos, d, nil)
	if d.Powered {
		tx.ScheduleBlockUpdate(pos, d, time.Millisecond*200)
	}
}

// ScheduledTick ...
func (d Dropper) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	slot, ok := d.firstSlotAvailableInventory()
	if !ok {
		tx.PlaySound(pos.Vec3(), sound.DispenseFail{})
		return
	}

	it, _ := d.Inventory(tx, pos).Item(slot)
	if c, ok := tx.Block(pos.Side(d.Facing)).(Container); ok {
		if _, err := c.Inventory(tx, pos).AddItem(it.Grow(-it.Count() + 1)); err != nil {
			return
		}
		_ = d.Inventory(tx, pos).SetItem(slot, it.Grow(-1))
		return
	}

	_ = d.Inventory(tx, pos).SetItem(slot, it.Grow(-1))

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

	create := tx.World().EntityRegistry().Config().Item
	opts := world.EntitySpawnOpts{Position: sourcePos, Velocity: mgl64.Vec3{
		(r.Float64()*2-1)*6*0.0075 + xOffset*xMultiplier*dist,
		(r.Float64()*2-1)*6*0.0075 + 0.2,
		(r.Float64()*2-1)*6*0.0075 + zOffset*zMultiplier*dist,
	}}

	tx.AddEntity(create(opts, it.Grow(-it.Count()+1)))
	tx.AddParticle(pos.Vec3(), particle.Dispense{})
	tx.PlaySound(pos.Vec3(), sound.Dispense{})
}

// firstSlotAvailableInventory returns the first available item from the inventory of the dropper. If the inventory is empty, the
// second return value is false.
func (d Dropper) firstSlotAvailableInventory() (int, bool) {
	for slot, it := range d.inventory.Slots() {
		if !it.Empty() {
			return slot, true
		}
	}
	return 0, false
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
