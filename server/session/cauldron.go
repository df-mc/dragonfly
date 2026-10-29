package session

import (
	"image/color"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// cauldronSoundEvent maps cauldron interactions to the level events that play
// their sound and particle effects on the client.
func cauldronSoundEvent(s world.Sound) (event int32, colour color.RGBA, ok bool) {
	switch s := s.(type) {
	case sound.CauldronFillWater:
		return packet.LevelEventCauldronFillWater, s.Colour, true
	case sound.CauldronTakeWater:
		return packet.LevelEventCauldronTakeWater, s.Colour, true
	case sound.CauldronFillPotion:
		return packet.LevelEventCauldronFillPotion, s.Colour, true
	case sound.CauldronTakePotion:
		return packet.LevelEventCauldronTakePotion, s.Colour, true
	case sound.CauldronFillLava:
		return packet.LevelEventCauldronFillLava, color.RGBA{}, true
	case sound.CauldronTakeLava:
		return packet.LevelEventCauldronTakeLava, color.RGBA{}, true
	case sound.CauldronFillPowderSnow:
		return packet.LevelEventCauldronFillPowderSnow, color.RGBA{}, true
	case sound.CauldronTakePowderSnow:
		return packet.LevelEventCauldronTakePowderSnow, color.RGBA{}, true
	case sound.CauldronAddDye:
		return packet.LevelEventCauldronAddDye, s.Colour, true
	case sound.CauldronDyeArmour:
		return packet.LevelEventCauldronDyeArmor, s.Colour, true
	case sound.CauldronCleanArmour:
		return packet.LevelEventCauldronCleanArmor, s.Colour, true
	case sound.CauldronCleanBanner:
		return packet.LevelEventCauldronCleanBanner, s.Colour, true
	case sound.CauldronExplode:
		return packet.LevelEventCauldronExplode, s.Colour, true
	}
	return 0, color.RGBA{}, false
}
