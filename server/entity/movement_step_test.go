package entity

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestStepOnBlockUsesSupportingBlock(t *testing.T) {
	for _, x := range []float64{0.5, 1.1} {
		w := world.Config{Synchronous: true}.New()
		w.Do(func(tx *world.Tx) {
			pos := cube.Pos{0, 64, 0}
			tx.SetBlock(pos, block.RedstoneOre{}, nil)
			h := (world.EntitySpawnOpts{Position: mgl64.Vec3{x, 65, 0.5}}).New(ArrowType, arrowConf)
			e := tx.AddEntity(h)
			StepOnBlock(tx, e, e.Position())
			if !tx.Block(pos).(block.RedstoneOre).Lit {
				t.Errorf("entity at x=%v, y=65 did not light supporting ore at y=64", x)
			}
		}).Wait(context.Background())
		_ = w.Close()
	}
}
