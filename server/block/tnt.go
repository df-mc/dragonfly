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

// ProjectileHit ignites TNT hit by a burning projectile, attributing it to the projectile owner.
func (t TNT) ProjectileHit(pos cube.Pos, tx *world.Tx, e world.Entity, _ cube.Face) {
	if f, ok := e.(flammableEntity); ok && f.OnFireDuration() > 0 {
		spawnTnt(pos, tx, time.Second*4, tntIgnitionSourceHandle(e), true)
	}
}

// Activate ignites TNT using a Fire Aspect item.
func (t TNT) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Enchantment(enchantment.FireAspect); ok {
		t.Ignite(pos, tx, u)
		ctx.DamageItem(1)
		return true
	}
	return false
}

// Ignite primes the TNT with source credited for the resulting explosion.
func (t TNT) Ignite(pos cube.Pos, tx *world.Tx, source world.Entity) bool {
	spawnTnt(pos, tx, time.Second*4, tntIgnitionSourceHandle(source), true)
	return true
}

// Explode primes TNT with a short random fuse, preserving the explosion source and shield blockability.
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

// ownerEntity exposes the owner of a projectile-like entity.
type ownerEntity interface {
	ProjectileOwner() *world.EntityHandle
}

// tntIgnitionSourceHandle returns the entity credited for an ignition, resolving projectile owners.
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

// spawnTnt replaces the block with primed TNT carrying its explosion source and shield blockability.
func spawnTnt(pos cube.Pos, tx *world.Tx, fuse time.Duration, source *world.EntityHandle, blockableByShield bool) {
	tx.PlaySound(pos.Vec3Centre(), sound.TNT{})
	tx.SetBlock(pos, nil, nil)
	opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
	conf := tx.World().EntityRegistry().Config()
	tx.AddEntity(conf.TNTWithSource(opts, fuse, source, blockableByShield))
}
