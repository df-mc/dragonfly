package item

// Shooter represents an item that is able to shoot projectiles, similarly to a bow or crossbow. Ammunition used by
// the item must implement Projectile in order to function properly.
type Shooter interface {
	// ShooterInfo returns the shooter information of the item.
	ShooterInfo() ShooterInfo
}

// ShooterInfo is a struct returned by items that implement Shooter. It contains the information required for the
// client to shoot projectiles using the item.
type ShooterInfo struct {
	// Ammunition is a list of ammunition entries that define which items can be used as projectiles for this
	// shooter.
	Ammunition []Ammunition
	// MaxDrawDuration is the maximum time in seconds that a player can draw the shooter before it automatically
	// fires or reaches maximum power.
	MaxDrawDuration float64
	// ChargeOnDraw is true if the shooter begins charging when the player starts drawing, similar to a crossbow.
	ChargeOnDraw bool
	// ScalePowerByDrawDuration is true if the projectile's launch power increases based on how long the player
	// holds the use button before releasing.
	ScalePowerByDrawDuration bool
}

// Ammunition represents an entry of ammunition that can be used by a Shooter. It specifies the item to be used as
// a projectile and where it may be taken from.
type Ammunition struct {
	// Item is the identifier of the item used as ammunition.
	Item string
	// SearchInventory is true if the inventory of the user may be searched for the ammunition.
	SearchInventory bool
	// UseInCreative is true if the ammunition may be used in creative mode.
	UseInCreative bool
	// UseOffHand is true if the off-hand of the user may be used for the ammunition.
	UseOffHand bool
}
