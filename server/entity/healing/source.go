package healing

type (
	// Source represents a source of healing for an entity. This source may be passed to the Living.Heal() method of a
	// living entity.
	Source interface {
		HealingSource()
	}

	// SourceFood is a healing source used for when an entity regenerates health automatically when their food
	// bar is at least 90% filled.
	SourceFood struct{}

	// SourceRegenerationEffect is a healing source used when an entity regenerates due to the Regeneration
	// effect.
	SourceRegenerationEffect struct{}

	// SourceInstantHealthEffect is a healing source used when an entity regenerations due to an Instant Health
	// effect.
	SourceInstantHealthEffect struct{}
)

func (SourceFood) HealingSource()                {}
func (SourceRegenerationEffect) HealingSource()  {}
func (SourceInstantHealthEffect) HealingSource() {}
