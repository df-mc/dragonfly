package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type tntSourceEntity struct {
	h     *world.EntityHandle
	owner *world.EntityHandle
}

func (e tntSourceEntity) Close() error           { return nil }
func (e tntSourceEntity) H() *world.EntityHandle { return e.h }
func (e tntSourceEntity) Position() mgl64.Vec3   { return mgl64.Vec3{} }
func (e tntSourceEntity) Rotation() cube.Rotation {
	return cube.Rotation{}
}
func (e tntSourceEntity) ProjectileOwner() *world.EntityHandle { return e.owner }

type tntTestEntityType struct{}

func (tntTestEntityType) Open(*world.Tx, *world.EntityHandle, *world.EntityData) world.Entity {
	return nil
}
func (tntTestEntityType) EncodeEntity() string                        { return "dragonfly:test_entity" }
func (tntTestEntityType) BBox(world.Entity) cube.BBox                 { return cube.Box(0, 0, 0, 0, 0, 0) }
func (tntTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (tntTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
func (tntTestEntityType) Apply(data *world.EntityData)                {}
func newTNTTestHandle() *world.EntityHandle {
	return world.EntitySpawnOpts{}.New(tntTestEntityType{}, tntTestEntityType{})
}

func TestTNTIgnitionSourcePrefersProjectileOwner(t *testing.T) {
	owner := newTNTTestHandle()
	projectile := tntSourceEntity{h: newTNTTestHandle(), owner: owner}

	if got := tntIgnitionSourceHandle(projectile); got != owner {
		t.Fatalf("expected TNT ignition source to use projectile owner handle %v, got %v", owner, got)
	}
}

func TestTNTIgnitionSourceFallsBackToIgnitingEntity(t *testing.T) {
	projectile := tntSourceEntity{h: newTNTTestHandle()}

	if got := tntIgnitionSourceHandle(projectile); got != projectile.H() {
		t.Fatalf("expected TNT ignition source to fall back to igniting entity handle %v, got %v", projectile.H(), got)
	}
}

func TestTNTExplosionSourceUsesExplosionConfigSource(t *testing.T) {
	source := tntSourceEntity{h: newTNTTestHandle()}

	if got := tntExplosionSourceHandle(ExplosionConfig{Source: source}); got != source.H() {
		t.Fatalf("expected chained TNT source to use explosion config source handle %v, got %v", source.H(), got)
	}
}
