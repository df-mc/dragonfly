package item

// RecoveryCompass is an item used to point to the location of the player's last death.
type RecoveryCompass struct{}

// EncodeItem ...
func (RecoveryCompass) EncodeItem() (name string, meta int16) {
	return "minecraft:recovery_compass", 0
}
