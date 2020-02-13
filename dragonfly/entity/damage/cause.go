package damage

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"

// Source represents the source of the damage dealt to an entity. This source may be passed to the Hurt()
// method of an entity in order to deal damage to an entity with a specific source.
type Source interface {
	__()
}

// SourceEntityAttack is used for damage caused by other entities, for example when a player attacks another
// player.
type SourceEntityAttack struct {
	// Attacker holds the attacking entity. The entity may be a player or any other entity.
	Attacker world.Entity

	source
}

type source struct{}

func (source) __() {}
