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

func TestTNTShieldBlockability(t *testing.T) {
	tests := []struct {
		name          string
		spawn         func(TNT, *world.Tx, tntSourceEntity)
		wantSource    bool
		wantBlockable bool
	}{
		{name: "entity ignition", spawn: func(b TNT, tx *world.Tx, source tntSourceEntity) {
			b.Ignite(cube.Pos{}, tx, source)
		}, wantSource: true, wantBlockable: true},
		{name: "unblockable chain explosion", spawn: func(b TNT, tx *world.Tx, _ tntSourceEntity) {
			b.Explode(cube.Pos{}.Vec3Centre(), cube.Pos{}, tx, ExplosionConfig{UnblockableByShield: true})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tntSourceEntity{h: newTNTTestHandle()}
			var gotSource *world.EntityHandle
			var gotBlockable bool
			registry := world.EntityRegistryConfig{
				TNT: func(opts world.EntitySpawnOpts, _ time.Duration) *world.EntityHandle {
					return opts.New(tntTestEntityType{}, tntTestEntityType{})
				},
				TNTWithSource: func(opts world.EntitySpawnOpts, _ time.Duration, src *world.EntityHandle, blockable bool) *world.EntityHandle {
					gotSource, gotBlockable = src, blockable
					return opts.New(tntTestEntityType{}, tntTestEntityType{})
				},
			}.New([]world.EntityType{tntTestEntityType{}})
			w := world.Config{Synchronous: true, Entities: registry}.New()
			defer w.Close()
			w.Do(func(tx *world.Tx) { tt.spawn(TNT{}, tx, source) })

			if (gotSource == source.H()) != tt.wantSource {
				t.Fatalf("source match = %v, want %v", gotSource == source.H(), tt.wantSource)
			}
			if gotBlockable != tt.wantBlockable {
				t.Fatalf("blockable = %v, want %v", gotBlockable, tt.wantBlockable)
			}
		})
	}

	t.Run("source-aware registry fallback", func(t *testing.T) {
		called := false
		registry := world.EntityRegistryConfig{TNT: func(opts world.EntitySpawnOpts, _ time.Duration) *world.EntityHandle {
			called = true
			return opts.New(tntTestEntityType{}, tntTestEntityType{})
		}}.New([]world.EntityType{tntTestEntityType{}})
		if h := registry.Config().TNTWithSource(world.EntitySpawnOpts{}, time.Second, newTNTTestHandle(), false); h == nil || !called {
			t.Fatal("TNTWithSource did not fall back to TNT")
		}
	})
}
