package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	splashable
	transform

	close bool

	owner world.Entity

	c *ProjectileComputer
}

// NewSplashPotion ...
func NewSplashPotion(pos mgl64.Vec3, owner world.Entity, t potion.Potion) *SplashPotion {
	s := &SplashPotion{
		owner:      owner,
		splashable: splashable{t: t, m: 0.75},
		c:          newProjectileComputer(0.05, 0.01),
	}
	s.transform = newTransform(s, pos)
	return s
}

// Type returns SplashPotionType.
func (*SplashPotion) Type() world.EntityType {
	return SplashPotionType{}
}

// Tick ...
func (s *SplashPotion) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, 0, 0)
	s.pos, s.vel = m.pos, m.vel
	s.mu.Unlock()

	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}

	if result != nil {
		s.splash(s, w, m.pos, result, s.Type().BBox(s))
		s.close = true
	}
}

// New creates a SplashPotion with the position and velocity provided. It doesn't spawn the SplashPotion,
// only returns it.
func (s *SplashPotion) New(pos, vel mgl64.Vec3, t potion.Potion, owner world.Entity) world.Entity {
	splash := NewSplashPotion(pos, owner, t)
	splash.vel = vel
	splash.owner = owner
	return splash
}

// Explode ...
func (s *SplashPotion) Explode(explosionPos mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	s.mu.Lock()
	s.vel = s.vel.Add(s.pos.Sub(explosionPos).Normalize().Mul(impact))
	s.mu.Unlock()
}

// Owner ...
func (s *SplashPotion) Owner() world.Entity {
	return s.owner
}

// SplashPotionType is a world.EntityType implementation for SplashPotion.
type SplashPotionType struct{}

func (SplashPotionType) EncodeEntity() string { return "minecraft:splash_potion" }
func (SplashPotionType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (SplashPotionType) DecodeNBT(m map[string]any) world.Entity {
	pot := NewSplashPotion(nbtconv.Vec3(m, "Pos"), nil, potion.From(nbtconv.Int32(m, "PotionId")))
	pot.vel = nbtconv.Vec3(m, "Motion")
	return pot
}

func (SplashPotionType) EncodeNBT(e world.Entity) map[string]any {
	pot := e.(*SplashPotion)
	return map[string]any{
		"Pos":      nbtconv.Vec3ToFloat32Slice(pot.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(pot.Velocity()),
		"PotionId": int32(pot.t.Uint8()),
	}
}
