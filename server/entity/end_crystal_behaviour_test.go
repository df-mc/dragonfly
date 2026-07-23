package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestEndCrystalPlacesFireOnlyInAir(t *testing.T) {
	tests := []struct {
		name     string
		block    world.Block
		support  world.Block
		wantFire bool
	}{
		{name: "unsupported air", block: block.Air{}, support: block.Air{}, wantFire: true},
		{name: "occupied block", block: block.ShortGrass{}, support: block.Obsidian{}, wantFire: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.Config{Dim: world.End, Synchronous: true}.New()
			t.Cleanup(func() { _ = w.Close() })

			pos := cube.Pos{0, 64, 0}
			handle := NewEndCrystal(world.EntitySpawnOpts{Position: mgl64.Vec3{0.5, 64, 0.5}})
			mustDo(t, w, func(tx *world.Tx) {
				tx.SetBlock(pos, tt.block, nil)
				tx.SetBlock(pos.Side(cube.FaceDown), tt.support, nil)
				e := tx.AddEntity(handle).(*Ent)
				e.Tick(tx, 0)
				_, fire := tx.Block(pos).(block.Fire)
				if fire != tt.wantFire {
					t.Fatalf("block at crystal position = %T, want fire %v", tx.Block(pos), tt.wantFire)
				}
			})
		})
	}
}

func TestEndCrystalOnlyProtectsBlocksBelowWithValidSupport(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	t.Cleanup(func() { _ = w.Close() })

	crystalPos := cube.Pos{0, 65, 0}
	tests := []struct {
		name    string
		support world.Block
		want    bool
	}{
		{name: "obsidian", support: block.Obsidian{}, want: true},
		{name: "bedrock", support: block.Bedrock{}, want: true},
		{name: "crying obsidian", support: block.Obsidian{Crying: true}, want: false},
		{name: "air", support: block.Air{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustDo(t, w, func(tx *world.Tx) {
				tx.SetBlock(crystalPos.Side(cube.FaceDown), tt.support, nil)
				if got := endCrystalProtectsBlocksBelow(tx, crystalPos); got != tt.want {
					t.Fatalf("endCrystalProtectsBlocksBelow() = %v, want %v", got, tt.want)
				}
			})
		})
	}
}
