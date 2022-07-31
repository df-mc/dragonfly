package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	transform
	age   int
	close bool

	owner world.Entity

	t potion.Potion
	c *ProjectileComputer
}

// SplashableBlock is a block that can be splashed with a splash bottle.
type SplashableBlock interface {
	Splash(pos cube.Pos, potion *SplashPotion)
}

// SplashableEntity is an entity that can be splashed with a splash bottle.
type SplashableEntity interface {
	Splash(potion *SplashPotion)
}

// NewSplashPotion ...
func NewSplashPotion(pos mgl64.Vec3, owner world.Entity, t potion.Potion) *SplashPotion {
	s := &SplashPotion{
		owner: owner,
		t:     t,
		c: &ProjectileComputer{&MovementComputer{
			Gravity:           0.05,
			Drag:              0.01,
			DragBeforeGravity: true,
		}},
	}
	s.transform = newTransform(s, pos)

	return s
}

// Name ...
func (s *SplashPotion) Name() string {
	return "Splash Potion"
}

// EncodeEntity ...
func (s *SplashPotion) EncodeEntity() string {
	return "minecraft:splash_potion"
}

// BBox ...
func (s *SplashPotion) BBox() cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

// Glint returns true if the splash potion should render with glint.
func (s *SplashPotion) Glint() bool {
	return len(s.t.Effects()) > 0
}

// Type returns the type of potion the splash potion will grant effects for when thrown.
func (s *SplashPotion) Type() potion.Potion {
	return s.t
}

// Tick ...
func (s *SplashPotion) Tick(w *world.World, current int64) {
	if s.close {
		_ = s.Close()
		return
	}
	s.mu.Lock()
	m, result := s.c.TickMovement(s, s.pos, s.vel, 0, 0, s.ignores)
	s.pos, s.vel = m.pos, m.vel
	s.mu.Unlock()

	s.age++
	m.Send()

	if m.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		s.close = true
		return
	}

	if result != nil {
		effects := s.t.Effects()
		box := s.BBox().Translate(m.pos)
		colour, _ := effect.ResultingColour(effects)
		if len(effects) > 0 {
			for _, e := range w.EntitiesWithin(box.GrowVec3(mgl64.Vec3{8.25, 4.25, 8.25}), func(entity world.Entity) bool {
				_, living := entity.(Living)
				return !living || entity == s
			}) {
				pos := e.Position()
				if !e.BBox().Translate(pos).IntersectsWith(box.GrowVec3(mgl64.Vec3{4.125, 2.125, 4.125})) {
					continue
				}

				dist := pos.Sub(m.pos).Len()
				if dist > 4 {
					continue
				}

				f := 1 - dist/4
				if entityResult, ok := result.(trace.EntityResult); ok && entityResult.Entity() == e {
					f = 1
				}

				splashed := e.(Living)
				for _, eff := range effects {
					if p, ok := eff.Type().(effect.PotentType); ok {
						splashed.AddEffect(effect.NewInstant(p.WithPotency(f), eff.Level()))
						continue
					}

					dur := time.Duration(float64(eff.Duration()) * 0.75 * f)
					if dur < time.Second {
						continue
					}
					splashed.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
				}
			}
		}
		// Splashing Entities and Blocks
		switch result := result.(type) {
		case trace.BlockResult:
			// we first check to see if it splashed an empty block that's splashable
			pos := result.BlockPosition().Side(result.Face())
			block := w.Block(pos)
			if splashable, ok := block.(SplashableBlock); ok {
				if _, ok := block.Model().(model.Empty); ok {
					splashable.Splash(pos, s)
					// Doesn't run rest of code if it's a splashable empty block
					break
				}
			}
			// splashable non-empty block
			pos = result.BlockPosition()
			if b, ok := w.Block(pos).(SplashableBlock); ok {
				b.Splash(pos, s)
			}
		}
		w.AddParticle(m.pos, particle.Splash{Colour: colour})
		w.PlaySound(m.pos, sound.GlassBreak{})

		s.close = true
	}
}

// ignores returns whether the SplashPotion should ignore collision with the entity passed.
func (s *SplashPotion) ignores(entity world.Entity) bool {
	_, ok := entity.(Living)
	return !ok || entity == s || (s.age < 5 && entity == s.owner)
}

// New creates a SplashPotion with the position, velocity, yaw, and pitch provided. It doesn't spawn the SplashPotion,
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.owner
}

// DecodeNBT decodes the properties in a map to a SplashPotion and returns a new SplashPotion entity.
func (s *SplashPotion) DecodeNBT(data map[string]any) any {
	return s.New(
		nbtconv.MapVec3(data, "Pos"),
		nbtconv.MapVec3(data, "Motion"),
		potion.From(nbtconv.Map[int32](data, "PotionId")),
		nil,
	)
}

// EncodeNBT encodes the SplashPotion entity's properties as a map and returns it.
func (s *SplashPotion) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":      nbtconv.Vec3ToFloat32Slice(s.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(s.Velocity()),
		"PotionId": s.t.Uint8(),
		"Yaw":      0.0,
		"Pitch":    0.0,
	}
}
