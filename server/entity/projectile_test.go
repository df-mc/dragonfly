package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestProjectileActivatesPressurePlateOnTopFaceCollision(t *testing.T) {
	for _, test := range []struct {
		name      string
		plateType block.PressurePlateType
		want      int
	}{
		{name: "wooden", plateType: block.OakPressurePlate(), want: 15},
		{name: "weighted", plateType: block.LightWeightedPressurePlate(), want: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: DefaultRegistry}.New()
			defer w.Close()

			platePos := cube.Pos{0, 64, 0}
			mustDo(t, w, func(tx *world.Tx) {
				tx.SetBlock(platePos.Side(cube.FaceDown), block.Stone{}, nil)
				tx.SetBlock(platePos, block.PressurePlate{Type: test.plateType}, nil)

				conf := arrowConf
				handle := world.EntitySpawnOpts{
					Position: mgl64.Vec3{0.5, 65, 0.5},
					Velocity: mgl64.Vec3{0, -1, 0},
				}.New(ArrowType, conf)
				arrow := tx.AddEntity(handle).(*Ent)
				arrow.Tick(tx, 0)

				plate := tx.Block(platePos).(block.PressurePlate)
				if plate.Power != test.want {
					t.Fatalf("pressure plate power after arrow collision = %d, want %d", plate.Power, test.want)
				}
			})
		})
	}
}
