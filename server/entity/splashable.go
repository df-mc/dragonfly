package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
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
	t potion.Potion
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

				dur := time.Duration(float64(eff.Duration()) * 0.75 * f)
				if dur < time.Second {
					continue
				}
				splashed.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
			}
		}
	} else if s.t == potion.Water() {
		switch result := res.(type) {
		case trace.BlockResult:
			blockPos := result.BlockPosition().Side(result.Face())
			if w.Block(blockPos) == fire() {
				w.SetBlock(blockPos, air(), nil)
			}

			for _, f := range cube.HorizontalFaces() {
				if h := blockPos.Side(f); w.Block(h) == fire() {
					w.SetBlock(h, air(), nil)
				}
			}
		case trace.EntityResult:
			// TODO: Damage endermen, blazes, striders and snow golems when implemented and rehydrate axolotls.
		}
	}

	w.AddParticle(pos, particle.Splash{Colour: colour})
	w.PlaySound(pos, sound.GlassBreak{})
}

// air returns an air block.
func air() world.Block {
	f, ok := world.BlockByName("minecraft:air", map[string]any{})
	if !ok {
		panic("could not find air block")
	}
	return f
}
