package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// potionSplash returns a function that creates a potion splash with a specific
// duration multiplier and potion type.
func potionSplash(durMul float64, pot potion.Potion, linger bool) func(e *Ent, res trace.Result) {
	return func(e *Ent, res trace.Result) {
		w, pos := e.World(), e.Position()

		effects := pot.Effects()
		box := e.Type().BBox(e).Translate(pos)

		ignores := func(entity world.Entity) bool {
			_, living := entity.(Living)
			return !living || entity == e
		}
		if len(effects) > 0 {
			for _, otherE := range w.EntitiesWithin(box.GrowVec3(mgl64.Vec3{8.25, 4.25, 8.25}), ignores) {
				otherPos := otherE.Position()
				if !otherE.Type().BBox(otherE).Translate(otherPos).IntersectsWith(box.GrowVec3(mgl64.Vec3{4.125, 2.125, 4.125})) {
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

					dur := time.Duration(float64(eff.Duration()) * durMul * f)
					if dur < time.Second {
						continue
					}
					splashed.AddEffect(effect.New(eff.Type().(effect.LastingType), eff.Level(), dur))
				}
			}
		} else if pot == potion.Water() {
			switch result := res.(type) {
			case trace.BlockResult:
				blockPos := result.BlockPosition().Side(result.Face())
				if w.Block(blockPos) == fire() {
					w.SetBlock(blockPos, nil, nil)
				}

				for _, f := range cube.HorizontalFaces() {
					if h := blockPos.Side(f); w.Block(h) == fire() {
						w.SetBlock(h, nil, nil)
					}
				}
			case trace.EntityResult:
				// TODO: Damage endermen, blazes, striders and snow golems when implemented and rehydrate axolotls.
			}
		}
		if linger {
			w.AddEntity(NewAreaEffectCloud(pos, pot))
		}
	}
}
