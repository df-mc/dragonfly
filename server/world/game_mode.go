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

// HasCollision ...
func (GameModeSurvival) HasCollision() bool {
	return true
}

// AllowsFlying ...
func (GameModeSurvival) AllowsFlying() bool {
	return false
}

// AllowsInteraction ...
func (GameModeSurvival) AllowsInteraction() bool {
	return true
}

// Visible ...
func (GameModeSurvival) Visible() bool {
	return true
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

// HasCollision ...
func (GameModeCreative) HasCollision() bool {
	return true
}

// AllowsFlying ...
func (GameModeCreative) AllowsFlying() bool {
	return true
}

// AllowsInteraction ...
func (GameModeCreative) AllowsInteraction() bool {
	return true
}

// Visible ...
func (GameModeCreative) Visible() bool {
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

// HasCollision ...
func (GameModeAdventure) HasCollision() bool {
	return true
}

// AllowsFlying ...
func (GameModeAdventure) AllowsFlying() bool {
	return false
}

// AllowsInteraction ...
func (GameModeAdventure) AllowsInteraction() bool {
	return true
}

// Visible ...
func (GameModeAdventure) Visible() bool {
	return true
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

// HasCollision ...
func (GameModeSpectator) HasCollision() bool {
	return false
}

// AllowsFlying ...
func (GameModeSpectator) AllowsFlying() bool {
	return true
}

// AllowsInteraction ...
func (GameModeSpectator) AllowsInteraction() bool {
	return true
}

// Visible ...
func (GameModeSpectator) Visible() bool {
	return false
}
