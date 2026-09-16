package entity

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TestButtonProjectileOverlap checks activation without a new collision with the support.
func TestButtonProjectileOverlap(t *testing.T) {
	for _, name := range []string{"flying", "attached"} {
		t.Run(name, func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()
			w.Do(func(tx *world.Tx) {
				pos := cube.Pos{0, 64, 0}
				support := pos.Side(cube.FaceNorth)
				tx.SetBlock(support, block.Stone{}, nil)
				opts := world.EntitySpawnOpts{Position: mgl64.Vec3{-0.25, 64.45, 0.06}, Velocity: mgl64.Vec3{0.8, 0, 0}}
				conf := arrowConf
				if name == "attached" {
					opts.Position, opts.Velocity = mgl64.Vec3{0.5, 64.45, 0.01}, mgl64.Vec3{}
					conf.CollisionPosition = support
				}
				e := tx.AddEntity(opts.New(ArrowType, conf)).(*Ent)
				tx.SetBlock(pos, block.Button{Type: block.OakButton(), Facing: cube.FaceSouth}, nil)
				e.Tick(tx, 1)
				if !tx.Block(pos).(block.Button).Pressed {
					t.Fatal("overlapping arrow did not press wooden button")
				}
			}).Wait(context.Background())
		})
	}
}
