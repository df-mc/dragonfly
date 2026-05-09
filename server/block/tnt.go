package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// TNT is an explosive block that can be primed to generate an explosion.
type TNT struct {
	solid
}

// ProjectileHit ...
func (t TNT) ProjectileHit(pos cube.Pos, tx *world.Tx, e world.Entity, _ cube.Face) {
	if f, ok := e.(flammableEntity); ok && f.OnFireDuration() > 0 {
		spawnTnt(pos, tx, time.Second*4, tntIgnitionSourceHandle(e), true)
	}
}

// Activate ...
func (t TNT) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Enchantment(enchantment.FireAspect); ok {
		t.Ignite(pos, tx, u)
		ctx.DamageItem(1)
		return true
	}
	return false
}

// Ignite ...
func (t TNT) Ignite(pos cube.Pos, tx *world.Tx, source world.Entity) bool {
	spawnTnt(pos, tx, time.Second*4, entityHandle(source), source != nil)
	return true
}

// Explode ...
func (t TNT) Explode(_ mgl64.Vec3, pos cube.Pos, tx *world.Tx, c ExplosionConfig) {
	spawnTnt(pos, tx, time.Second/2+time.Duration(rand.IntN(int(time.Second+time.Second/2))), tntExplosionSourceHandle(c), !c.UnblockableByShield)
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

// spawnTnt creates a new TNT entity at the given position with the given fuse duration.
type ownerEntity interface {
	ProjectileOwner() *world.EntityHandle
}

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

func tntExplosionSourceHandle(c ExplosionConfig) *world.EntityHandle {
	return entityHandle(c.Source)
}

func entityHandle(e world.Entity) *world.EntityHandle {
	if e == nil {
		return nil
	}
	return e.H()
}

func spawnTnt(pos cube.Pos, tx *world.Tx, fuse time.Duration, source *world.EntityHandle, blockableByShield bool) {
	tx.PlaySound(pos.Vec3Centre(), sound.TNT{})
	tx.SetBlock(pos, nil, nil)
	opts := world.EntitySpawnOpts{Position: pos.Vec3Centre()}
	conf := tx.World().EntityRegistry().Config()
	if (source != nil || blockableByShield) && conf.TNTWithSource != nil {
		var e world.Entity
		if source != nil {
			e, _ = source.Entity(tx)
		}
		tx.AddEntity(conf.TNTWithSource(opts, fuse, e, blockableByShield))
		return
	}
	tx.AddEntity(conf.TNT(opts, fuse))
}
