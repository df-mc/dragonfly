package healing

// SourceFood is a healing source used for when the player regenerates health automatically when their food
// bar is at least 90% filled.
type SourceFood struct{}

// SourceCustom is a healing source that may be used by users to represent a custom healing source.
type SourceCustom struct{}

// Source represents a source of healing for an entity. This source may be passed to the Heal() method of a
// living entity.
type Source interface {
	__()
}

func (SourceFood) __()   {}
func (SourceCustom) __() {}
