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

// TestBlocksUnder verifies the range of blocks a Player is standing on. Both the block the Player came to rest on and
// the block it steps on are looked up through this: it is the layer below the Player's feet, and it spans every column
// the Player's bounding box covers rather than only the one its centre is over.
func TestBlocksUnder(t *testing.T) {
	tests := []struct {
		name      string
		pos       mgl64.Vec3
		low, high cube.Pos
	}{
		{name: "on top of a block", pos: mgl64.Vec3{0.5, 11, 0.5}, low: cube.Pos{0, 10, 0}, high: cube.Pos{0, 10, 0}},
		{name: "on top of a slab", pos: mgl64.Vec3{0.5, 10.5, 0.5}, low: cube.Pos{0, 10, 0}, high: cube.Pos{0, 10, 0}},
		{name: "over the edge of a block", pos: mgl64.Vec3{-0.25, 11, 0.5}, low: cube.Pos{-1, 10, 0}, high: cube.Pos{0, 10, 0}},
		{name: "over a corner of a block", pos: mgl64.Vec3{1.25, 11, 1.25}, low: cube.Pos{0, 10, 0}, high: cube.Pos{1, 10, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			defer w.Close()

			conf := Config{Name: "test", UUID: uuid.New(), Position: tt.pos, GameMode: world.GameModeSurvival}
			handle := world.EntitySpawnOpts{Position: tt.pos}.New(Type, conf)

			w.Do(func(tx *world.Tx) {
				p := tx.AddEntity(handle).(*Player)
				p.data.Pos = tt.pos

				low, high := p.blocksUnder()
				if low != tt.low || high != tt.high {
					t.Errorf("blocksUnder() at %v = %v..%v, want %v..%v", tt.pos, low, high, tt.low, tt.high)
				}
			})
		})
	}
}

// TestFallOnBlockEdge verifies that a block landed on is found anywhere under the Player rather than only below its
// centre. The Player is 0.6 blocks wide, so it may come to rest on the edge of a block with its centre over the block
// beside it, which used to hide blocks such as slime from the fall damage that landing on them cancels.
func TestFallOnBlockEdge(t *testing.T) {
	// The slime block occupies x and z in [0, 1] and y in [10, 11], so the Player stands at y 11.
	slime := cube.Pos{0, 10, 0}

	tests := []struct {
		name string
		// x is the horizontal centre of the Player. The Player is supported by the slime block for any centre within
		// 0.3 of it, which reaches past the edges of the block itself.
		x        float64
		wantHurt bool
	}{
		{name: "centre over the block", x: 0.5},
		{name: "centre over the edge of the block", x: 0.05},
		{name: "centre past the edge of the block", x: -0.25},
		{name: "centre past the far edge of the block", x: 1.25},
		{name: "beside the block", x: 3, wantHurt: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			defer w.Close()

			pos := mgl64.Vec3{tt.x, 11, 0.5}
			conf := Config{Name: "test", UUID: uuid.New(), Position: pos, GameMode: world.GameModeSurvival}
			handle := world.EntitySpawnOpts{Position: pos}.New(Type, conf)

			w.Do(func(tx *world.Tx) {
				tx.SetBlock(slime, block.Slime{}, nil)
				p := tx.AddEntity(handle).(*Player)
				p.data.Pos = pos

				before := p.Health()
				p.fall(20)

				if hurt := p.Health() < before; hurt != tt.wantHurt {
					t.Errorf("fall(20) at x %v hurt = %v, want %v: the slime block was not found", tt.x, hurt, tt.wantHurt)
				}
			})
		})
	}
}
