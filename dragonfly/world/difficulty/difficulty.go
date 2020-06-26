package difficulty

// Difficulty represents the difficulty of a Minecraft world. The difficulty of a world influences all kinds
// of aspects of the world, such as the damage enemies deal to players, the way hunger depletes, whether
// hostile monsters spawn or not and more.
type Difficulty interface {
	__()
}

// Peaceful difficulty prevents most hostile mobs from spawning and makes players rapidly regenerate health
// and food.
type Peaceful struct{}

// Easy difficulty has mobs deal less damage to players than normal and starvation won't occur if a player
// has less than 5 hearts of health.
type Easy struct{}

// Normal difficulty has mobs that deal normal damage to players. Starvation will occur until the player is
// down to a single heart.
type Normal struct{}

// Hard difficulty has mobs that deal above average damage to players. Starvation will kill players with too
// little food and monsters will get additional effects.
type Hard struct{}

func (Peaceful) __() {}
func (Easy) __()     {}
func (Normal) __()   {}
func (Hard) __()     {}
