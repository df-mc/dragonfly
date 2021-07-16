package world

// Difficulty represents the difficulty of a Minecraft world. The difficulty of a world influences all kinds
// of aspects of the world, such as the damage enemies deal to players, the way hunger depletes, whether
// hostile monsters spawn or not and more.
type Difficulty interface {
	// FoodRegenerates specifies if players' food levels should automatically regenerate with this difficulty.
	FoodRegenerates() bool
	// StarvationHealthLimit specifies the amount of health at which a player will no longer receive damage from
	// starvation.
	StarvationHealthLimit() float64
	// FireSpreadIncrease returns a number that increases the rate at which fire spreads.
	FireSpreadIncrease() int
}

// DifficultyPeaceful difficulty prevents most hostile mobs from spawning and makes players rapidly regenerate
// health and food.
type DifficultyPeaceful struct{}

// FoodRegenerates ...
func (DifficultyPeaceful) FoodRegenerates() bool {
	return true
}

// StarvationHealthLimit ...
func (DifficultyPeaceful) StarvationHealthLimit() float64 {
	return 20
}

// FireSpreadIncrease ...
func (DifficultyPeaceful) FireSpreadIncrease() int {
	return 0
}

// DifficultyEasy difficulty has mobs deal less damage to players than normal and starvation won't occur if
// a player has less than 5 hearts of health.
type DifficultyEasy struct{}

// FoodRegenerates ...
func (DifficultyEasy) FoodRegenerates() bool {
	return false
}

// StarvationHealthLimit ...
func (DifficultyEasy) StarvationHealthLimit() float64 {
	return 10
}

// FireSpreadIncrease ...
func (DifficultyEasy) FireSpreadIncrease() int {
	return 7
}

// DifficultyNormal difficulty has mobs that deal normal damage to players. Starvation will occur until the
// player is down to a single heart.
type DifficultyNormal struct{}

// FoodRegenerates ...
func (DifficultyNormal) FoodRegenerates() bool {
	return false
}

// StarvationHealthLimit ...
func (DifficultyNormal) StarvationHealthLimit() float64 {
	return 2
}

// FireSpreadIncrease ...
func (DifficultyNormal) FireSpreadIncrease() int {
	return 14
}

// DifficultyHard difficulty has mobs that deal above average damage to players. Starvation will kill players
// with too little food and monsters will get additional effects.
type DifficultyHard struct{}

// FoodRegenerates ...
func (DifficultyHard) FoodRegenerates() bool {
	return false
}

// StarvationHealthLimit ...
func (DifficultyHard) StarvationHealthLimit() float64 {
	return -1
}

// FireSpreadIncrease ...
func (DifficultyHard) FireSpreadIncrease() int {
	return 21
}
