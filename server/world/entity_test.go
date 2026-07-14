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

func TestHurtResult(t *testing.T) {
	tests := []struct {
		name      string
		result    HurtResult
		damaged   bool
		blocked   bool
		cancelled bool
	}{
		{name: "zero value is immune", result: HurtImmune},
		{name: "damaged", result: HurtDamaged, damaged: true},
		{name: "blocked", result: HurtBlocked, blocked: true},
		{name: "cancelled", result: HurtCancelled, cancelled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Damaged(); got != tt.damaged {
				t.Errorf("Damaged() = %v, want %v", got, tt.damaged)
			}
			if got := tt.result.Blocked(); got != tt.blocked {
				t.Errorf("Blocked() = %v, want %v", got, tt.blocked)
			}
			if got := tt.result.Cancelled(); got != tt.cancelled {
				t.Errorf("Cancelled() = %v, want %v", got, tt.cancelled)
			}
		})
	}
}

func TestEntityRegistryConfigTNTWithSourceFallsBackToTNT(t *testing.T) {
	called := false
	reg := EntityRegistryConfig{
		TNT: func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle {
			called = true
			return opts.New(entityRegistryTestType{}, entityRegistryTestType{})
		},
	}.New([]EntityType{entityRegistryTestType{}})

	source := EntitySpawnOpts{}.New(entityRegistryTestType{}, entityRegistryTestType{})
	if h := reg.Config().TNTWithSource(EntitySpawnOpts{}, time.Second, source, false); h == nil {
		t.Fatal("expected fallback TNTWithSource to create TNT through TNT")
	}
	if !called {
		t.Fatal("expected fallback TNTWithSource to call TNT")
	}
}
