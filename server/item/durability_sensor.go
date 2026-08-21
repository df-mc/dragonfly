package item

// DurabilitySensor represents an item that emits effects when it receives damage. The item also
// needs a minecraft:durability component.
type DurabilitySensor interface {
	// DurabilitySensorInfo returns the durability sensor information of the item.
	DurabilitySensorInfo() DurabilitySensorInfo
}

// DurabilitySensorInfo is a struct returned by items that implement DurabilitySensor. It contains
// the thresholds and effects emitted when durability is reduced.
type DurabilitySensorInfo struct {
	// SoundEvent is the sound effect emitted when any threshold is met.
	SoundEvent string
	// DurabilityThresholds is a list of thresholds at which effects are emitted.
	DurabilityThresholds []DurabilityThreshold
}

// DurabilityThreshold defines the durability threshold and effects emitted when that threshold
// is met.
type DurabilityThreshold struct {
	// Durability is the durability value at which effects are emitted. Effects are emitted when
	// the item durability value is less than or equal to this value.
	Durability int
	// ParticleType is the particle effect to emit when the threshold is met.
	ParticleType string
	// SoundEvent is the sound effect to emit when the threshold is met.
	SoundEvent string
}
