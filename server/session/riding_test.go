package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestRidingLinkUsesRoleAndRemovalImmediacy(t *testing.T) {
	rider, rideable := &world.EntityHandle{}, &world.EntityHandle{}
	s := &Session{
		entityRuntimeIDs: map[*world.EntityHandle]uint64{rider: 12, rideable: 34},
		packets:          make(chan packet.Packet, 4),
		closeBackground:  make(chan struct{}),
	}

	s.viewEntityMountHandles(rider, rideable, true)
	s.viewEntityMountHandles(rider, rideable, false)
	s.viewEntityDismountHandles(rider, rideable, false)
	s.viewEntityDismountHandles(rider, rideable, true)

	want := []struct {
		linkType  byte
		immediate bool
	}{
		{linkType: protocol.EntityLinkRider},
		{linkType: protocol.EntityLinkPassenger},
		{linkType: protocol.EntityLinkRemove},
		{linkType: protocol.EntityLinkRemove, immediate: true},
	}
	for i, expected := range want {
		pk := (<-s.packets).(*packet.SetActorLink)
		if pk.EntityLink.Type != expected.linkType || pk.EntityLink.Immediate != expected.immediate {
			t.Fatalf("link %d: got type=%d immediate=%t, want type=%d immediate=%t", i, pk.EntityLink.Type, pk.EntityLink.Immediate, expected.linkType, expected.immediate)
		}
		if pk.EntityLink.RiderEntityUniqueID != 12 || pk.EntityLink.RiddenEntityUniqueID != 34 {
			t.Fatalf("link %d has rider=%d ridden=%d", i, pk.EntityLink.RiderEntityUniqueID, pk.EntityLink.RiddenEntityUniqueID)
		}
	}
}
