package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"image/color"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	transform
	yaw, pitch float64

	ticksLived int

	closeNextTick bool

	owner world.Entity

	variant potion.Potion

	c *ProjectileComputer
}

// NewSplashPotion ...
func NewSplashPotion(pos mgl64.Vec3, yaw, pitch float64, owner world.Entity) *SplashPotion {
	s := &SplashPotion{
		yaw:   yaw,
		pitch: pitch,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
		owner: owner,
	}
	s.transform = newTransform(s, pos)

	return s
}

// Name ...
func (s *SplashPotion) Name() string {
	return "Splash Variant"
}

// EncodeEntity ...
func (s *SplashPotion) EncodeEntity() string {
	return "minecraft:splash_potion"
}

// AABB ...
func (s *SplashPotion) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{-0.125, 0, -0.125}, mgl64.Vec3{0.125, 0.25, 0.125})
}

// SetVariant ...
func (s *SplashPotion) SetVariant(variant potion.Potion) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variant = variant
}

// Variant ...
func (s *SplashPotion) Variant() potion.Potion {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.variant
}

// Rotation ...
func (s *SplashPotion) Rotation() (float64, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.yaw, s.pitch
}

// Tick ...
func (s *SplashPotion) Tick(current int64) {
	if s.closeNextTick {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, s.yaw, s.pitch, s.ignores)
	s.pos, s.vel, s.yaw, s.pitch = m.pos, m.vel, m.yaw, m.pitch
	s.mu.Unlock()

	s.ticksLived++
	m.Send()

	if m.pos[1] < cube.MinY && current%10 == 0 {
		s.closeNextTick = true
		return
	}

	if result != nil {
		w := s.World()
		pos := s.Position()

		pot := s.variant
		effects := pot.Effects
		hasEffects := len(effects) > 0

		colour := color.RGBA{R: 0x38, G: 0x5d, B: 0xc6, A: 0xff}
		if hasEffects {
			colour, _ = effect.ResultingColour(effects)
		}

		w.AddParticle(pos, particle.Splash{Colour: colour})
		w.PlaySound(pos, sound.GlassBreak{})

		// TODO: Actually apply the effects to the nearby entities.

		s.closeNextTick = true
	}
}

// ignores returns whether the SplashPotion should ignore collision with the entity passed.
func (s *SplashPotion) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == s || (s.ticksLived < 5 && entity == s.owner)
}

// New creates a SplashPotion with the position, velocity, yaw, and pitch provided. It doesn't spawn the SplashPotion,
// only returns it.
func (s *SplashPotion) New(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity {
	splash := NewSplashPotion(pos, yaw, pitch, nil)
	splash.vel = vel
	return splash
}

// Owner ...
func (s *SplashPotion) Owner() world.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// Own ...
func (s *SplashPotion) Own(owner world.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owner = owner
}

// DecodeNBT decodes the properties in a map to a SplashPotion and returns a new SplashPotion entity.
func (s *SplashPotion) DecodeNBT(data map[string]interface{}) interface{} {
	return s.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		float64(nbtconv.MapFloat32(data, "Pitch")),
		float64(nbtconv.MapFloat32(data, "Yaw")),
	)
}

// EncodeNBT encodes the SplashPotion entity's properties as a map and returns it.
func (s *SplashPotion) EncodeNBT() map[string]interface{} {
	yaw, pitch := s.Rotation()
	return map[string]interface{}{
		"Pos":    nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Yaw":    yaw,
		"Pitch":  pitch,
		"Motion": nbtconv.Vec3ToFloat32Slice(s.Velocity()),
		"Damage": 0.0,
	}
}
