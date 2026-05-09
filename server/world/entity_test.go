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

func TestEntityRegistryConfigTNTWithSourceFallbackRejectsSourceAwareTNT(t *testing.T) {
	reg := EntityRegistryConfig{
		TNT: func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle {
			return opts.New(entityRegistryTestType{}, entityRegistryTestType{})
		},
	}.New([]EntityType{entityRegistryTestType{}})

	defer func() {
		if recover() == nil {
			t.Fatal("expected fallback TNTWithSource to reject source-aware TNT")
		}
	}()
	reg.Config().TNTWithSource(EntitySpawnOpts{}, time.Second, EntitySpawnOpts{}.New(entityRegistryTestType{}, entityRegistryTestType{}), true)
}

func TestEntityRegistryConfigTNTWithSourceFallbackRejectsShieldUnblockableTNT(t *testing.T) {
	reg := EntityRegistryConfig{
		TNT: func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle {
			return opts.New(entityRegistryTestType{}, entityRegistryTestType{})
		},
	}.New([]EntityType{entityRegistryTestType{}})

	defer func() {
		if recover() == nil {
			t.Fatal("expected fallback TNTWithSource to reject shield-unblockable TNT")
		}
	}()
	reg.Config().TNTWithSource(EntitySpawnOpts{}, time.Second, nil, false)
}
