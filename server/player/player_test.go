package player

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

// TestCheckOnGround verifies that a Player is only considered to be on the ground when a block is below it. The box
// checked is extended back over the movement of the tick so that ground moved past within one tick is not missed, but
// extending it back over a fall grows it upwards, which used to make a block moved past sideways count as ground.
func TestCheckOnGround(t *testing.T) {
	// The platform occupies x and z in [0, 1] and y in [10, 11].
	platform := cube.Pos{0, 10, 0}

	tests := []struct {
		name     string
		pos      mgl64.Vec3
		deltaPos mgl64.Vec3
		want     bool
	}{
		{
			// Falling alongside the platform and moving into its column on the tick the player passes its bottom
			// face. The player is a full block below the platform and touches nothing.
			name:     "falling past the bottom edge of an overhang",
			pos:      mgl64.Vec3{-0.1, 7.7, 0.5},
			deltaPos: mgl64.Vec3{0.2, -1, 0},
			want:     false,
		},
		{
			name:     "falling straight down alongside the platform",
			pos:      mgl64.Vec3{-0.5, 7.7, 0.5},
			deltaPos: mgl64.Vec3{0, -1, 0},
			want:     false,
		},
		{
			// Standing on top of the platform.
			name:     "standing on a block",
			pos:      mgl64.Vec3{0.5, 11, 0.5},
			deltaPos: mgl64.Vec3{},
			want:     true,
		},
		{
			// Moving horizontally off the top of the platform within a single tick, as when running down stairs. The
			// backwards sweep over the horizontal movement must still find the block that was left.
			name:     "moving off a block within one tick",
			pos:      mgl64.Vec3{1.6, 11, 0.5},
			deltaPos: mgl64.Vec3{1, 0, 0},
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			defer w.Close()

			conf := Config{Name: "test", UUID: uuid.New(), Position: tt.pos, GameMode: world.GameModeSurvival}
			handle := world.EntitySpawnOpts{Position: conf.Position}.New(Type, conf)

			w.Do(func(tx *world.Tx) {
				tx.SetBlock(platform, block.Stone{}, nil)
				p := tx.AddEntity(handle).(*Player)
				p.data.Pos = tt.pos

				if got := p.checkOnGround(tt.deltaPos); got != tt.want {
					t.Errorf("checkOnGround(%v) at %v = %v, want %v", tt.deltaPos, tt.pos, got, tt.want)
				}
			})
		})
	}
}
