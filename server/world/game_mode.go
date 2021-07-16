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
}

// GameModeSurvival represents the survival game mode: Players with this game mode have limited supplies and
// can break blocks using only the right tools.
type GameModeSurvival struct{}

// AllowsEditing ...
func (GameModeSurvival) AllowsEditing() bool {
	return true
}

// AllowsTakingDamage ...
func (GameModeSurvival) AllowsTakingDamage() bool {
	return true
}

// CreativeInventory ...
func (GameModeSurvival) CreativeInventory() bool {
	return false
}

// GameModeCreative represents the creative game mode: Players with this game mode have infinite blocks and
// items and can break blocks instantly. Players with creative mode can also fly.
type GameModeCreative struct{}

// AllowsEditing ...
func (GameModeCreative) AllowsEditing() bool {
	return true
}

// AllowsTakingDamage ...
func (GameModeCreative) AllowsTakingDamage() bool {
	return false
}

// CreativeInventory ...
func (GameModeCreative) CreativeInventory() bool {
	return true
}

// GameModeAdventure represents the adventure game mode: Players with this game mode cannot edit the world
// (placing or breaking blocks).
type GameModeAdventure struct{}

// AllowsEditing ...
func (GameModeAdventure) AllowsEditing() bool {
	return false
}

// AllowsTakingDamage ...
func (GameModeAdventure) AllowsTakingDamage() bool {
	return true
}

// CreativeInventory ...
func (GameModeAdventure) CreativeInventory() bool {
	return false
}

// GameModeSpectator represents the spectator game mode: Players with this game mode cannot interact with the
// world and cannot be seen by other players. GameModeSpectator players can fly, like creative mode, and can
// move through blocks.
type GameModeSpectator struct{}

// AllowsEditing ...
func (GameModeSpectator) AllowsEditing() bool {
	return false
}

// AllowsTakingDamage ...
func (GameModeSpectator) AllowsTakingDamage() bool {
	return false
}

// CreativeInventory ...
func (GameModeSpectator) CreativeInventory() bool {
	return true
}
