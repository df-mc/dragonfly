package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestExplosionExposureAccountsForObstructions(t *testing.T) {
	entityType := explosionExposureTestEntityType{}
	registry := world.EntityRegistryConfig{}.New([]world.EntityType{entityType})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer w.Close()

	h := world.EntitySpawnOpts{Position: mgl64.Vec3{2, 0, 0}}.New(entityType, explosionExposureTestEntityConfig{})
	origin := mgl64.Vec3{0, 1, 0}
	w.Do(func(tx *world.Tx) {
		target := tx.AddEntity(h)
		if got := ExplosionExposure(tx, origin, target); got != 1 {
			t.Fatalf("unobstructed exposure = %v, want 1", got)
		}

		for y := range 2 {
			for z := -1; z <= 0; z++ {
				tx.SetBlock(cube.Pos{1, y, z}, Stone{}, nil)
			}
		}
		if got := ExplosionExposure(tx, origin, target); got != 0 {
			t.Fatalf("obstructed exposure = %v, want 0", got)
		}
	})
}

type explosionExposureTestEntityConfig struct{}

func (explosionExposureTestEntityConfig) Apply(*world.EntityData) {}

type explosionExposureTestEntityType struct{}

func (explosionExposureTestEntityType) Open(_ *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return explosionExposureTestEntity{handle: handle, data: data}
}
func (explosionExposureTestEntityType) EncodeEntity() string { return "test:explosion_exposure" }
func (explosionExposureTestEntityType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3)
}
func (explosionExposureTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (explosionExposureTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type explosionExposureTestEntity struct {
	handle *world.EntityHandle
	data   *world.EntityData
}

func (e explosionExposureTestEntity) Close() error            { return nil }
func (e explosionExposureTestEntity) H() *world.EntityHandle  { return e.handle }
func (e explosionExposureTestEntity) Position() mgl64.Vec3    { return e.data.Pos }
func (e explosionExposureTestEntity) Rotation() cube.Rotation { return e.data.Rot }
