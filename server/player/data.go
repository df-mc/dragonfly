package player

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

// Data is a struct that contains all the data of that player to be passed on to the Provider and saved.
type Data struct {
	// UUID is the player's unique identifier for their account
	UUID uuid.UUID
	// Username is the last username the player joined with.
	Username string
	// Position is the last position the player was located at.
	// Velocity is the speed at which the player was moving.
	Position, Velocity mgl64.Vec3
	// Yaw and Pitch represent the rotation of the player.
	Yaw, Pitch float64
	// Health, MaxHealth ...
	Health, MaxHealth float64
	// Hunger is the amount of hunger points the player currently has.
	// This should be between 0-20.
	Hunger int
	// FoodTick see player.hungerManager
	FoodTick int
	// ExhaustionLevel, SaturationLevel see player.hungerManager
	ExhaustionLevel, SaturationLevel float64
	// Gamemode is the last gamemode the user had, like creative or survival.
	Gamemode world.GameMode
	// Inventory contains all the items in the inventory, including armor, main inventory and offhand.
	Inventory InventoryData
	// Effects contains all the currently active potions effects the player has.
	Effects []effect.Effect
	// FireTicks is the amount of ticks the player will be on fire for.
	FireTicks int64
	// FallDistance is the distance the player has currently been falling.
	// This is used to calculate fall damage.
	FallDistance float64
}

// InventoryData is a struct that contains all data of the player inventories.
type InventoryData struct {
	// Items contains all the items in the player's main inventory.
	// This excludes armor and offhand.
	Items []item.Stack
	// Armor contains all armor items the player is wearing.
	Armor [4]item.Stack
	// Offhand is what the player is carrying in their non-main hand, like a shield or arrows.
	Offhand item.Stack
	// Mainhand saves the slot in the hotbar that the player is currently switched to.
	// Should be between 0-8.
	Mainhand uint32
}
