package item

// Projectile represents an item that is a projectile, which may be shot from dispensers or used as ammunition by
// items implementing Shooter. When combined with Throwable, this specifies the entity spawned when the item is
// thrown.
type Projectile interface {
	// ProjectileInfo returns the projectile information of the item.
	ProjectileInfo() ProjectileInfo
}

// ProjectileInfo is a struct returned by items that implement Projectile. It contains the information required for
// the client to use the item as a projectile.
type ProjectileInfo struct {
	// ProjectileEntity is the identifier of the entity fired as a projectile.
	ProjectileEntity string
	// MinimumCriticalPower is how long a player must charge a projectile for it to critically hit, in seconds.
	MinimumCriticalPower float64
}
