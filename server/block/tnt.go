package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// TNT is an explosive block that can be primed to generate an explosion.
type TNT struct {
	solid
}

var _ world.RedstonePowerAction = TNT{}

func (TNT) RedstoneNonConductive() {}

// RedstonePowerAction primes TNT when it first receives redstone power.
func (t TNT) RedstonePowerAction(pos cube.Pos, tx *world.Tx, oldPower, newPower int) {
	if oldPower > 0 || newPower == 0 {
		return
	}
	t.Ignite(pos, tx, nil)
}

// ProjectileHit ignites the TNT if the projectile that hit it is on fire, such as a flaming arrow. The
// resulting explosion is credited to the entity that shot the projectile rather than to the projectile itself.
func (t TNT) ProjectileHit(pos cube.Pos, tx *world.Tx, e world.Entity, _ cube.Face) {
	if f, ok := e.(flammableEntity); ok && f.OnFireDuration() > 0 {
		spawnTnt(pos, tx, time.Second*4, tntIgnitionSourceHandle(e), true)
	}
}

// Activate ignites the TNT if the user activates it with an item enchanted with Fire Aspect, crediting the
// user for the resulting explosion and costing the item a point of durability.
func (t TNT) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Enchantment(enchantment.FireAspect); ok {
		t.Ignite(pos, tx, u)
		ctx.DamageItem(1)
		return true
	}
	return false
}

// Ignite primes the TNT, replacing the block with a primed TNT entity that has a fuse of four seconds. source
// is the (nullable) entity that lit the TNT: it is credited for the resulting explosion, so that a player
// cannot shield themselves from a blast they set off. Ignite always returns true, as TNT is always ignitable.
func (t TNT) Ignite(pos cube.Pos, tx *world.Tx, source world.Entity) bool {
	spawnTnt(pos, tx, time.Second*4, tntIgnitionSourceHandle(source), true)
	return true
}

// Explode primes the TNT instead of destroying it when it is caught in another explosion, giving it a random
// fuse between half a second and two seconds so that a chain reaction does not detonate all at once. The
// primed TNT inherits the source and the shield-blockability of the explosion that lit it, keeping the whole
// chain attributed to whoever set it off.
func (t TNT) Explode(_ mgl64.Vec3, pos cube.Pos, tx *world.Tx, c ExplosionConfig) {
	var source *world.EntityHandle
	if c.Source != nil {
		source = c.Source.H()
	}
	spawnTnt(pos, tx, time.Second/2+time.Duration(rand.IntN(int(time.Second+time.Second/2))), source, !c.UnblockableByShield)
}

// BreakInfo ...
func (t TNT) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// FlammabilityInfo ...
func (t TNT) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(15, 100, true)
}

// EncodeItem ...
func (t TNT) EncodeItem() (name string, meta int16) {
	return "minecraft:tnt", 0
}

// EncodeBlock ...
func (t TNT) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:tnt", map[string]interface{}{"explode_bit": false}
}

// ownerEntity is implemented by entities that were shot by another entity, such as projectiles.
type ownerEntity interface {
	ProjectileOwner() *world.EntityHandle
}

// tntIgnitionSourceHandle resolves the entity that should be credited for igniting TNT. Projectiles are
// attributed to their shooter, so that a player who lights TNT with a flaming arrow cannot block their own
// blast with a shield. It returns nil if the ignition had no entity source, e.g. redstone.
func tntIgnitionSourceHandle(source world.Entity) *world.EntityHandle {
	if source == nil {
		return nil
	}
	if o, ok := source.(ownerEntity); ok {
		if owner := o.ProjectileOwner(); owner != nil {
			return owner
		}
	}
	return source.H()
}

// spawnTnt creates a new TNT entity at the given position with the given fuse duration. source is the
// (nullable) entity credited for the ignition and blockableByShield reports whether the resulting explosion
// may be blocked by a shield, both of which are carried over to the primed TNT entity.
func spawnTnt(pos cube.Pos, tx *world.Tx, fuse time.Duration, source *world.EntityHandle, blockableByShield bool) {
	tx.PlaySound(pos.Vec3Centre(), sound.TNT{})
	tx.SetBlock(pos, nil, nil)
	opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
	conf := tx.World().EntityRegistry().Config()
	tx.AddEntity(conf.TNTWithSource(opts, fuse, source, blockableByShield))
}
