package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
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
func (t TNT) Explode(pos cube.Pos, c ExplosionConfig) {
	ent, ok := world.EntityByName("minecraft:tnt")
	if !ok {
		return
	}

	c.World.SetBlock(pos, nil, nil)
	if p, ok := ent.(interface {
		New(pos mgl64.Vec3, fuse time.Duration) world.Entity
	}); ok {
		c.World.AddEntity(p.New(pos.Vec3Centre(), time.Second/2+time.Duration(rand.Intn(int(time.Second+time.Second/2)))))
	}
}

// BreakInfo ...
func (t TNT) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(t))
}

// EncodeItem ...
func (t TNT) EncodeItem() (name string, meta int16) {
	return "minecraft:tnt", 0
}

// EncodeBlock ...
func (t TNT) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:tnt", map[string]interface{}{"allow_underwater_bit": false, "explode_bit": false}
}
