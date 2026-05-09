package block

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type tntSourceEntity struct {
	h     *world.EntityHandle
	owner *world.EntityHandle
	pos   mgl64.Vec3
}

func (e tntSourceEntity) Close() error           { return nil }
func (e tntSourceEntity) H() *world.EntityHandle { return e.h }
func (e tntSourceEntity) Position() mgl64.Vec3   { return e.pos }
func (e tntSourceEntity) Rotation() cube.Rotation {
	return cube.Rotation{}
}
func (e tntSourceEntity) ProjectileOwner() *world.EntityHandle { return e.owner }

type tntTestEntityType struct{}

func (tntTestEntityType) Open(_ *world.Tx, h *world.EntityHandle, data *world.EntityData) world.Entity {
	return tntSourceEntity{h: h}.withPosition(data.Pos)
}
func (tntTestEntityType) EncodeEntity() string                        { return "dragonfly:test_entity" }
func (tntTestEntityType) BBox(world.Entity) cube.BBox                 { return cube.Box(0, 0, 0, 0, 0, 0) }
func (tntTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (tntTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
func (tntTestEntityType) Apply(data *world.EntityData)                {}

func (e tntSourceEntity) withPosition(pos mgl64.Vec3) tntSourceEntity {
	e.pos = pos
	return e
}
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

func TestTNTSpawnCanCreateShieldBlockableTNTWithoutSource(t *testing.T) {
	var blockable bool
	var source world.Entity
	registry := world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
		TNTWithSource: func(opts world.EntitySpawnOpts, fuse time.Duration, src world.Entity, blockableByShield bool) *world.EntityHandle {
			source, blockable = src, blockableByShield
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
	}.New([]world.EntityType{tntTestEntityType{}})
	w := world.Config{Entities: registry}.New()
	defer func() {
		_ = w.Close()
	}()

	<-w.Exec(func(tx *world.Tx) {
		spawnTnt(cube.Pos{}, tx, time.Second, nil, true)
	})
	if source != nil {
		t.Fatalf("expected no TNT source entity, got %T", source)
	}
	if !blockable {
		t.Fatal("expected source-less environmental TNT to be shield blockable")
	}
}
