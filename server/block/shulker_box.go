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
)

type ShulkerBox struct {
	solid      // TODO: I don't think it should be solid
	Type       ShulkerBoxType
	Facing     cube.Face
	CustomName string

	inventory *inventory.Inventory
	viewerMu  *sync.RWMutex
	viewers   map[ContainerViewer]struct{}
}

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

func (s ShulkerBox) Inventory(tx *world.Tx, pos cube.Pos) *inventory.Inventory {
	return s.inventory
}

func (s ShulkerBox) Activate(pos cube.Pos, clickedFace cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		if d, ok := tx.Block(pos.Side(s.Facing)).(LightDiffuser); ok && d.LightDiffusionLevel() <= 2 {
			opener.OpenBlockContainer(pos, tx)
		}
		return true
	}

	return false
}

func (s ShulkerBox) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}
	//noinspection GoAssignmentToReceiver
	s = NewShulkerBox()
	s.Facing = face
	place(tx, pos, s, user, ctx)
	return placed(ctx)
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
	tx.PlaySound(pos.Vec3Centre(), sound.ShulkerBoxClose{}) //TODO: Make the sound delayed to sync with the closing action
}

func (s ShulkerBox) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(10)
}

func (s ShulkerBox) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + s.Type.String(), nil
}

func (s ShulkerBox) EncodeItem() (id string, meta int16) {
	return "minecraft:" + s.Type.String(), 0
}

func (s ShulkerBox) DecodeNBT(data map[string]any) any {
	//noinspection GoAssignmentToReceiver
	s = NewShulkerBox()
	nbtconv.InvFromNBT(s.inventory, nbtconv.Slice(data, "Items"))
	s.Facing = cube.Face(nbtconv.Uint8(data, "facing"))
	s.CustomName = nbtconv.String(data, "CustomName")
	return s
}

func (s ShulkerBox) EncodeNBT() map[string]any {
	if s.inventory == nil {
		facing, customName := s.Facing, s.CustomName
		//noinspection GoAssignmentToReceiver
		s = NewShulkerBox()
		s.Facing, s.CustomName = facing, customName
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

func allShulkerBox() (shulkerboxes []world.Block) {
	for _, t := range ShulkerBoxTypes() {
		shulkerboxes = append(shulkerboxes, ShulkerBox{Type: t})
	}

	return
}
