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
	"math/rand"
	"strings"
	"sync"
	"time"
)

// ShulkerBox is a dye-able block that stores items. Unlike other blocks, it keeps its contents when broken.
type ShulkerBox struct {
	solid // TODO: I don't think it should be solid
	transparent
	sourceWaterDisplacer
	// Type is the type of shulker box of the block.
	Type ShulkerBoxType
	// Facing is the direction that the shulker box is facing.
	Facing cube.Face
	// CustomName is the custom name of the shulker box. This name is displayed when the shulker box is opened, and may
	// include colour codes.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// NewShulkerBox creates a new initialised shulker box. The inventory is properly initialised.
func NewShulkerBox() ShulkerBox {
	s := ShulkerBox{
		viewerMu: new(sync.RWMutex),
		viewers:  make(map[ContainerViewer]struct{}, 1),
	}

	s.inventory = inventory.New(27, func(slot int, _, after item.Stack) {
		s.viewerMu.RLock()
		defer s.viewerMu.RUnlock()
		for viewer := range s.viewers {
			viewer.ViewSlotChange(slot, after)
		}
	})

	return s
}

// WithName returns the shulker box after applying a specific name to the block.
func (s ShulkerBox) WithName(a ...any) world.Item {
	s.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return s
}

// AddViewer adds a viewer to the shulker box, so that it is updated whenever the inventory of the shulker box is changed.
func (s ShulkerBox) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	s.viewerMu.Lock()
	defer s.viewerMu.Unlock()
	if len(s.viewers) == 0 {
		s.open(tx, pos)
	}

	s.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the shulker box, so that slot updates in the inventory are no longer sent to it.
func (s ShulkerBox) RemoveViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	s.viewerMu.Lock()
	defer s.viewerMu.Unlock()
	if len(s.viewers) == 0 {
		return
	}
	delete(s.viewers, v)
	if len(s.viewers) == 0 {
		s.close(tx, pos)
	}
}

// Inventory returns the inventory of the shulker box.
func (s ShulkerBox) Inventory(tx *world.Tx, pos cube.Pos) *inventory.Inventory {
	return s.inventory
}

// Activate ...
func (s ShulkerBox) Activate(pos cube.Pos, clickedFace cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		if d, ok := tx.Block(pos.Side(s.Facing)).(LightDiffuser); ok && d.LightDiffusionLevel() <= 2 {
			opener.OpenBlockContainer(pos, tx)
		}
		return true
	}

	return false
}

// UseOnBlock ...
func (s ShulkerBox) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}

	if s.inventory == nil {
		typ, customName := s.Type, s.CustomName
		//noinspection GoAssignmentToReceiver
		s = NewShulkerBox()
		s.Type, s.CustomName = typ, customName
	}

	s.Facing = face
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// open opens the shulker box, displaying the animation and playing a sound.
func (s ShulkerBox) open(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, OpenAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxOpen{})
}

// close closes the shulker box, displaying the animation and playing a sound.
func (s ShulkerBox) close(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, CloseAction{})
	}
	tx.ScheduleBlockUpdate(pos, s, time.Millisecond*50*9)
}

// ScheduledTick ...
func (s ShulkerBox) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if len(s.viewers) == 0 {
		tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxClose{})
	}
}

// BreakInfo ...
func (s ShulkerBox) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, pickaxeEffective, oneOf(s))
}

// MaxCount always returns 1.
func (s ShulkerBox) MaxCount() int {
	return 1
}

// EncodeBlock ...
func (s ShulkerBox) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + s.Type.String(), nil
}

// EncodeItem ...
func (s ShulkerBox) EncodeItem() (id string, meta int16) {
	return "minecraft:" + s.Type.String(), 0
}

// DecodeNBT ...
func (s ShulkerBox) DecodeNBT(data map[string]any) any {
	typ := s.Type
	//noinspection GoAssignmentToReceiver
	s = NewShulkerBox()
	s.Type = typ
	nbtconv.InvFromNBT(s.inventory, nbtconv.Slice(data, "Items"))
	s.Facing = cube.Face(nbtconv.Uint8(data, "facing"))
	s.CustomName = nbtconv.String(data, "CustomName")
	return s
}

// EncodeNBT ..
func (s ShulkerBox) EncodeNBT() map[string]any {
	if s.inventory == nil {
		typ, facing, customName := s.Type, s.Facing, s.CustomName
		//noinspection GoAssignmentToReceiver
		s = NewShulkerBox()
		s.Type, s.Facing, s.CustomName = typ, facing, customName
	}

	m := map[string]any{
		"Items":  nbtconv.InvToNBT(s.inventory),
		"id":     "ShulkerBox",
		"facing": uint8(s.Facing),
	}

	if s.CustomName != "" {
		m["CustomName"] = s.CustomName
	}
	return m
}

// allShulkerBoxes ...
func allShulkerBoxes() (shulkerboxes []world.Block) {
	for _, t := range ShulkerBoxTypes() {
		shulkerboxes = append(shulkerboxes, ShulkerBox{Type: t})
	}

	return
}
