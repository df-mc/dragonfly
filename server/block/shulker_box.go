package block

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	shulkerStateClosed int32 = iota
	shulkerStateOpening
	shulkerStateOpened
	shulkerStateClosing
)

// shulkerLidTicks is the number of scheduled ticks between fully closed and fully open.
const shulkerLidTicks int32 = 10

// ShulkerBox is a dye-able block that stores items. Unlike other blocks, it keeps its contents when broken.
type ShulkerBox struct {
	transparent
	sourceWaterDisplacer

	// Colour is the colour of the shulker box. A zero OptionalColour represents
	// the undyed variant (minecraft:undyed_shulker_box).
	Colour item.OptionalColour
	// Facing is the direction that the shulker box is facing.
	Facing cube.Face
	// CustomName is the custom name of the shulker box. This name is displayed when the shulker box is opened, and may
	// include colour codes.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
	// progress is the lid opening progress in [0, 10].
	progress *atomic.Int32
	// animationStatus is the current openness state of the shulker box (whether it's opened, closing, etc.).
	animationStatus *atomic.Int32
}

// NewShulkerBox creates a new initialised shulker box. The inventory is properly initialised.
func NewShulkerBox() ShulkerBox {
	s := ShulkerBox{
		viewerMu:        new(sync.RWMutex),
		viewers:         make(map[ContainerViewer]struct{}, 1),
		progress:        new(atomic.Int32),
		animationStatus: new(atomic.Int32),
	}

	s.inventory = inventory.New(27, func(slot int, _, after item.Stack) {
		s.viewerMu.RLock()
		defer s.viewerMu.RUnlock()
		for viewer := range s.viewers {
			viewer.ViewSlotChange(slot, after)
		}
	})
	s.inventory.SlotValidatorFunc(canStoreInShulkerBox)

	return s
}

// canStoreInShulkerBox rejects nested shulker boxes.
func canStoreInShulkerBox(s item.Stack, _ int) bool {
	if s.Empty() {
		return true
	}
	_, nested := s.Item().(ShulkerBox)
	return !nested
}

func (s ShulkerBox) Model() world.BlockModel {
	return model.Shulker{Facing: s.Facing, Progress: s.progress.Load()}
}

func (s ShulkerBox) WithName(a ...any) world.Item {
	s.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return s
}

func (s ShulkerBox) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	s.viewerMu.Lock()
	defer s.viewerMu.Unlock()
	if len(s.viewers) == 0 {
		s.open(tx, pos)
	}
	s.viewers[v] = struct{}{}
}

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

func (s ShulkerBox) Inventory(*world.Tx, cube.Pos) *inventory.Inventory {
	return s.inventory
}

func (s ShulkerBox) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	opener, ok := u.(ContainerOpener)
	if !ok {
		return false
	}
	if d, ok := tx.Block(pos.Side(s.Facing)).(LightDiffuser); ok && d.LightDiffusionLevel() <= 2 {
		opener.OpenBlockContainer(pos, tx)
	}
	return true
}

func (s ShulkerBox) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}
	s = s.initialised()
	s.Facing = face
	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// initialised lazily populates runtime fields on values created via struct
// literal (e.g. those returned from allShulkerBoxes).
func (s ShulkerBox) initialised() ShulkerBox {
	if s.inventory != nil {
		return s
	}
	n := NewShulkerBox()
	n.Colour, n.Facing, n.CustomName = s.Colour, s.Facing, s.CustomName
	return n
}

// open opens the shulker box, displaying the animation and playing a sound.
func (s ShulkerBox) open(tx *world.Tx, pos cube.Pos) {
	s.animationStatus.Store(shulkerStateOpening)
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, OpenAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxOpen{})
	tx.ScheduleBlockUpdate(pos, s, 0)
}

// close closes the shulker box, displaying the animation and playing a sound.
func (s ShulkerBox) close(tx *world.Tx, pos cube.Pos) {
	s.animationStatus.Store(shulkerStateClosing)
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, CloseAction{})
	}
	tx.ScheduleBlockUpdate(pos, s, 0)
}

