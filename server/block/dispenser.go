package block

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Dispenser is a nine-slot container that dispenses an item when activated by redstone.
type Dispenser struct {
	solid
	sourceWaterDisplacer

	// Facing is the direction items are dispensed towards.
	Facing cube.Face
	// Triggered is whether the dispenser currently receives redstone power.
	Triggered bool
	// CustomName is the custom name displayed when the dispenser is opened.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

var (
	_ world.RedstonePowerConsumer    = Dispenser{}
	_ world.RedstonePowerPostUpdater = Dispenser{}
)

const dispenserDelay = time.Second / 5

// NewDispenser creates an initialised dispenser.
func NewDispenser() Dispenser {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)
	return Dispenser{
		inventory: inventory.New(9, func(slot int, _, stack item.Stack) {
			m.RLock()
			defer m.RUnlock()
			for viewer := range v {
				viewer.ViewSlotChange(slot, stack)
			}
		}),
		viewerMu: m,
		viewers:  v,
	}
}

// Inventory returns the dispenser inventory.
func (d Dispenser) Inventory(*world.Tx, cube.Pos) *inventory.Inventory { return d.inventory }

// WithName returns the dispenser with a custom name.
func (d Dispenser) WithName(a ...any) world.Item {
	d.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return d
}

// AddViewer adds a viewer to the dispenser inventory.
func (d Dispenser) AddViewer(v ContainerViewer, _ *world.Tx, _ cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	d.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the dispenser inventory.
func (d Dispenser) RemoveViewer(v ContainerViewer, _ *world.Tx, _ cube.Pos) {
	d.viewerMu.Lock()
	defer d.viewerMu.Unlock()
	delete(d.viewers, v)
}

// Activate opens the dispenser inventory.
func (Dispenser) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

// RedstonePowerUpdate updates the dispenser's triggered state. In addition to direct power, dispensers accept power at
// the block above them, matching Java Edition's quasi-connectivity behaviour.
func (d Dispenser) RedstonePowerUpdate(pos cube.Pos, tx *world.Tx, power int) (world.Block, bool) {
	powered := power > 0 || tx.RedstonePower(pos.Side(cube.FaceUp)) > 0
	if d.Triggered == powered {
		return d, false
	}
	d.Triggered = powered
	return d, true
}

// RedstonePowerPostUpdate schedules a dispense after an uncancelled rising edge.
func (Dispenser) RedstonePowerPostUpdate(pos cube.Pos, tx *world.Tx, before, after world.Block, _, _ int) {
	beforeDispenser, beforeOK := before.(Dispenser)
	afterDispenser, afterOK := after.(Dispenser)
	if !beforeOK || !afterOK || beforeDispenser.Triggered || !afterDispenser.Triggered {
		return
	}
	// Scheduled block updates are keyed by block state. Queue both states so a short pulse still fires after the
	// delay, while only the state that remains at execution time is run.
	tx.ScheduleBlockUpdate(pos, afterDispenser, dispenserDelay)
	afterDispenser.Triggered = false
	tx.ScheduleBlockUpdate(pos, afterDispenser, dispenserDelay)
}

// ScheduledTick dispenses one item after the activation delay.
func (d Dispenser) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if !d.dispense(pos, tx, r) {
		tx.PlaySound(pos.Vec3Centre(), sound.ClickFail{})
	}
}

func (d Dispenser) dispense(pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	if d.inventory == nil {
		return false
	}
	slots := d.inventory.Slots()
	nonEmpty := make([]int, 0, len(slots))
	for slot, stack := range slots {
		if !stack.Empty() {
			nonEmpty = append(nonEmpty, slot)
		}
	}
	if len(nonEmpty) == 0 {
		return false
	}
	slot := nonEmpty[r.IntN(len(nonEmpty))]
	stack := slots[slot]
	if behaviour, ok := stack.Item().(item.Dispensable); ok {
		ctx := &item.DispenseContext{Rand: r}
		switch behaviour.Dispense(pos, d.Facing, tx, ctx) {
		case item.DispenseSuccess:
			return d.applyDispenseContext(slot, stack, ctx, pos, tx, r)
		case item.DispenseFailure:
			return false
		}
	}
	return d.dropDispensedItem(slot, stack, pos, tx, r)
}

