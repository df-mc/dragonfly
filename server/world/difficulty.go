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

var (
	// DifficultyPeaceful prevents most hostile mobs from spawning and makes players rapidly regenerate health and food.
	DifficultyPeaceful difficultyPeaceful
	// DifficultyEasy has mobs that deal less damage to players than normal and starvation won't occur if a player has
	// less than 5 hearts of health.
	DifficultyEasy difficultyEasy
	// DifficultyNormal has mobs that deal normal damage to players. Starvation will occur until the player is down to
	// a single heart.
	DifficultyNormal difficultyNormal
	// DifficultyHard has mobs that deal above average damage to players. Starvation will kill players with too little
	// food and monsters will get additional effects.
	DifficultyHard difficultyHard
)

// difficultyPeaceful difficulty prevents most hostile mobs from spawning and makes players rapidly regenerate
// health and food.
type difficultyPeaceful struct{}

func (difficultyPeaceful) FoodRegenerates() bool          { return true }
func (difficultyPeaceful) StarvationHealthLimit() float64 { return 20 }
func (difficultyPeaceful) FireSpreadIncrease() int        { return 0 }

// difficultyEasy difficulty has mobs deal less damage to players than normal and starvation won't occur if
// a player has less than 5 hearts of health.
type difficultyEasy struct{}

func (difficultyEasy) FoodRegenerates() bool          { return false }
func (difficultyEasy) StarvationHealthLimit() float64 { return 10 }
func (difficultyEasy) FireSpreadIncrease() int        { return 7 }

// difficultyNormal difficulty has mobs that deal normal damage to players. Starvation will occur until the
// player is down to a single heart.
type difficultyNormal struct{}

func (difficultyNormal) FoodRegenerates() bool          { return false }
func (difficultyNormal) StarvationHealthLimit() float64 { return 2 }
func (difficultyNormal) FireSpreadIncrease() int        { return 14 }

// difficultyHard difficulty has mobs that deal above average damage to players. Starvation will kill players
// with too little food and monsters will get additional effects.
type difficultyHard struct{}

func (difficultyHard) FoodRegenerates() bool          { return false }
func (difficultyHard) StarvationHealthLimit() float64 { return -1 }
func (difficultyHard) FireSpreadIncrease() int        { return 21 }
