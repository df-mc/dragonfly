package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type heldItemStateTestEntity struct {
	h       *world.EntityHandle
	updates *int
}

func (e heldItemStateTestEntity) Close() error           { return nil }
func (e heldItemStateTestEntity) H() *world.EntityHandle { return e.h }
func (e heldItemStateTestEntity) Position() mgl64.Vec3   { return mgl64.Vec3{} }
func (e heldItemStateTestEntity) Rotation() cube.Rotation {
	return cube.Rotation{}
}
func (e heldItemStateTestEntity) UpdateHeldItemState() {
	*e.updates++
}

type heldItemStateTestConfig struct {
	updates *int
}

func (c heldItemStateTestConfig) Apply(data *world.EntityData) {
	data.Data = c.updates
}

type heldItemStateTestType struct{}

func (heldItemStateTestType) Open(_ *world.Tx, h *world.EntityHandle, data *world.EntityData) world.Entity {
	return heldItemStateTestEntity{h: h, updates: data.Data.(*int)}
}
func (heldItemStateTestType) EncodeEntity() string { return "dragonfly:held_item_state_test" }
func (heldItemStateTestType) BBox(world.Entity) cube.BBox {
	return cube.Box(0, 0, 0, 0, 0, 0)
}
func (heldItemStateTestType) DecodeNBT(map[string]any, *world.EntityData) {}
func (heldItemStateTestType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

func TestItemStackRequestHeldSlotMutationUpdatesHeldItemState(t *testing.T) {
	var updates int
	handle := world.EntitySpawnOpts{}.New(heldItemStateTestType{}, heldItemStateTestConfig{updates: &updates})
	heldSlot := uint32(2)
	s := &Session{
		ent:      handle,
		heldSlot: &heldSlot,
		inv:      inventory.New(36, nil),
		offHand:  inventory.New(1, nil),
	}
	h := &ItemStackRequestHandler{
		changes:         map[byte]map[byte]changeInfo{},
		responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{},
	}
	w := world.Config{Entities: world.EntityRegistryConfig{}.New([]world.EntityType{heldItemStateTestType{}})}.New()
	defer func() {
		_ = w.Close()
	}()

	<-w.Exec(func(tx *world.Tx) {
		tx.AddEntity(handle)
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			Container: protocol.FullContainerName{ContainerID: protocol.ContainerInventory},
			Slot:      byte(heldSlot),
		}, item.NewStack(item.Shield{}, 1), s, tx)
	})
	if updates != 1 {
		t.Fatalf("expected held slot mutation to update held item state once, got %v", updates)
	}
}

func TestItemStackRequestOffHandMutationUpdatesHeldItemState(t *testing.T) {
	var updates int
	handle := world.EntitySpawnOpts{}.New(heldItemStateTestType{}, heldItemStateTestConfig{updates: &updates})
	heldSlot := uint32(2)
	s := &Session{
		ent:      handle,
		heldSlot: &heldSlot,
		inv:      inventory.New(36, nil),
		offHand:  inventory.New(1, nil),
	}
	h := &ItemStackRequestHandler{
		changes:         map[byte]map[byte]changeInfo{},
		responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{},
	}
	w := world.Config{Entities: world.EntityRegistryConfig{}.New([]world.EntityType{heldItemStateTestType{}})}.New()
	defer func() {
		_ = w.Close()
	}()

	<-w.Exec(func(tx *world.Tx) {
		tx.AddEntity(handle)
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			Container: protocol.FullContainerName{ContainerID: protocol.ContainerOffhand},
			Slot:      0,
		}, item.NewStack(item.Shield{}, 1), s, tx)
	})
	if updates != 1 {
		t.Fatalf("expected off-hand mutation to update held item state once, got %v", updates)
	}
}
