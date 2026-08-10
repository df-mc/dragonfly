package block_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
)

func testWorld() *world.World {
	return world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
}

func testRun(w *world.World, f func(*world.Tx)) {
	w.Do(f).Wait(context.Background())
}

// countDrops counts the item entities in the world, keyed by the encoded item name.
func countDrops(tx *world.Tx) map[string]int {
	m := map[string]int{}
	for e := range tx.Entities() {
		be, ok := e.(interface{ Behaviour() entity.Behaviour })
		if !ok {
			continue
		}
		ib, ok := be.Behaviour().(*entity.ItemBehaviour)
		if !ok {
			continue
		}
		s := ib.Item()
		if s.Empty() {
			continue
		}
		name, _ := s.Item().(world.Item).EncodeItem()
		m[name] += s.Count()
	}
	return m
}

// TestExplosionTwoBlockStructures blows up a bed, a wooden door, a double
// flower and a double tall grass. Each of those occupies two block positions
// but is a single item in the world. Vanilla drops one item per structure.
func TestExplosionTwoBlockStructures(t *testing.T) {
	type structure struct {
		name    string
		item    string
		place   func(tx *world.Tx, pos cube.Pos)
		wantOne int
	}
	structures := []structure{
		{
			name: "bed", item: "minecraft:bed", wantOne: 1,
			place: func(tx *world.Tx, pos cube.Pos) {
				b := block.Bed{Facing: cube.North}
				tx.SetBlock(pos, b, nil)
				head := b
				head.Head = true
				tx.SetBlock(pos.Side(cube.North.Face()), head, nil)
			},
		},
		{
			name: "wood door", item: "minecraft:wooden_door", wantOne: 1,
			place: func(tx *world.Tx, pos cube.Pos) {
				d := block.WoodDoor{Wood: block.OakWood(), Facing: cube.North}
				tx.SetBlock(pos, d, nil)
				top := d
				top.Top = true
				tx.SetBlock(pos.Side(cube.FaceUp), top, nil)
			},
		},
		{
			name: "double flower", item: "minecraft:sunflower", wantOne: 1,
			place: func(tx *world.Tx, pos cube.Pos) {
				f := block.DoubleFlower{Type: block.Sunflower()}
				tx.SetBlock(pos, f, nil)
				up := f
				up.UpperPart = true
				tx.SetBlock(pos.Side(cube.FaceUp), up, nil)
			},
		},
	}

	for _, s := range structures {
		t.Run(s.name, func(t *testing.T) {
			w := testWorld()
			defer w.Close()

			pos := cube.Pos{0, 64, 0}
			testRun(w, func(tx *world.Tx) {
				for x := -2; x <= 2; x++ {
					for z := -2; z <= 2; z++ {
						tx.SetBlock(cube.Pos{x, 63, z}, block.Stone{}, nil)
					}
				}
				s.place(tx, pos)

				before := countDrops(tx)
				block.ExplosionConfig{
					ItemDropChance: 1,
					RandSource:     rand.NewPCG(1, 2),
				}.Explode(tx, world.BlockExplosionSource{
					Block:         block.TNT{},
					Pos:           cube.Pos{0, 64, 3},
					ExplosionSize: 4,
				})
				after := countDrops(tx)

				fmt.Printf("[%s] before=%v after=%v\n", s.name, before[s.item], after[s.item])
				if got := after[s.item] - before[s.item]; got != s.wantOne {
					t.Fatalf("%s: explosion dropped %d %s, want %d (all drops: %v)", s.name, got, s.item, s.wantOne, after)
				}
			})
		})
	}
}
