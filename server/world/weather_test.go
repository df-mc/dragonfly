package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

type testLivingConfig struct{}

func (testLivingConfig) Apply(*EntityData) {}

type testLivingType struct{}

func (testLivingType) Open(_ *Tx, handle *EntityHandle, data *EntityData) Entity {
	return &testLiving{testEntity{handle: handle, data: data}}
}

func (testLivingType) EncodeEntity() string                  { return "dragonfly:test_living" }
func (testLivingType) BBox(Entity) cube.BBox                 { return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3) }
func (testLivingType) DecodeNBT(map[string]any, *EntityData) {}
func (testLivingType) EncodeNBT(*EntityData) map[string]any  { return nil }

type testLiving struct{ testEntity }

func (h *testLiving) Health() float64 { return 20 }
func (h *testLiving) Tick(*Tx, int64) {}

// TestLightningTargetsByHeight verifies that whether an entity can be struck by lightning depends on how high it
// stands, not on where it is along the Z axis. The highest block of a column is looked up by X and Z, and compared
// against the entity Y.
func TestLightningTargetsByHeight(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	// A tall pillar in the column (x=4, z=40). Nothing is above either entity.
	<-w.exec(func(tx *Tx) {
		stone, ok := tx.World().conf.Blocks.BlockByName("minecraft:stone", nil)
		if !ok {
			t.Fatal("stone not registered")
		}
		tx.SetBlock(cube.Pos{4, 100, 40}, stone, nil)
	})

	// Two identical entities standing in the open air at the same height. The
	// only difference is their Z coordinate.
	posPositiveZ := mgl64.Vec3{4.5, 40, 200.5}
	posNegativeZ := mgl64.Vec3{4.5, 40, 5.5}

	var posHit, negHit bool
	hPos := EntitySpawnOpts{Position: posPositiveZ}.New(testLivingType{}, testLivingConfig{})
	hNeg := EntitySpawnOpts{Position: posNegativeZ}.New(testLivingType{}, testLivingConfig{})
	<-w.exec(func(tx *Tx) {
		tx.AddEntity(hPos)
		got := weather{w: w}.adjustPositionToEntities(tx, mgl64.Vec3{posPositiveZ[0], float64(tx.Range()[0]), posPositiveZ[2]})
		posHit = got.ApproxEqual(posPositiveZ)
		tx.RemoveEntity(mustEntityOf(t, tx, hPos))

		tx.AddEntity(hNeg)
		got = weather{w: w}.adjustPositionToEntities(tx, mgl64.Vec3{posNegativeZ[0], float64(tx.Range()[0]), posNegativeZ[2]})
		negHit = got.ApproxEqual(posNegativeZ)
		tx.RemoveEntity(mustEntityOf(t, tx, hNeg))
	})

	t.Logf("entity at (4.5, 40, 200.5) targeted by lightning: %v", posHit)
	t.Logf("entity at (4.5, 40, 5.5)   targeted by lightning: %v", negHit)
	if posHit != negHit {
		t.Fatalf("lightning eligibility depends on the entity's Z coordinate instead of its height: "+
			"two identical entities at y=40, both in the open air: z=200.5 -> %v, z=5.5 -> %v; "+
			"HighestBlock(pos[0], pos[1]) is called with (x, y) and compared against pos[2]", posHit, negHit)
	}
}

func mustEntityOf(t *testing.T, tx *Tx, h *EntityHandle) Entity {
	t.Helper()
	e, ok := h.Entity(tx)
	if !ok {
		t.Fatal("entity not in world")
	}
	return e
}
