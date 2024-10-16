package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// TNT is an explosive block that can be primed to generate an explosion.
type TNT struct {
	solid
	igniter world.Entity
}

// Activate ...
func (t TNT) Activate(pos cube.Pos, clickedFace cube.Face, tx *world.Tx, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Enchantment(enchantment.FireAspect{}); ok {
		t.Ignite(pos, tx, u)
		ctx.DamageItem(1)
		return true
	}
	return false
}

// Ignite ...
func (t TNT) Ignite(pos cube.Pos, tx *world.Tx, igniter world.Entity) bool {
	t.igniter = igniter
	spawnTnt(pos, tx, time.Second*4, t.igniter)
	return true
}

// Igniter returns the entity that ignited the TNT.
// It is nil if ignited by a world source like fire.
func (t TNT) Igniter() world.Entity {
	return t.igniter
}

// Explode ...
func (t TNT) Explode(explosionPos mgl64.Vec3, pos cube.Pos, tx *world.Tx, c ExplosionConfig) {
	spawnTnt(pos, tx, time.Second/2+time.Duration(rand.Intn(int(time.Second+time.Second/2))), t.igniter)
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
func spawnTnt(pos cube.Pos, tx *world.Tx, fuse time.Duration, igniter world.Entity) {
	tx.PlaySound(pos.Vec3Centre(), sound.TNT{})
	tx.SetBlock(pos, nil, nil)
	tx.AddEntity(tx.World().EntityRegistry().Config().TNT(pos.Vec3Centre(), fuse, igniter))
}
