package block

import (
	"fmt"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// ShulkerBox is a dyeable container block that keeps its contents when broken.
// The empty value is valid as a registry entry, but placed/decoded shulker boxes
// should be initialised through NewShulkerBox().
type ShulkerBox struct {
	solid
	transparent
	sourceWaterDisplacer

	// Dyed reports if the shulker box has a colour.
	Dyed bool
	// Colour is the colour of the shulker box when Dyed is true.
	Colour item.Colour
	// Facing is the face the shulker opens towards.
	Facing cube.Face
	// CustomName is the displayed container name.
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

// baseShulkerBoxItem is an item-only alias for the legacy undyed shulker item
// runtime that uses minecraft:shulker_box instead of minecraft:undyed_shulker_box.
type baseShulkerBoxItem struct{ ShulkerBox }

// NewShulkerBox creates a new initialised shulker box.
func NewShulkerBox() ShulkerBox {
	s := ShulkerBox{
		viewerMu: new(sync.RWMutex),
		viewers:  make(map[ContainerViewer]struct{}, 1),
	}
	s.inventory = inventory.New(27, func(slot int, _, after item.Stack) {
		s.viewerMu.RLock()
		defer s.viewerMu.RUnlock()

		for viewer := range s.viewers {
			if isShulkerBoxItem(after.Item()) {
				// Bedrock blocks shulker-in-shulker storage client-side. Avoid sending
				// impossible contents back to viewers here too.
				continue
			}
			viewer.ViewSlotChange(slot, after)
		}
	})
	return s
}

// WithName returns the shulker box with a custom name applied.
func (s ShulkerBox) WithName(a ...any) world.Item {
	s.CustomName = strings.TrimSuffix(fmt.Sprintln(a...), "\n")
	return s
}

// SideClosed ...
func (ShulkerBox) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// Inventory returns the shulker box inventory.
func (s ShulkerBox) Inventory(*world.Tx, cube.Pos) *inventory.Inventory {
	return s.inventory
}

// AddViewer adds a viewer to the shulker box inventory.
func (s ShulkerBox) AddViewer(v ContainerViewer, tx *world.Tx, pos cube.Pos) {
	s.viewerMu.Lock()
	defer s.viewerMu.Unlock()

	if len(s.viewers) == 0 {
		s.open(tx, pos)
	}
	s.viewers[v] = struct{}{}
}

// RemoveViewer removes a viewer from the shulker box inventory.
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

// Activate ...
func (s ShulkerBox) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		if d, ok := tx.Block(pos.Side(s.Facing)).(LightDiffuser); ok && d.LightDiffusionLevel() <= 2 {
			opener.OpenBlockContainer(pos, tx)
		}
		return true
	}
	return false
}

// UseOnBlock ...
func (s ShulkerBox) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}
	if s.inventory == nil {
		dyed, colour, customName := s.Dyed, s.Colour, s.CustomName
		s = NewShulkerBox()
		s.Dyed, s.Colour, s.CustomName = dyed, colour, customName
	}
	s.Facing = face

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s ShulkerBox) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, pickaxeEffective, oneOf(s))
}

// MaxCount always returns 1.
func (ShulkerBox) MaxCount() int {
	return 1
}

func (s ShulkerBox) open(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, OpenAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxOpen{})
}

func (s ShulkerBox) close(tx *world.Tx, pos cube.Pos) {
	for _, v := range tx.Viewers(pos.Vec3()) {
		v.ViewBlockAction(pos, CloseAction{})
	}
	tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxClose{})
}

// EncodeBlock ...
func (s ShulkerBox) EncodeBlock() (string, map[string]any) {
	name := "minecraft:"
	if s.Dyed {
		name += s.Colour.String() + "_"
	} else {
		name += "undyed_"
	}
	name += "shulker_box"
	return name, nil
}

// EncodeItem ...
func (s ShulkerBox) EncodeItem() (string, int16) {
	name, _ := s.EncodeBlock()
	return name, 0
}

// DecodeNBT ...
func (s ShulkerBox) DecodeNBT(data map[string]any) any {
	dyed, colour := s.Dyed, s.Colour
	s = NewShulkerBox()
	s.Dyed, s.Colour = dyed, colour

	nbtconv.InvFromNBT(s.inventory, shulkerNBTItems(data))
	s.Facing = cube.Face(nbtconv.Uint8(data, "facing"))
	s.CustomName = nbtconv.String(data, "CustomName")
	return s
}

// EncodeNBT ...
func (s ShulkerBox) EncodeNBT() map[string]any {
	if s.inventory == nil {
		dyed, colour, facing, customName := s.Dyed, s.Colour, s.Facing, s.CustomName
		s = NewShulkerBox()
		s.Dyed, s.Colour, s.Facing, s.CustomName = dyed, colour, facing, customName
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

func allShulkerBoxes() (boxes []world.Block) {
	boxes = append(boxes, ShulkerBox{})
	for _, colour := range item.Colours() {
		boxes = append(boxes, ShulkerBox{Dyed: true, Colour: colour})
	}
	return boxes
}

func (baseShulkerBoxItem) EncodeItem() (string, int16) {
	return "minecraft:shulker_box", 0
}

func (s baseShulkerBoxItem) DecodeNBT(data map[string]any) any {
	return baseShulkerBoxItem{ShulkerBox: s.ShulkerBox.DecodeNBT(data).(ShulkerBox)}
}

func (s baseShulkerBoxItem) WithName(a ...any) world.Item {
	s.ShulkerBox = s.ShulkerBox.WithName(a...).(ShulkerBox)
	return s
}

func shulkerNBTItems(data map[string]any) []any {
	if items := nbtconv.Slice(data, "Items"); items != nil {
		return items
	}
	if items, ok := data["Items"].([]map[string]any); ok {
		out := make([]any, 0, len(items))
		for _, entry := range items {
			out = append(out, entry)
		}
		return out
	}
	return nil
}

func isShulkerBoxItem(it world.Item) bool {
	switch it.(type) {
	case ShulkerBox, baseShulkerBoxItem:
		return true
	default:
		return false
	}
}
