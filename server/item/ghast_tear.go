package item

// Ghast tears are brewing items dropped by ghasts.
type GhastTear struct{}

// EncodeItem ...
func (GhastTear) EncodeItem() (name string, meta int16) {
	return "minecraft:ghast_tear", 0
}
