package world

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
)

type entityRegistryTestType struct{}

func (entityRegistryTestType) Open(*Tx, *EntityHandle, *EntityData) Entity { return nil }
func (entityRegistryTestType) EncodeEntity() string                        { return "dragonfly:entity_registry_test" }
func (entityRegistryTestType) BBox(Entity) cube.BBox                       { return cube.Box(0, 0, 0, 0, 0, 0) }
func (entityRegistryTestType) DecodeNBT(map[string]any, *EntityData)       {}
func (entityRegistryTestType) EncodeNBT(*EntityData) map[string]any        { return nil }
func (entityRegistryTestType) Apply(*EntityData)                           {}

func TestEntityRegistryConfigTNTWithSourceFallbackAllowsDefaultTNT(t *testing.T) {
	called := false
	reg := EntityRegistryConfig{
		TNT: func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle {
			called = true
			return opts.New(entityRegistryTestType{}, entityRegistryTestType{})
		},
	}.New([]EntityType{entityRegistryTestType{}})

	if h := reg.Config().TNTWithSource(EntitySpawnOpts{}, time.Second, nil, true); h == nil {
		t.Fatal("expected fallback TNTWithSource to create TNT through TNT")
	}
	if !called {
		t.Fatal("expected fallback TNTWithSource to call TNT")
	}
}

func TestEntityRegistryConfigTNTWithSourceFallbackAllowsSourceAwareTNT(t *testing.T) {
	called := false
	reg := EntityRegistryConfig{
		TNT: func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle {
			called = true
			return opts.New(entityRegistryTestType{}, entityRegistryTestType{})
		},
	}.New([]EntityType{entityRegistryTestType{}})

	if h := reg.Config().TNTWithSource(EntitySpawnOpts{}, time.Second, EntitySpawnOpts{}.New(entityRegistryTestType{}, entityRegistryTestType{}), true); h == nil {
		t.Fatal("expected fallback TNTWithSource to create TNT through TNT")
	}
	if !called {
		t.Fatal("expected fallback TNTWithSource to call TNT")
	}
}

func TestEntityRegistryConfigTNTWithSourceFallbackAllowsShieldUnblockableTNT(t *testing.T) {
	called := false
	reg := EntityRegistryConfig{
		TNT: func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle {
			called = true
			return opts.New(entityRegistryTestType{}, entityRegistryTestType{})
		},
	}.New([]EntityType{entityRegistryTestType{}})

	if h := reg.Config().TNTWithSource(EntitySpawnOpts{}, time.Second, nil, false); h == nil {
		t.Fatal("expected fallback TNTWithSource to create TNT through TNT")
	}
	if !called {
		t.Fatal("expected fallback TNTWithSource to call TNT")
	}
}
