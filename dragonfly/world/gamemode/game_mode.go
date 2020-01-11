package gamemode

// GameMode represents a game mode that may be assigned to a player. Upon joining the world, players will be
// given the default game mode that the world holds.
// Game modes specify the way that a player interacts with and plays in the world.
type GameMode interface {
	__()
}

// Survival represents the survival game mode: Players with this game mode have limited supplies and can break
// blocks using only the right tools.
type Survival struct{}

// Creative represents the creative game mode: Players with this game mode have infinite blocks and items and
// can break blocks instantly. Players with creative mode can also fly.
type Creative struct{}

// Adventure represents the adventure game mode: Players with this game mode cannot edit the world (placing or
// breaking blocks).
type Adventure struct{}

// Spectator represents the spectator game mode: Players with this game mode cannot interact with the world
// and cannot be seen by other players. Spectator players can fly, like creative mode, and can move through
// blocks.
type Spectator struct{}

func (Survival) __()  {}
func (Creative) __()  {}
func (Adventure) __() {}
func (Spectator) __() {}
