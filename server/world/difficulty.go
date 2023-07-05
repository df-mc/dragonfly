package world

// Difficulty represents the difficulty of a Minecraft world. The difficulty of
// a world influences all kinds of aspects of the world, such as the damage
// enemies deal to players, the way hunger depletes, whether hostile monsters
// spawn or not and more.
type Difficulty interface {
	// FoodRegenerates specifies if players' food levels should automatically
	// regenerate with this difficulty.
	FoodRegenerates() bool
	// StarvationHealthLimit specifies the amount of health at which a player
	// will no longer receive damage from starvation.
	StarvationHealthLimit() float64
	// FireSpreadIncrease returns a number that increases the rate at which fire
	// spreads.
	FireSpreadIncrease() int
}

var (
	// DifficultyPeaceful prevents most hostile mobs from spawning and makes
	// players rapidly regenerate health and food.
	DifficultyPeaceful difficultyPeaceful
	// DifficultyEasy has mobs that deal less damage to players than normal and
	// starvation won't occur if a player has less than 5 hearts of health.
	DifficultyEasy difficultyEasy
	// DifficultyNormal has mobs that deal normal damage to players. Starvation
	// will occur until the player is down to a single heart.
	DifficultyNormal difficultyNormal
	// DifficultyHard has mobs that deal above average damage to players.
	// Starvation will kill players with too little food and monsters will get
	// additional effects.
	DifficultyHard difficultyHard
)

var difficultyReg = newDifficultyRegistry(map[int]Difficulty{
	0: DifficultyPeaceful,
	1: DifficultyEasy,
	2: DifficultyNormal,
	3: DifficultyHard,
})

// DifficultyByID looks up a Difficulty for the ID passed, returning
// DifficultyPeaceful for 0, DifficultyEasy for 1, DifficultyNormal for 2 and
// DifficultyHard for 3. If the ID is unknown, the bool returned is false. In
// this case the Difficulty returned is DifficultyNormal.
func DifficultyByID(id int) (Difficulty, bool) {
	return difficultyReg.Lookup(id)
}

// DifficultyID looks up the ID that a Difficulty was registered with. If not
// found, false is returned.
func DifficultyID(diff Difficulty) (int, bool) {
	return difficultyReg.LookupID(diff)
}

type difficultyRegistry struct {
	difficulties map[int]Difficulty
	ids          map[Difficulty]int
}

// newDifficultyRegistry returns an initialised difficultyRegistry.
func newDifficultyRegistry(diff map[int]Difficulty) *difficultyRegistry {
	ids := make(map[Difficulty]int, len(diff))
	for k, v := range diff {
		ids[v] = k
	}
	return &difficultyRegistry{difficulties: diff, ids: ids}
}

// Lookup looks up a Difficulty for the ID passed, returning DifficultyPeaceful
// for 0, DifficultyEasy for 1, DifficultyNormal for 2 and DifficultyHard for
// 3. If the ID is unknown, the bool returned is false. In this case the
// Difficulty returned is DifficultyNormal.
func (reg *difficultyRegistry) Lookup(id int) (Difficulty, bool) {
	dim, ok := reg.difficulties[id]
	if !ok {
		dim = DifficultyNormal
	}
	return dim, ok
}

// LookupID looks up the ID that a Difficulty was registered with. If not found,
// false is returned.
func (reg *difficultyRegistry) LookupID(diff Difficulty) (int, bool) {
	id, ok := reg.ids[diff]
	return id, ok
}

// difficultyPeaceful difficulty prevents most hostile mobs from spawning and
// makes players rapidly regenerate health and food.
type difficultyPeaceful struct{}

func (difficultyPeaceful) FoodRegenerates() bool          { return true }
func (difficultyPeaceful) StarvationHealthLimit() float64 { return 20 }
func (difficultyPeaceful) FireSpreadIncrease() int        { return 0 }

// difficultyEasy difficulty has mobs deal less damage to players than normal
// and starvation won't occur if a player has less than 5 hearts of health.
type difficultyEasy struct{}

func (difficultyEasy) FoodRegenerates() bool          { return false }
func (difficultyEasy) StarvationHealthLimit() float64 { return 10 }
func (difficultyEasy) FireSpreadIncrease() int        { return 7 }

// difficultyNormal difficulty has mobs that deal normal damage to players.
// Starvation will occur until the player is down to a single heart.
type difficultyNormal struct{}

func (difficultyNormal) FoodRegenerates() bool          { return false }
func (difficultyNormal) StarvationHealthLimit() float64 { return 2 }
func (difficultyNormal) FireSpreadIncrease() int        { return 14 }

// difficultyHard difficulty has mobs that deal above average damage to
// players. Starvation will kill players with too little food and monsters will
// get additional effects.
type difficultyHard struct{}

func (difficultyHard) FoodRegenerates() bool          { return false }
func (difficultyHard) StarvationHealthLimit() float64 { return -1 }
func (difficultyHard) FireSpreadIncrease() int        { return 21 }
