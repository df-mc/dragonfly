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
	var source *world.EntityHandle
	registry := world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
		TNTWithSource: func(opts world.EntitySpawnOpts, fuse time.Duration, src *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
			source, blockable = src, blockableByShield
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
	}.New([]world.EntityType{tntTestEntityType{}})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer func() {
		_ = w.Close()
	}()

	w.Do(func(tx *world.Tx) {
		spawnTnt(cube.Pos{}, tx, time.Second, nil, true)
	})
	if source != nil {
		t.Fatalf("expected no TNT source entity, got %T", source)
	}
	if !blockable {
		t.Fatal("expected source-less environmental TNT to be shield blockable")
	}
}

func TestTNTIgniteWithoutSourceIsShieldBlockable(t *testing.T) {
	var blockable bool
	var source *world.EntityHandle
	registry := world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
		TNTWithSource: func(opts world.EntitySpawnOpts, fuse time.Duration, src *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
			source, blockable = src, blockableByShield
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
	}.New([]world.EntityType{tntTestEntityType{}})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer func() {
		_ = w.Close()
	}()

	w.Do(func(tx *world.Tx) {
		TNT{}.Ignite(cube.Pos{}, tx, nil)
	})
	if source != nil {
		t.Fatalf("expected no TNT source entity, got %T", source)
	}
	if !blockable {
		t.Fatal("expected source-less TNT ignition to be shield-blockable")
	}
}

func TestTNTIgniteWithSourceIsShieldBlockable(t *testing.T) {
	source := tntSourceEntity{h: newTNTTestHandle()}
	var blockable bool
	var gotSource *world.EntityHandle
	registry := world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
		TNTWithSource: func(opts world.EntitySpawnOpts, fuse time.Duration, src *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
			gotSource, blockable = src, blockableByShield
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
	}.New([]world.EntityType{tntTestEntityType{}})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer func() {
		_ = w.Close()
	}()

	w.Do(func(tx *world.Tx) {
		TNT{}.Ignite(cube.Pos{}, tx, source)
	})
	if gotSource != source.H() {
		t.Fatalf("expected TNT source entity %v, got %v", source.H(), gotSource)
	}
	if !blockable {
		t.Fatal("expected source-aware TNT ignition to stay shield-blockable for other players")
	}
}

func TestTNTSpawnPreservesUnavailableSourceHandle(t *testing.T) {
	wantSource := newTNTTestHandle()
	var gotSource *world.EntityHandle
	registry := world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
		TNTWithSource: func(opts world.EntitySpawnOpts, fuse time.Duration, src *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
			gotSource = src
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
	}.New([]world.EntityType{tntTestEntityType{}})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer func() {
		_ = w.Close()
	}()

	w.Do(func(tx *world.Tx) {
		spawnTnt(cube.Pos{}, tx, time.Second, wantSource, true)
	})
	if gotSource != wantSource {
		t.Fatalf("expected unavailable source handle %v to be passed through, got %v", wantSource, gotSource)
	}
}

func TestTNTSpawnPreservesSourceLessUnblockableExplosion(t *testing.T) {
	var blockable bool
	registry := world.EntityRegistryConfig{
		TNT: func(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
		TNTWithSource: func(opts world.EntitySpawnOpts, fuse time.Duration, src *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
			blockable = blockableByShield
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		},
	}.New([]world.EntityType{tntTestEntityType{}})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer func() {
		_ = w.Close()
	}()

	w.Do(func(tx *world.Tx) {
		TNT{}.Explode(cube.Pos{}.Vec3Centre(), cube.Pos{}, tx, ExplosionConfig{UnblockableByShield: true})
	})
	if blockable {
		t.Fatal("expected source-less unblockable explosion to prime shield-unblockable TNT")
	}
}