func (d Dispenser) applyDispenseContext(slot int, stack item.Stack, ctx *item.DispenseContext, pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	stack = stack.Damage(ctx.Damage).Grow(-ctx.CountSub)
	if ctx.NewItem.Empty() {
		return d.inventory.SetItem(slot, stack) == nil
	}
	if stack.Empty() {
		return d.inventory.SetItem(slot, ctx.NewItem) == nil
	}
	added, err := d.inventory.AddItem(ctx.NewItem)
	if err != nil {
		create := tx.World().EntityRegistry().Config().Item
		if create == nil {
			return false
		}
		remaining := ctx.NewItem.Grow(added - ctx.NewItem.Count())
		tx.AddEntity(create(dispenserDropOpts(pos, d.Facing, r), remaining))
	}
	return d.inventory.SetItem(slot, stack) == nil
}

func (d Dispenser) dropDispensedItem(slot int, stack item.Stack, pos cube.Pos, tx *world.Tx, r *rand.Rand) bool {
	create := tx.World().EntityRegistry().Config().Item
	if create == nil {
		return false
	}

	opts := dispenserDropOpts(pos, d.Facing, r)
	dropped := stack.Grow(1 - stack.Count())
	if err := d.inventory.SetItem(slot, stack.Grow(-1)); err != nil {
		return false
	}
	tx.AddEntity(create(opts, dropped))
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
	return true
}

// dispenserDirection returns the unit vector pointing out of the front of a dispenser with the face passed. An invalid
// face yields a zero vector.
func dispenserDirection(face cube.Face) mgl64.Vec3 {
	return cube.Pos{}.Side(face).Vec3()
}

// dispenserDropOpts returns the spawn options for an item dropped out of the front of a dispenser.
func dispenserDropOpts(pos cube.Pos, facing cube.Face, r *rand.Rand) world.EntitySpawnOpts {
	direction := dispenserDirection(facing)
	return world.EntitySpawnOpts{
		Position: pos.Vec3Centre().Add(direction.Mul(0.7)),
		Velocity: direction.Mul(0.25).Add(mgl64.Vec3{r.Float64()*0.04 - 0.02, 0.1, r.Float64()*0.04 - 0.02}),
	}
}

// UseOnBlock places the dispenser facing the player.
func (d Dispenser) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, d)
	if !used {
		return false
	}
	d = NewDispenser()
	d.Facing = calculateFace(user, pos)
	place(tx, pos, d, user, ctx)
	return placed(ctx)
}

// BreakInfo returns the dispenser's breaking properties.
func (d Dispenser) BreakInfo() BreakInfo {
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(Dispenser{})).withBlastResistance(17.5).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		for _, stack := range d.Inventory(tx, pos).Clear() {
			dropItem(tx, stack, pos.Vec3())
		}
	})
}

// DecodeNBT decodes dispenser block-entity data.
func (d Dispenser) DecodeNBT(data map[string]any) any {
	facing, triggered := d.Facing, d.Triggered
	d = NewDispenser()
	d.Facing, d.Triggered = facing, triggered
	d.CustomName = nbtconv.String(data, "CustomName")
	nbtconv.InvFromNBT(d.inventory, nbtconv.Slice(data, "Items"))
	return d
}

// EncodeNBT encodes dispenser block-entity data.
func (d Dispenser) EncodeNBT() map[string]any {
	if d.inventory == nil {
		facing, triggered, customName := d.Facing, d.Triggered, d.CustomName
		d = NewDispenser()
		d.Facing, d.Triggered, d.CustomName = facing, triggered, customName
	}
	m := map[string]any{"Items": nbtconv.InvToNBT(d.inventory), "id": "Dispenser"}
	if d.CustomName != "" {
		m["CustomName"] = d.CustomName
	}
	return m
}

// EncodeBlock encodes the dispenser block state.
func (d Dispenser) EncodeBlock() (string, map[string]any) {
	return "minecraft:dispenser", map[string]any{"facing_direction": int32(d.Facing), "triggered_bit": boolByte(d.Triggered)}
}

// EncodeItem encodes the dispenser item.
func (Dispenser) EncodeItem() (string, int16) { return "minecraft:dispenser", 0 }

func allDispensers() (blocks []world.Block) {
	for _, f := range cube.Faces() {
		blocks = append(blocks, Dispenser{Facing: f}, Dispenser{Facing: f, Triggered: true})
	}
	return
}
