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
}

// Activate ...
func (t TNT) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	held, _ := u.HeldItems()
	if _, ok := held.Enchantment(enchantment.FireAspect{}); ok {
		t.Ignite(pos, w)
		ctx.DamageItem(1)
		return true
	}
	return false
}

// Ignite ...
func (t TNT) Ignite(pos cube.Pos, w *world.World) bool {
	ent, ok := world.EntityByName("minecraft:tnt")
	if !ok {
		return false
	}

	w.PlaySound(pos.Vec3Centre(), sound.TNT{})
	w.SetBlock(pos, nil, nil)
	if p, ok := ent.(interface {
		New(pos mgl64.Vec3, fuse time.Duration) world.Entity
	}); ok {
		w.AddEntity(p.New(pos.Vec3Centre(), time.Second*4))
	}
	return true
}

// Explode ...
func (t TNT) Explode(_ mgl64.Vec3, pos cube.Pos, w *world.World, c ExplosionConfig) {
	ent, ok := world.EntityByName("minecraft:tnt")
	if !ok {
		return
	}

	w.SetBlock(pos, nil, nil)
	if p, ok := ent.(interface {
		New(pos mgl64.Vec3, fuse time.Duration) world.Entity
	}); ok {
		w.AddEntity(p.New(pos.Vec3Centre(), time.Second/2+time.Duration(rand.Intn(int(time.Second+time.Second/2)))))
	}
}

// BreakInfo ...
func (t TNT) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// FlammabilityInfo ...
func (t TNT) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// EncodeItem ...
func (t TNT) EncodeItem() (name string, meta int16) {
	return "minecraft:tnt", 0
}

// EncodeBlock ...
func (t TNT) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:tnt", map[string]interface{}{"allow_underwater_bit": false, "explode_bit": false}
}
