package portal_test

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
)

func do(t *testing.T, w *world.World, f func(*world.Tx)) {
	t.Helper()
	if err := w.Do(f).Wait(context.Background()); err != nil {
		t.Fatalf("world task: %v", err)
	}
}

func TestNetherPortalFromPos(t *testing.T) {
	tests := []struct {
		name  string
		build func(tx *world.Tx, origin cube.Pos)
		pos   cube.Pos
		ok    bool
	}{
		{
			name: "valid vertical frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 2, 3)
			},
			pos: cube.Pos{},
			ok:  true,
		},
		{
			name: "valid X axis frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.X, 2, 3)
			},
			pos: cube.Pos{},
			ok:  true,
		},
		{
			name: "maximum size frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 21, 21)
			},
			pos: cube.Pos{},
			ok:  true,
		},
		{
			name: "too wide frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 22, 3)
			},
			pos: cube.Pos{},
		},
		{
			name: "too tall frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 2, 22)
			},
			pos: cube.Pos{},
		},
		{
			name: "horizontal frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildHorizontalFrame(tx, origin)
			},
			pos: cube.Pos{1, 0, 1},
		},
		{
			name: "crying obsidian does not complete frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 2, 3)
				tx.SetBlock(origin.Side(cube.FaceNorth), block.Obsidian{Crying: true}, nil)
			},
			pos: cube.Pos{},
		},
		{
			name: "missing side frame does not complete frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 2, 3)
				tx.SetBlock(origin.Side(cube.FaceNorth), nil, nil)
			},
			pos: cube.Pos{},
		},
		{
			name: "soul fire does not activate frame",
			build: func(tx *world.Tx, origin cube.Pos) {
				buildVerticalFrame(tx, origin, cube.Z, 2, 3)
				tx.SetBlock(origin, block.Fire{Type: block.SoulFire()}, nil)
			},
			pos: cube.Pos{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := world.New()
			t.Cleanup(func() { _ = w.Close() })

			origin := cube.Pos{8, 10, 8}
			do(t, w, func(tx *world.Tx) {
				tt.build(tx, origin)
				p, ok := portal.NetherPortalFromPos(tx, origin.Add(tt.pos))
				if ok != tt.ok {
					t.Fatalf("NetherPortalFromPos() ok = %v, want %v, portal = %#v", ok, tt.ok, p)
				}
				if ok && !p.Framed() {
					t.Fatal("NetherPortalFromPos() returned an unframed portal")
				}
			})
		})
	}
}

func TestPortalModelHasNoCollisionBBox(t *testing.T) {
	for _, axis := range []cube.Axis{cube.X, cube.Z} {
		if boxes := (model.Portal{Axis: axis}).BBox(cube.Pos{}, nil); len(boxes) != 0 {
			t.Fatalf("BBox() returned %d boxes, want 0", len(boxes))
		}
	}
}

func TestActivateNetherPortal(t *testing.T) {
	for _, axis := range []cube.Axis{cube.Z, cube.X} {
		t.Run(axis.String(), func(t *testing.T) {
			w := world.New()
			t.Cleanup(func() { _ = w.Close() })

			origin := cube.Pos{8, 10, 8}
			do(t, w, func(tx *world.Tx) {
				buildVerticalFrame(tx, origin, axis, 2, 3)
				if !portal.ActivateNetherPortal(tx, origin) {
					t.Fatal("ActivateNetherPortal() = false, want true")
				}
				for x := range 2 {
					for y := range 3 {
						pos := origin.Add(widthOffset(axis, x)).Add(cube.Pos{0, y})
						pb, ok := tx.Block(pos).(block.Portal)
						if !ok {
							t.Fatalf("portal block not placed at interior offset %d,%d", x, y)
						}
						if pb.Axis != axis {
							t.Fatalf("portal block at interior offset %d,%d has axis %v, want %v", x, y, pb.Axis, axis)
						}
					}
				}
			})
		})
	}
}

func TestFireChargeActivatesNetherPortal(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	origin := cube.Pos{8, 10, 8}
	do(t, w, func(tx *world.Tx) {
		buildVerticalFrame(tx, origin, cube.Z, 2, 3)
		ctx := &item.UseContext{}
		if ok := (item.FireCharge{}).UseOnBlock(origin.Side(cube.FaceDown), cube.FaceUp, cube.Pos{}.Vec3(), tx, nil, ctx); !ok {
			t.Fatal("FireCharge.UseOnBlock() = false, want true")
		}
		if ctx.CountSub != 1 {
			t.Fatalf("FireCharge.UseOnBlock() subtracted %d items, want 1", ctx.CountSub)
		}
		if _, ok := tx.Block(origin).(block.Portal); !ok {
			t.Fatal("FireCharge.UseOnBlock() did not activate portal")
		}
	})
}

func TestActivatedPortalCleanupOnBrokenFrame(t *testing.T) {
	w := world.New()
	t.Cleanup(func() { _ = w.Close() })

	origin := cube.Pos{8, 10, 8}
	do(t, w, func(tx *world.Tx) {
		buildVerticalFrame(tx, origin, cube.Z, 2, 3)
		if !portal.ActivateNetherPortal(tx, origin) {
			t.Fatal("ActivateNetherPortal() = false, want true")
		}

		broken := origin.Add(widthOffset(cube.Z, 2)).Add(cube.Pos{0, 1})
		tx.SetBlock(broken, nil, nil)

		updated := origin.Add(widthOffset(cube.Z, 1)).Add(cube.Pos{0, 1})
		pb, ok := tx.Block(updated).(block.Portal)
		if !ok {
			t.Fatalf("block at updated position = %T, want block.Portal", tx.Block(updated))
		}
		pb.NeighbourUpdateTick(updated, broken, tx)

		var remaining []cube.Pos
		for x := range 2 {
			for y := range 3 {
				p := origin.Add(widthOffset(cube.Z, x)).Add(cube.Pos{0, y})
				if _, ok := tx.Block(p).(block.Portal); ok {
					remaining = append(remaining, p)
				}
			}
		}
		if len(remaining) != 0 {
			t.Fatalf("after frame break: %d orphan portal blocks remain at %v", len(remaining), remaining)
		}
	})
}

func buildVerticalFrame(tx *world.Tx, origin cube.Pos, axis cube.Axis, width, height int) {
	for x := 0; x < width; x++ {
		p := origin.Add(widthOffset(axis, x))
		tx.SetBlock(p.Side(cube.FaceDown), block.Obsidian{}, nil)
		tx.SetBlock(p.Add(cube.Pos{0, height}), block.Obsidian{}, nil)
	}
	negative := cube.FaceNorth
	if axis == cube.X {
		negative = cube.FaceWest
	}
	for y := 0; y < height; y++ {
		p := origin.Add(cube.Pos{0, y})
		tx.SetBlock(p.Side(negative), block.Obsidian{}, nil)
		tx.SetBlock(p.Add(widthOffset(axis, width)), block.Obsidian{}, nil)
	}
}

func buildHorizontalFrame(tx *world.Tx, origin cube.Pos) {
	for x := 0; x < 3; x++ {
		for z := 0; z < 3; z++ {
			if x == 1 && z == 1 {
				continue
			}
			tx.SetBlock(origin.Add(cube.Pos{x, 0, z}), block.Obsidian{}, nil)
		}
	}
}

func widthOffset(axis cube.Axis, width int) cube.Pos {
	if axis == cube.X {
		return cube.Pos{width, 0, 0}
	}
	return cube.Pos{0, 0, width}
}
