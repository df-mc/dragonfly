package session

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/player/skin"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/gamemode"
	"github.com/google/uuid"
)

// Controllable represents an entity that may be controlled by a Session. Generally, a Controllable is
// implemented in the form of a Player.
// Methods in Controllable will be added as Session needs them in order to handle packets.
type Controllable interface {
	world.CarryingEntity

	Chat(message string)
	ExecuteCommand(commandLine string)
	SetGameMode(mode gamemode.GameMode)
	BreakBlock(pos world.BlockPos) error

	// Name returns the display name of the controllable. This name is shown in-game to other viewers of the
	// world.
	Name() string
	// UUID returns the UUID of the controllable. It must be unique for all controllables present in the
	// server.
	UUID() uuid.UUID
	// XUID returns the XBOX Live User ID of the controllable. Every controllable must have one of these, as
	// they must be connected to an XBOX Live account.
	XUID() string
	// Skin returns the skin of the controllable. Each controllable must have a skin, as it defines how the
	// entity looks in the world.
	Skin() skin.Skin
}
