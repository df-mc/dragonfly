package item

// GhastTear is a brewing item dropped by ghasts.
type GhastTear struct{}

func (GhastTear) EncodeItem() (name string, meta int16) {
	return "minecraft:ghast_tear", 0
}
