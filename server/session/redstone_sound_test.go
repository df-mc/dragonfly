package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// TestRedstoneSoundMaterial checks that click packets carry their block's material.
func TestRedstoneSoundMaterial(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	button := block.Button{Type: block.OakButton(), Pressed: true}
	plate := block.PressurePlate{Type: block.HeavyWeightedPressurePlate(), Power: 1}
	for _, tt := range []struct {
		s     world.Sound
		b     world.Block
		event string
	}{
		{sound.ButtonClickOn{Block: button}, button, packet.SoundEventButtonClickOn},
		{sound.ButtonClickOff{Block: button}, button, packet.SoundEventButtonClickOff},
		{sound.PressurePlateClickOn{Block: plate}, plate, packet.SoundEventPressurePlateClickOn},
		{sound.PressurePlateClickOff{Block: plate}, plate, packet.SoundEventPressurePlateClickOff},
	} {
		s := &Session{br: world.DefaultBlockRegistry, packets: make(chan packet.Packet, 1)}
		s.playSound(mgl64.Vec3{}, tt.s, false)
		pk := (<-s.packets).(*packet.LevelSoundEvent)
		if pk.SoundType != tt.event || pk.ExtraData != int32(s.br.BlockRuntimeID(tt.b)) {
			t.Errorf("%T: sound=%s data=%d, want sound=%s data=%d", tt.s, pk.SoundType, pk.ExtraData, tt.event, s.br.BlockRuntimeID(tt.b))
		}
	}
}
