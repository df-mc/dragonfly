package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type metadataBlockingEntity struct{}

func (metadataBlockingEntity) Blocking() bool { return true }

type metadataShieldBlockingEntity struct{}

func (metadataShieldBlockingEntity) ShieldBlocking() bool { return true }

type heldItemStateTestEntity struct {
	h       *world.EntityHandle
	updates *int
}

func (e heldItemStateTestEntity) Close() error           { return nil }
func (e heldItemStateTestEntity) H() *world.EntityHandle { return e.h }
func (heldItemStateTestEntity) Position() mgl64.Vec3     { return mgl64.Vec3{} }
func (heldItemStateTestEntity) Rotation() cube.Rotation  { return cube.Rotation{} }
func (e heldItemStateTestEntity) UpdateHeldItemState()   { *e.updates++ }

type heldItemStateTestConfig struct{ updates *int }

func (c heldItemStateTestConfig) Apply(data *world.EntityData) { data.Data = c.updates }

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

type heldItemDropper struct{}

func (heldItemDropper) Drop(stack item.Stack) int { return stack.Count() }

func TestShieldBlockingInput(t *testing.T) {
	tests := []struct {
		name                  string
		flags                 []int
		wasSneaking, sneaking bool
		wantDown, wantUpdated bool
	}{
		{name: "held raw wins over stop", flags: []int{packet.InputFlagSneakCurrentRaw, packet.InputFlagStopSneaking}, wasSneaking: true, wantDown: true, wantUpdated: true},
		{name: "release stops", flags: []int{packet.InputFlagSneakReleasedRaw}, wasSneaking: true, wantUpdated: true},
		{name: "item use is unrelated", flags: []int{packet.InputFlagStartUsingItem}},
		{name: "accepted start begins", flags: []int{packet.InputFlagStartSneaking}, sneaking: true, wantDown: true, wantUpdated: true},
		{name: "cancelled start ignored", flags: []int{packet.InputFlagStartSneaking, packet.InputFlagSneakCurrentRaw}},
		{name: "held raw after cancellation ignored", flags: []int{packet.InputFlagSneakCurrentRaw}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
			for _, flag := range tt.flags {
				flags.Set(flag)
			}
			down, updated := shieldBlockingInput(flags, tt.wasSneaking, tt.sneaking)
			if down != tt.wantDown || updated != tt.wantUpdated {
				t.Fatalf("shieldBlockingInput() = (%v, %v), want (%v, %v)", down, updated, tt.wantDown, tt.wantUpdated)
			}
		})
	}

	t.Run("inventory transaction drop", func(t *testing.T) {
		var updates int
		handle := world.EntitySpawnOpts{}.New(heldItemStateTestType{}, heldItemStateTestConfig{updates: &updates})
		heldSlot := uint32(2)
		s := &Session{ent: handle, heldSlot: &heldSlot, inv: inventory.New(36, nil)}
		held := item.NewStack(item.Apple{}, 2)
		_ = s.inv.SetItem(int(heldSlot), held)
		w := world.Config{Synchronous: true, Entities: world.EntityRegistryConfig{}.New([]world.EntityType{heldItemStateTestType{}})}.New()
		defer w.Close()
		var err error
		w.Do(func(tx *world.Tx) {
			tx.AddEntity(handle)
			err = (&InventoryTransactionHandler{}).handleNormalTransaction(&packet.InventoryTransaction{Actions: []protocol.InventoryAction{
				{SourceType: protocol.InventoryActionSourceWorld, NewItem: instanceFromItem(s.br, item.NewStack(item.Apple{}, 1))},
				{SourceType: protocol.InventoryActionSourceContainer, WindowID: protocol.WindowIDInventory, InventorySlot: heldSlot, OldItem: instanceFromItem(s.br, held)},
			}}, s, tx, heldItemDropper{})
		})
		if err != nil || updates != 1 {
			t.Fatalf("drop transaction error=%v, held item updates=%v; want nil, 1", err, updates)
		}
	})
}

func TestShieldBlockingMetadata(t *testing.T) {
	tests := []struct {
		name string
		ent  any
		want bool
	}{
		{name: "shield blocking", ent: metadataShieldBlockingEntity{}, want: true},
		{name: "unrelated blocking", ent: metadataBlockingEntity{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := protocol.NewEntityMetadata()
			new(Session).addSpecificMetadata(tt.ent, metadata)
			if got := metadata.Flag(protocol.EntityDataKeyFlagsTwo, protocol.EntityDataFlagBlocking&63); got != tt.want {
				t.Fatalf("blocking metadata = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShieldInventoryChangesUpdateHeldItemState(t *testing.T) {
	tests := []struct {
		name string
		run  func(*ItemStackRequestHandler, *Session, *world.Tx)
	}{
		{name: "held slot mutation", run: func(h *ItemStackRequestHandler, s *Session, tx *world.Tx) {
			h.setItemInSlot(protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: protocol.ContainerInventory}, Slot: byte(*s.heldSlot)}, item.NewStack(item.Shield{}, 1), s, tx)
		}},
		{name: "off-hand rollback", run: func(h *ItemStackRequestHandler, s *Session, tx *world.Tx) {
			before := item.NewStack(item.Shield{}, 1)
			_ = s.offHand.SetItem(0, before)
			h.changes = map[byte]map[byte]changeInfo{protocol.ContainerOffhand: {0: {before: before}}}
			h.reject(1, s, tx)
			if got, _ := s.offHand.Item(0); !got.Equal(before) {
				t.Fatalf("off-hand rollback restored %v, want %v", got, before)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updates int
			handle := world.EntitySpawnOpts{}.New(heldItemStateTestType{}, heldItemStateTestConfig{updates: &updates})
			heldSlot := uint32(2)
			s := &Session{ent: handle, heldSlot: &heldSlot, inv: inventory.New(36, nil), offHand: inventory.New(1, nil), packets: make(chan packet.Packet, 1)}
			h := &ItemStackRequestHandler{changes: map[byte]map[byte]changeInfo{}, responseChanges: map[int32]map[*inventory.Inventory]map[byte]responseChange{}}
			w := world.Config{Synchronous: true, Entities: world.EntityRegistryConfig{}.New([]world.EntityType{heldItemStateTestType{}})}.New()
			defer w.Close()
			w.Do(func(tx *world.Tx) {
				tx.AddEntity(handle)
				tt.run(h, s, tx)
			})
			if updates != 1 {
				t.Fatalf("held item state updated %v times, want 1", updates)
			}
		})
	}
}
