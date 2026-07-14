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

func TestTNTIgnitionSource(t *testing.T) {
	owner := newTNTTestHandle()
	projectile := tntSourceEntity{h: newTNTTestHandle(), owner: owner}
	tests := []struct {
		name   string
		source tntSourceEntity
		want   *world.EntityHandle
	}{
		{name: "projectile owner", source: projectile, want: owner},
		{name: "igniting entity", source: tntSourceEntity{h: newTNTTestHandle()}, want: nil},
	}
	tests[1].want = tests[1].source.H()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tntIgnitionSourceHandle(tt.source); got != tt.want {
				t.Fatalf("tntIgnitionSourceHandle() = %v, want %v", got, tt.want)
			}
		})
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
