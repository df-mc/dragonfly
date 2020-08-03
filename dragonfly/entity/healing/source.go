package healing

// SourceFood is a healing source used for when an entity regenerates health automatically when their food
// bar is at least 90% filled.
type SourceFood struct{}

// SourceRegenerationEffect is a healing source used when an entity regenerates due to the Regeneration
// effect.
type SourceRegenerationEffect struct{}

// SourceInstantHealthEffect is a healing source used when an entity regenerations due to an Instant Health
// effect.
type SourceInstantHealthEffect struct{}

// SourceCustom is a healing source that may be used by users to represent a custom healing source.
type SourceCustom struct{}

// Source represents a source of healing for an entity. This source may be passed to the Heal() method of a
// living entity.
type Source interface {
	__()
}

func (SourceFood) __()                {}
func (SourceCustom) __()              {}
func (SourceRegenerationEffect) __()  {}
func (SourceInstantHealthEffect) __() {}
