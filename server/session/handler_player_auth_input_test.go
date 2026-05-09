package session

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestShieldBlockingInputPrefersHeldRawInputOverStopSneaking(t *testing.T) {
	flags := protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
	flags.Set(packet.InputFlagSneakCurrentRaw)
	flags.Set(packet.InputFlagStopSneaking)

	down, ok := shieldBlockingInput(flags, true, false)
	if !ok {
		t.Fatal("expected shield input to be updated")
	}
	if !down {
		t.Fatal("expected held raw sneak input to keep shield input active")
	}
}

func TestShieldBlockingInputStopsOnSneakRelease(t *testing.T) {
	flags := protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
	flags.Set(packet.InputFlagSneakReleasedRaw)

	down, ok := shieldBlockingInput(flags, true, false)
	if !ok {
		t.Fatal("expected shield input to be updated")
	}
	if down {
		t.Fatal("expected released raw sneak input to stop shield input")
	}
}

func TestShieldBlockingInputIgnoresUseItem(t *testing.T) {
	flags := protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
	flags.Set(packet.InputFlagStartUsingItem)

	if down, ok := shieldBlockingInput(flags, false, false); ok || down {
		t.Fatal("expected generic shield input helper not to start from item use")
	}
}

func TestShieldBlockingInputIgnoresCancelledStartSneaking(t *testing.T) {
	flags := protocol.NewBitset(packet.PlayerAuthInputBitsetSize)
	flags.Set(packet.InputFlagStartSneaking)
	flags.Set(packet.InputFlagSneakCurrentRaw)

	if down, ok := shieldBlockingInput(flags, false, false); ok || down {
		t.Fatal("expected cancelled start sneaking not to start shield input")
	}
}
