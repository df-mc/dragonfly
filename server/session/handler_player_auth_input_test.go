package session

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestShieldBlockingInput(t *testing.T) {
	tests := []struct {
		name        string
		flags       []int
		wasSneaking bool
		sneaking    bool
		wantDown    bool
		wantUpdated bool
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
}