func (s ShulkerBox) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	switch s.animationStatus.Load() {
	case shulkerStateClosed:
		s.progress.Store(0)
	case shulkerStateOpening:
		s.progress.Add(1)
		s.pushEntities(pos, tx)
		if s.progress.Load() >= shulkerLidTicks {
			s.progress.Store(shulkerLidTicks)
			s.animationStatus.Store(shulkerStateOpened)
		}
		tx.ScheduleBlockUpdate(pos, s, 0)
	case shulkerStateOpened:
		s.progress.Store(shulkerLidTicks)
	case shulkerStateClosing:
		s.progress.Add(-1)
		if s.progress.Load() <= 0 {
			tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxClose{})
			s.progress.Store(0)
			s.animationStatus.Store(shulkerStateClosed)
		}
		tx.ScheduleBlockUpdate(pos, s, 0)
	}
}

// pushEntities pushes all entities touching the shulker box lid during opening.
func (s ShulkerBox) pushEntities(pos cube.Pos, tx *world.Tx) {
	searchBox := s.physicalBBox().Translate(pos.Vec3()).Grow(0.35)
	for e := range tx.EntitiesWithin(searchBox) {
		s.push(pos, tx, e)
	}
}

// push pushes entities when the shulker box lid is opening.
func (s ShulkerBox) push(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if s.animationStatus.Load() != shulkerStateOpening {
		return
	}
	mover, ok := e.(interface {
		Displace(deltaPos mgl64.Vec3)
	})
	if !ok {
		return
	}
	shulkerBBox := s.physicalBBox().Translate(pos.Vec3())
	entityBBox := e.H().Type().BBox(e).Translate(e.Position())
	if !shulkerBBox.IntersectsWith(entityBBox) {
		return
	}

	// Move the entity out along the lid's facing axis by the penetration depth
	// between the shulker lid box and the entity box.
	delta := shulkerPushDelta(s.Facing, shulkerBBox, entityBBox)
	if delta != (mgl64.Vec3{}) {
		mover.Displace(delta)
	}
}

func (s ShulkerBox) physicalBBox() cube.BBox {
	return (model.Shulker{Facing: s.Facing, Progress: s.progress.Load()}).PhysicalBBox()
}

func shulkerPushDelta(facing cube.Face, shulkerBBox, entityBBox cube.BBox) (delta mgl64.Vec3) {
	switch facing {
	case cube.FaceDown:
		delta[1] = shulkerBBox.Min().Y() - entityBBox.Max().Y()
	case cube.FaceUp:
		delta[1] = shulkerBBox.Max().Y() - entityBBox.Min().Y()
	case cube.FaceEast:
		delta[0] = shulkerBBox.Max().X() - entityBBox.Min().X()
	case cube.FaceWest:
		delta[0] = shulkerBBox.Min().X() - entityBBox.Max().X()
	case cube.FaceSouth:
		delta[2] = shulkerBBox.Max().Z() - entityBBox.Min().Z()
	case cube.FaceNorth:
		delta[2] = shulkerBBox.Min().Z() - entityBBox.Max().Z()
	}
	return delta
}

func (s ShulkerBox) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, pickaxeEffective, oneOf(s))
}

func (s ShulkerBox) MaxCount() int {
	return 1
}

func (s ShulkerBox) EncodeBlock() (name string, properties map[string]any) {
	if c, ok := s.Colour.Colour(); ok {
		return "minecraft:" + c.String() + "_shulker_box", nil
	}
	return "minecraft:undyed_shulker_box", nil
}

func (s ShulkerBox) EncodeItem() (id string, meta int16) {
	name, _ := s.EncodeBlock()
	return name, 0
}

func (s ShulkerBox) DecodeNBT(data map[string]any) any {
	s = s.initialised()
	nbtconv.InvFromNBT(s.inventory, nbtconv.Slice(data, "Items"))
	s.Facing = cube.Face(nbtconv.Uint8(data, "facing"))
	s.CustomName = nbtconv.String(data, "CustomName")
	return s
}

func (s ShulkerBox) EncodeNBT() map[string]any {
	s = s.initialised()
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

// allShulkerBoxes returns one shulker box per item.OptionalColour, including the undyed variant.
func allShulkerBoxes() (boxes []world.Block) {
	for _, c := range item.OptionalColours() {
		boxes = append(boxes, ShulkerBox{Colour: c})
	}
	return boxes
}
