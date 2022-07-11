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
	// UUID is the player's unique identifier for their account.
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
	// Hunger is the amount of hunger points the player currently has, shown on the hunger bar.
	// This should be between 0-20.
	Hunger int
	// FoodTick this variable is used when the hunger exceeds 17 or is equal to 0. It is used to heal
	// the player using saturation or make the player starve and receive damage if the hunger is at 0.
	// This value should be between 0-80.
	FoodTick int
	// ExhaustionLevel determines how fast the hunger level depletes and is controlled by the kinds
	// of food the player has eaten. SaturationLevel determines how fast the saturation level depletes.
	ExhaustionLevel, SaturationLevel float64
	// Experience is the current experience the player has.
	Experience int
	// AirSupply is the current tick of the player's air supply.
	AirSupply int64
	// MaxAirSupply is the maximum air supply the player can have.
	MaxAirSupply int64
	// GameMode is the last gamemode the user had, like creative or survival.
	GameMode world.GameMode
	// Inventory contains all the items in the inventory, including armor, main inventory and offhand.
	Inventory InventoryData
	// Effects contains all the currently active potions effects the player has.
	Effects []effect.Effect
	// FireTicks is the amount of ticks the player will be on fire for.
	FireTicks int64
	// FallDistance is the distance the player has currently been falling. This is used to calculate fall damage.
	FallDistance float64
	// Dimension is the ID of the dimension that the player was last in. The player is added to the correct world based
	// on this number.
	Dimension int
}

// InventoryData is a struct that contains all data of the player inventories.
type InventoryData struct {
	// Items contains all the items in the player's main inventory.
	// This excludes armor and offhand.
	Items []item.Stack
	// Boots, Leggings, Chestplate, Helmet are armor pieces that belong to the slot corresponding to the name.
	Boots      item.Stack
	Leggings   item.Stack
	Chestplate item.Stack
	Helmet     item.Stack
	// OffHand is what the player is carrying in their non-main hand, like a shield or arrows.
	OffHand item.Stack
	// MainHandSlot saves the slot in the hotbar that the player is currently switched to.
	// Should be between 0-8.
	MainHandSlot uint32
}
