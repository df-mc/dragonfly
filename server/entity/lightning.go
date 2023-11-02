package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// NewLightning creates a lightning entity. The lightning entity will be
// positioned at the position passed. Lightning is a lethal element to
// thunderstorms. Lightning momentarily increases the skylight's brightness to
// slightly greater than full daylight.
func NewLightning(pos mgl64.Vec3) *Ent {
	return NewLightningWithDamage(pos, 5, true, time.Second*8)
}

// NewLightningWithDamage creates a new lightning entities using the damage and
// fire properties passed.
func NewLightningWithDamage(pos mgl64.Vec3, dmg float64, blockFire bool, entityFireDuration time.Duration) *Ent {
	state := &lightningState{
		Damage:             dmg,
		EntityFireDuration: entityFireDuration,
		BlockFire:          blockFire,
		state:              2,
		lifetime:           rand.Intn(4) + 1,
	}
	conf := lightningConf
	conf.Tick = state.tick
	return Config{Behaviour: conf.New()}.New(LightningType{}, pos)
}

var lightningConf = StationaryBehaviourConfig{SpawnSounds: []world.Sound{sound.Explosion{}, sound.Thunder{}}}

// lightningState holds the state of a lightning entity.
type lightningState struct {
	Damage             float64
	EntityFireDuration time.Duration
	BlockFire          bool
	state, lifetime    int
}

// tick carries out lightning logic, dealing damage and setting blocks/entities
// on fire when appropriate.
func (s *lightningState) tick(e *Ent) {
	w, pos := e.World(), e.Position()

	if s.state--; s.state < 0 {
		if s.lifetime == 0 {
			_ = e.Close()
		} else if s.state < -rand.Intn(10) {
			s.lifetime--
			s.state = 1

			if s.BlockFire && w.Difficulty().FireSpreadIncrease() >= 10 {
				s.spreadFire(w, cube.PosFromVec3(pos))
			}
		}
	}
	if s.state > 0 {
		s.dealDamage(e)
	}
}

// dealDamage deals damage to all entities around the lightning and sets them
// on fire.
func (s *lightningState) dealDamage(e *Ent) {
	w, pos := e.World(), e.Position()
	bb := e.Type().BBox(e).GrowVec3(mgl64.Vec3{3, 6, 3}).Translate(pos.Add(mgl64.Vec3{0, 3}))
	for _, e := range w.EntitiesWithin(bb, nil) {
		// Only damage entities that weren't already dead.
		if l, ok := e.(Living); ok && l.Health() > 0 {
			if s.Damage > 0 {
				l.Hurt(s.Damage, LightningDamageSource{})
			}
			if f, ok := e.(Flammable); ok && f.OnFireDuration() < s.EntityFireDuration {
				f.SetOnFire(s.EntityFireDuration)
			}
		}
	}
}

// spreadFire attempts to place fire at the position of the lightning and does
// 4 additional attempts to spread it around that position.
func (s *lightningState) spreadFire(w *world.World, pos cube.Pos) {
	s.fire().Start(w, pos)
	for i := 0; i < 4; i++ {
		pos.Add(cube.Pos{rand.Intn(3) - 1, rand.Intn(3) - 1, rand.Intn(3) - 1})
		s.fire().Start(w, pos)
	}
}

// fire returns a fire block.
func (s *lightningState) fire() interface {
	Start(w *world.World, pos cube.Pos)
} {
	return fire().(interface {
		Start(w *world.World, pos cube.Pos)
	})
}

// fire returns a fire block.
func fire() world.Block {
	f, ok := world.BlockByName("minecraft:fire", map[string]any{"age": int32(0)})
	if !ok {
		panic("could not find fire block")
	}
	return f
}

// LightningType is a world.EntityType implementation for Lightning.
type LightningType struct{}

func (LightningType) EncodeEntity() string                  { return "minecraft:lightning_bolt" }
func (LightningType) BBox(world.Entity) cube.BBox           { return cube.BBox{} }
func (LightningType) DecodeNBT(map[string]any) world.Entity { return nil }
func (LightningType) EncodeNBT(world.Entity) map[string]any {
	return map[string]any{}
}
