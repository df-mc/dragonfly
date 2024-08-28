package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"strings"
	"sync"
)

// Hopper is a low-capacity storage block that can be used to collect item entities directly above it, as well as to
// transfer items into and out of other containers.
type Hopper struct {
	transparent
	sourceWaterDisplacer

	// Facing is the direction the hopper is facing.
	Facing cube.Face
	// Powered is whether the hopper is powered or not. If the hopper is powered it will be locked and will stop
	// moving items into or out of itself.
	Powered bool
	// CustomName is the custom name of the hopper. This name is displayed when the hopper is opened, and may include
	// colour codes.
	CustomName string

	// LastTick is the last world tick that the hopper was ticked.
	LastTick int64
	// TransferCooldown is the duration in ticks until the hopper can transfer items again.
	TransferCooldown int64
	// CollectCooldown is the duration in ticks until the hopper can collect items again.
	CollectCooldown int64

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// NewHopper creates a new initialised hopper. The inventory is properly initialised.
func NewHopper() Hopper {
	m := new(sync.RWMutex)
	v := make(map[ContainerViewer]struct{}, 1)
	return Hopper{
		inventory: inventory.New(5, func(slot int, _, item item.Stack) {
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

// Model ...
func (Hopper) Model() world.BlockModel {
	return model.Hopper{}
}

// SideClosed ...
func (Hopper) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (h Hopper) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(h)).withBlastResistance(24).withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
		for _, i := range h.Inventory(w, pos).Clear() {
			dropItem(w, i, pos.Vec3())
		}
	})
}

// Inventory returns the inventory of the hopper.
func (h Hopper) Inventory(*world.World, cube.Pos) *inventory.Inventory {
	return h.inventory
}

// WithName returns the hopper after applying a specific name to the block.
func (h Hopper) WithName(a ...any) world.Item {
	h.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return h
}

// AddViewer adds a viewer to the hopper, so that it is updated whenever the inventory of the hopper is changed.
func (h Hopper) AddViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	h.viewerMu.Lock()
	defer h.viewerMu.Unlock()
	h.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the hopper, so that slot updates in the inventory are no longer sent to it.
func (h Hopper) RemoveViewer(v ContainerViewer, _ *world.World, _ cube.Pos) {
	h.viewerMu.Lock()
	defer h.viewerMu.Unlock()
	delete(h.viewers, v)
}

// Activate ...
func (Hopper) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

// UseOnBlock ...
func (h Hopper) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, h)
	if !used {
		return false
	}

	//noinspection GoAssignmentToReceiver
	h = NewHopper()
	h.Facing = cube.FaceDown
	if h.Facing != face {
		h.Facing = face.Opposite()
	}

	place(w, pos, h, user, ctx)
	return placed(ctx)
}

// Tick ...
func (h Hopper) Tick(currentTick int64, pos cube.Pos, w *world.World) {
	h.TransferCooldown--
	h.CollectCooldown--
	h.LastTick = currentTick

	if !h.Powered && h.TransferCooldown <= 0 {
		inserted := h.insertItem(pos, w)
		extracted := h.extractItem(pos, w)
		if inserted || extracted {
			h.TransferCooldown = 8
		}
	}

	w.SetBlock(pos, h, nil)
}

// HopperInsertable represents a block that can have its contents inserted into by a hopper.
type HopperInsertable interface {
	// InsertItem handles the insert logic for that block.
	InsertItem(h Hopper, pos cube.Pos, w *world.World) bool
}

// insertItem inserts an item into a block that can receive contents from the hopper.
func (h Hopper) insertItem(pos cube.Pos, w *world.World) bool {
	destPos := pos.Side(h.Facing)
	dest := w.Block(destPos)

	if e, ok := dest.(HopperInsertable); ok {
		return e.InsertItem(h, pos.Side(h.Facing), w)
	}

	if container, ok := dest.(Container); ok {
		for sourceSlot, sourceStack := range h.inventory.Slots() {
			if sourceStack.Empty() {
				continue
			}

			_, err := container.Inventory(w, pos).AddItem(sourceStack.Grow(-sourceStack.Count() + 1))
			if err != nil {
				// The destination is full.
				return false
			}

			_ = h.inventory.SetItem(sourceSlot, sourceStack.Grow(-1))

			if hopper, ok := dest.(Hopper); ok {
				hopper.TransferCooldown = 8
				w.SetBlock(destPos, hopper, nil)
			}

			return true
		}
	}
	return false
}

// HopperExtractable represents a block that can have its contents extracted by a hopper.
type HopperExtractable interface {
	// ExtractItem handles the extract logic for that block.
	ExtractItem(h Hopper, pos cube.Pos, w *world.World) bool
}

// extractItem extracts an item from a container into the hopper.
func (h Hopper) extractItem(pos cube.Pos, w *world.World) bool {
	originPos := pos.Side(cube.FaceUp)
	origin := w.Block(originPos)

	if e, ok := origin.(HopperExtractable); ok {
		return e.ExtractItem(h, pos, w)
	}

	if containerOrigin, ok := origin.(Container); ok {
		for slot, stack := range containerOrigin.Inventory(w, originPos).Slots() {
			if stack.Empty() {
				// We don't have any items to extract.
				continue
			}

			_, err := h.inventory.AddItem(stack.Grow(-stack.Count() + 1))
			if err != nil {
				// The hopper is full.
				continue
			}

			_ = containerOrigin.Inventory(w, originPos).SetItem(slot, stack.Grow(-1))

			if hopper, ok := origin.(Hopper); ok {
				hopper.TransferCooldown = 8
				w.SetBlock(originPos, hopper, nil)
			}

			return true
		}
	}
	return false
}

// EncodeItem ...
func (Hopper) EncodeItem() (name string, meta int16) {
	return "minecraft:hopper", 0
}

// EncodeBlock ...
func (h Hopper) EncodeBlock() (string, map[string]any) {
	return "minecraft:hopper", map[string]any{
		"facing_direction": int32(h.Facing),
		"toggle_bit":       h.Powered,
	}
}

// EncodeNBT ...
func (h Hopper) EncodeNBT() map[string]any {
	if h.inventory == nil {
		facing, powered, customName := h.Facing, h.Powered, h.CustomName
		//noinspection GoAssignmentToReceiver
		h = NewHopper()
		h.Facing, h.Powered, h.CustomName = facing, powered, customName
	}
	m := map[string]any{
		"Items":            nbtconv.InvToNBT(h.inventory),
		"TransferCooldown": int32(h.TransferCooldown),
		"id":               "Hopper",
	}
	if h.CustomName != "" {
		m["CustomName"] = h.CustomName
	}
	return m
}

// DecodeNBT ...
func (h Hopper) DecodeNBT(data map[string]any) any {
	facing, powered := h.Facing, h.Powered
	//noinspection GoAssignmentToReceiver
	h = NewHopper()
	h.Facing = facing
	h.Powered = powered
	h.CustomName = nbtconv.String(data, "CustomName")
	h.TransferCooldown = int64(nbtconv.Int32(data, "TransferCooldown"))
	nbtconv.InvFromNBT(h.inventory, nbtconv.Slice(data, "Items"))
	return h
}

// allHoppers ...
func allHoppers() (hoppers []world.Block) {
	for _, f := range cube.Faces() {
		hoppers = append(hoppers, Hopper{Facing: f})
		hoppers = append(hoppers, Hopper{Facing: f, Powered: true})
	}
	return hoppers
}
