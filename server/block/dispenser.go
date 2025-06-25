package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

// Dispenser is a low-capacity storage block that can fire projectiles, use certain items or tools or place certain blocks,
// fluids or entities when given a redstone signal. Items that do not have unique dispenser functions are instead ejected into the world.
type Dispenser struct {
	solid

	// Facing is the direction the dispenser is facing.
	Facing cube.Face
	// Powered is whether the dispenser is powered or not.
	Powered bool
	// CustomName is the custom name of the dispenser. This name is displayed when the dispenser is opened, and may include
	// colour codes.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// NewDispenser creates a new initialised dispenser. The inventory is properly initialised.
func NewDispenser() Dispenser {
	d := Dispenser{
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
func (d Dispenser) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		for _, i := range d.Inventory(tx, pos).Clear() {
			dropItem(tx, i, pos.Vec3())
		}
	})
}

// WithName returns the dispenser after applying a specific name to the block.
func (d Dispenser) WithName(a ...any) world.Item {
	d.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return d
}

// Inventory returns the inventory of the dispenser. The size of the inventory will be 9.
func (d Dispenser) Inventory(*world.Tx, cube.Pos) *inventory.Inventory {
	return d.inventory
}

// AddViewer adds a viewer to the dropper, so that it is updated whenever the inventory of the dispenser is changed.
func (d Dispenser) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	d.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the dispenser, so that slot updates in the inventory are no longer sent to
// it.
func (d Dispenser) RemoveViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	if len(d.viewers) == 0 {
		return
	}
	delete(d.viewers, v)
}

// Activate ...
func (d Dispenser) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

// UseOnBlock ...
func (d Dispenser) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, d)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	d = NewDispenser()
	d.Facing = calculateFace(user, pos, true)

	place(tx, pos, d, user, ctx)
	return placed(ctx)
}

// RedstoneUpdate ...
func (d Dispenser) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
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
func (d Dispenser) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	slot, ok := d.randomSlotFromInventory(r)
	if !ok {
		tx.PlaySound(pos.Vec3(), sound.DispenseFail{})
		return
	}

	it, _ := d.Inventory(tx, pos).Item(slot)
	sidePos := pos.Add(cube.Pos{}.Side(d.Facing))

	switch it.Item().(type) {
	case item.FlintAndSteel:
		if t, ok := tx.Block(sidePos).(TNT); ok {
			t.Ignite(pos, tx, nil)
		} else if _, ok := tx.Block(sidePos).(Air); ok {
			tx.SetBlock(sidePos, Fire{}, nil)
		}
		it.Damage(1)
	case item.GlassBottle:
		if _, ok := tx.Block(sidePos).(Water); ok {
			d.Inventory(tx, sidePos).AddItem(item.NewStack(item.Potion{Type: potion.Water()}, 1))
		}
		_ = d.Inventory(tx, pos).SetItem(slot, it.Grow(-1))
	case TNT:
		tx.PlaySound(sidePos.Vec3Centre(), sound.TNT{})
		opts := world.EntitySpawnOpts{Position: sidePos.Vec3Centre()}
		tx.AddEntity(tx.World().EntityRegistry().Config().TNT(opts, time.Second*4))

		_ = d.Inventory(tx, pos).SetItem(slot, it.Grow(-1))
	case item.BoneMeal:
		if b, ok := tx.Block(sidePos).(item.BoneMealAffected); ok {
			b.BoneMeal(pos, tx)
		}
	default:
		create := tx.World().EntityRegistry().Config().Item

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

		opts := world.EntitySpawnOpts{Position: sourcePos, Velocity: mgl64.Vec3{
			(r.Float64()*2-1)*6*0.0075 + xOffset*xMultiplier*dist,
			(r.Float64()*2-1)*6*0.0075 + 0.2,
			(r.Float64()*2-1)*6*0.0075 + zOffset*zMultiplier*dist,
		}}

		tx.AddEntity(create(opts, it.Grow(-it.Count()+1)))
		_ = d.Inventory(tx, pos).SetItem(slot, it.Grow(-1))
	}

	tx.AddParticle(pos.Vec3(), particle.Dispense{})
	tx.PlaySound(pos.Vec3(), sound.Dispense{})
}

// randomSlotFromInventory returns a random slot from the inventory of the dispenser. If the inventory is empty, the
// second return value is false.
func (d Dispenser) randomSlotFromInventory(r *rand.Rand) (int, bool) {
	slots := make([]int, 0, d.inventory.Size())
	for slot, it := range d.inventory.Slots() {
		if !it.Empty() {
			slots = append(slots, slot)
		}
	}
	if len(slots) == 0 {
		return 0, false
	}
	return slots[r.IntN(len(slots))], true
}

// EncodeItem ...
func (Dispenser) EncodeItem() (name string, meta int16) {
	return "minecraft:dispenser", 0
}

// EncodeBlock ...
func (d Dispenser) EncodeBlock() (string, map[string]any) {
	return "minecraft:dispenser", map[string]any{
		"facing_direction": int32(d.Facing),
		"triggered_bit":    d.Powered,
	}
}

// EncodeNBT ...
func (d Dispenser) EncodeNBT() map[string]any {
	if d.inventory == nil {
		facing, powered, customName := d.Facing, d.Powered, d.CustomName
		//noinspection GoAssignmentToReceiver
		d = NewDispenser()
		d.Facing, d.Powered, d.CustomName = facing, powered, customName
	}
	m := map[string]any{
		"Items": nbtconv.InvToNBT(d.inventory),
		"id":    "Dispenser",
	}
	if d.CustomName != "" {
		m["CustomName"] = d.CustomName
	}
	return m
}

// DecodeNBT ...
func (d Dispenser) DecodeNBT(data map[string]any) any {
	facing, powered := d.Facing, d.Powered
	//noinspection GoAssignmentToReceiver
	d = NewDispenser()
	d.Facing = facing
	d.Powered = powered
	d.CustomName = nbtconv.String(data, "CustomName")
	nbtconv.InvFromNBT(d.inventory, nbtconv.Slice(data, "Items"))
	return d
}

// allDispensers ...
func allDispensers() (dispensers []world.Block) {
	for _, f := range cube.Faces() {
		for _, p := range []bool{false, true} {
			dispensers = append(dispensers, Dispenser{Facing: f, Powered: p})
		}
	}
	return dispensers
}
