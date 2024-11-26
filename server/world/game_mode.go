package world

// GameMode represents a game mode that may be assigned to a player. Upon joining the world, players will be
// given the default game mode that the world holds.
// Game modes specify the way that a player interacts with and plays in the world.
type GameMode interface {
	// AllowsEditing specifies if a player with this GameMode can edit the World it's in.
	AllowsEditing() bool
	// AllowsTakingDamage specifies if a player with this GameMode can take damage from other entities.
	AllowsTakingDamage() bool
	// CreativeInventory specifies if a player with this GameMode has access to the creative inventory.
	CreativeInventory() bool
	// HasCollision specifies if a player with this GameMode can collide with blocks or entities in the world.
	HasCollision() bool
	// AllowsFlying specifies if a player with this GameMode can fly freely.
	AllowsFlying() bool
	// AllowsInteraction specifies if a player with this GameMode can interact with the world through entities or if it
	// can use items in the world.
	AllowsInteraction() bool
	// Visible specifies if a player with this GameMode can be visible to other players. If false, the player will be
	// invisible under any circumstance.
	Visible() bool
}

var (
	// GameModeSurvival is the survival game mode: Players with this game mode have limited supplies and can break blocks
	// after taking some time.
	GameModeSurvival survival
	// GameModeCreative represents the creative game mode: Players with this game mode have infinite blocks and
	// items and can break blocks instantly. Players with creative mode can also fly.
	GameModeCreative creative
	// GameModeAdventure represents the adventure game mode: Players with this game mode cannot edit the world
	// (placing or breaking blocks).
	GameModeAdventure adventure
	// GameModeSpectator represents the spectator game mode: Players with this game mode cannot interact with the
	// world and cannot be seen by other players. spectator players can fly, like creative mode, and can
	// move through blocks.
	GameModeSpectator spectator
)

var gameModeReg = newGameModeRegistry(map[int]GameMode{
	0: GameModeSurvival,
	1: GameModeCreative,
	2: GameModeAdventure,
	3: GameModeSpectator,
})

// GameModeByID looks up a GameMode for the ID passed, returning
// GameModeSurvival for 0, GameModeCreative for 1, GameModeAdventure for 2 and
// GameModeSpectator for 3. If the ID is unknown, the bool returned is false. In
// this case the GameMode returned is GameModeSurvival.
func GameModeByID(id int) (GameMode, bool) {
	return gameModeReg.Lookup(id)
}

// GameModeID looks up the ID that a GameMode was registered with. If not
// found, false is returned.
func GameModeID(mode GameMode) (int, bool) {
	return gameModeReg.LookupID(mode)
}

type gameModeRegistry struct {
	gameModes map[int]GameMode
	ids       map[GameMode]int
}

// newGameModeRegistry returns an initialised gameModeRegistry.
func newGameModeRegistry(mode map[int]GameMode) *gameModeRegistry {
	ids := make(map[GameMode]int, len(mode))
	for k, v := range mode {
		ids[v] = k
	}
	return &gameModeRegistry{gameModes: mode, ids: ids}
}

// Lookup looks up a GameMode for the ID passed, returning GameModeSurvival for
// 0, GameModeCreative for 1, GameModeAdventure for 2 and GameModeSpectator for
// 3. If the ID is unknown, the bool returned is false. In this case the
// GameMode returned is GameModeSurvival.
func (reg *gameModeRegistry) Lookup(id int) (GameMode, bool) {
	mode, ok := reg.gameModes[id]
	if !ok {
		mode = GameModeSurvival
	}
	return mode, ok
}

// LookupID looks up the ID that a GameMode was registered with. If not found,
// false is returned.
func (reg *gameModeRegistry) LookupID(mode GameMode) (int, bool) {
	id, ok := reg.ids[mode]
	return id, ok
}

// survival is the survival game mode: Players with this game mode have limited supplies and can break blocks after
// taking some time.
type survival struct{}

func (survival) AllowsEditing() bool      { return true }
func (survival) AllowsTakingDamage() bool { return true }
func (survival) CreativeInventory() bool  { return false }
func (survival) HasCollision() bool       { return true }
func (survival) AllowsFlying() bool       { return false }
func (survival) AllowsInteraction() bool  { return true }
func (survival) Visible() bool            { return true }

// creative represents the creative game mode: Players with this game mode have infinite blocks and
// items and can break blocks instantly. Players with creative mode can also fly.
type creative struct{}

func (creative) AllowsEditing() bool      { return true }
func (creative) AllowsTakingDamage() bool { return false }
func (creative) CreativeInventory() bool  { return true }
func (creative) HasCollision() bool       { return true }
func (creative) AllowsFlying() bool       { return true }
func (creative) AllowsInteraction() bool  { return true }
func (creative) Visible() bool            { return true }

// adventure represents the adventure game mode: Players with this game mode cannot edit the world
// (placing or breaking blocks).
type adventure struct{}

func (adventure) AllowsEditing() bool      { return false }
func (adventure) AllowsTakingDamage() bool { return true }
func (adventure) CreativeInventory() bool  { return false }
func (adventure) HasCollision() bool       { return true }
func (adventure) AllowsFlying() bool       { return false }
func (adventure) AllowsInteraction() bool  { return true }
func (adventure) Visible() bool            { return true }

// spectator represents the spectator game mode: Players with this game mode cannot interact with the
// world and cannot be seen by other players. spectator players can fly, like creative mode, and can
// move through blocks.
type spectator struct{}

func (spectator) AllowsEditing() bool      { return false }
func (spectator) AllowsTakingDamage() bool { return false }
func (spectator) CreativeInventory() bool  { return false }
func (spectator) HasCollision() bool       { return false }
func (spectator) AllowsFlying() bool       { return true }
func (spectator) AllowsInteraction() bool  { return false }
func (spectator) Visible() bool            { return false }
