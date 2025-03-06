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

// SplashableBlock is a block that can be splashed with a splash bottle.
type SplashableBlock interface {
	world.Block
	// Splash is called when a water bottle splashes onto a block.
	Splash(tx *world.Tx, pos cube.Pos)
}

// SplashableEntity is an entity that can be splashed with a splash bottle.
type SplashableEntity interface {
	world.Entity
	// Splash is called when a water bottle splashes onto an entity.
	Splash(tx *world.Tx, pos mgl64.Vec3)
}

// potionSplash returns a function that creates a potion splash with a specific
// duration multiplier and potion type.
func potionSplash(durMul float64, pot potion.Potion, linger bool) func(e *Ent, tx *world.Tx, res trace.Result) {
	return func(e *Ent, tx *world.Tx, res trace.Result) {
		pos := e.Position()
		effects := pot.Effects()
		box := e.H().Type().BBox(e).Translate(pos)

		if len(effects) > 0 {
			for otherE := range filterLiving(tx.EntitiesWithin(box.GrowVec3(mgl64.Vec3{8.25, 4.25, 8.25}))) {
				otherPos := otherE.Position()
				if !otherE.H().Type().BBox(otherE).Translate(otherPos).IntersectsWith(box.GrowVec3(mgl64.Vec3{4.125, 2.125, 4.125})) {
					continue
				}

				dist := otherPos.Sub(pos).Len()
				if dist > 4 {
					continue
				}

				f := 1 - dist/4
				if entityResult, ok := res.(trace.EntityResult); ok && entityResult.Entity().H() == otherE.H() {
					f = 1
				}

				splashed := otherE.(Living)
				for _, eff := range effects {
					if _, ok := eff.Type().(effect.LastingType); !ok {
						splashed.AddEffect(effect.NewInstantWithPotency(eff.Type(), eff.Level(), f))
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
				if tx.Block(blockPos) == fire() {
					tx.SetBlock(blockPos, nil, nil)
				}

				for _, f := range cube.HorizontalFaces() {
					if h := blockPos.Side(f); tx.Block(h) == fire() {
						tx.SetBlock(h, nil, nil)
					}

					if b, ok := tx.Block(blockPos.Side(f)).(SplashableBlock); ok {
						b.Splash(tx, blockPos.Side(f))
					}
				}

				resultPos := result.BlockPosition()
				if b, ok := tx.Block(resultPos).(SplashableBlock); ok {
					b.Splash(tx, resultPos)
				}
			case trace.EntityResult:
				// TODO: Damage endermen, blazes, striders and snow golems when implemented and rehydrate axolotls.
			}

			for otherE := range filterLiving(tx.EntitiesWithin(box.GrowVec3(mgl64.Vec3{8.25, 4.25, 8.25}))) {
				if splashE, ok := otherE.(SplashableEntity); ok {
					splashE.Splash(tx, otherE.Position())
				}
			}
		}
		if linger {
			tx.AddEntity(NewAreaEffectCloud(world.EntitySpawnOpts{Position: pos}, pot))
		}
	}
}
