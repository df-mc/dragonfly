package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// splashable is a struct that can be embedded by splashable projectiles, such as splash potions or lingering potions.
type splashable struct {
	m float64
	t potion.Potion
}

// SplashableBlock is a block that can be splashed with a splash bottle.
type SplashableBlock interface {
	world.Block
	// Splash is called when a type that implements splashable splashes onto a block.
	Splash(w *world.World, pos cube.Pos, p potion.Potion)
}

// SplashableEntity is an entity that can be splashed with a splash bottle.
type SplashableEntity interface {
	world.Entity
	// Splash is called when a type that implements splashable splashes onto a block.
	Splash(w *world.World, pos mgl64.Vec3, p potion.Potion)
}

// Glint returns true if the splashable should render with glint.
func (s *splashable) Glint() bool {
	return len(s.t.Effects()) > 0
}

// Type returns the type of potion the splashable will grant effects for when thrown.
func (s *splashable) Type() potion.Potion {
	return s.t
}

// splash splashes the projectile at the given position.
func (s *splashable) splash(e world.Entity, w *world.World, pos mgl64.Vec3, res trace.Result, box cube.BBox) {
	effects := s.t.Effects()
	box = box.Translate(pos)
	colour, _ := effect.ResultingColour(effects)
	if len(effects) > 0 {
		for _, otherE := range w.EntitiesWithin(box.GrowVec3(mgl64.Vec3{8.25, 4.25, 8.25}), func(entity world.Entity) bool {
			_, living := entity.(Living)
			return !living || entity == e
		}) {
			otherPos := otherE.Position()
			if !otherE.BBox().Translate(otherPos).IntersectsWith(box.GrowVec3(mgl64.Vec3{4.125, 2.125, 4.125})) {
				continue
			}

			dist := otherPos.Sub(pos).Len()
			if dist > 4 {
				continue
			}

			f := 1 - dist/4
			if entityResult, ok := res.(trace.EntityResult); ok && entityResult.Entity() == e {
				f = 1
			}

			splashed := otherE.(Living)
			for _, eff := range effects {
				if p, ok := eff.Type().(effect.PotentType); ok {
					splashed.AddEffect(effect.NewInstant(p.WithPotency(f), eff.Level()))
					continue
				}

				dur := time.Duration(float64(eff.Duration()) * s.m * f)
				if dur < time.Second {
					continue
				}
				splashed.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
			}
		}
	}
	switch result := res.(type) {
	case trace.BlockResult:
		pos := result.BlockPosition().Side(res.Face())
		if b, ok := w.Block(pos).(SplashableBlock); ok {
			if _, ok := b.Model().(model.Empty); ok {
				b.Splash(w, pos, s.Type())
				break
			}
		}

		pos = result.BlockPosition()
		if b, ok := w.Block(pos).(SplashableBlock); ok {
			b.Splash(w, pos, s.Type())
		}
	case trace.EntityResult:
		if e, ok := result.Entity().(SplashableEntity); ok {
			e.Splash(w, pos, s.Type())
		}
	}
	w.AddParticle(pos, particle.Splash{Colour: colour})
	w.PlaySound(pos, sound.GlassBreak{})
}
