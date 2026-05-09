package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type tntTestEntityType struct{}

func (tntTestEntityType) Open(*world.Tx, *world.EntityHandle, *world.EntityData) world.Entity {
	return nil
}
func (tntTestEntityType) EncodeEntity() string                        { return "dragonfly:test_entity" }
func (tntTestEntityType) BBox(world.Entity) cube.BBox                 { return cube.Box(0, 0, 0, 0, 0, 0) }
func (tntTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (tntTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
func (tntTestEntityType) Apply(*world.EntityData)                     {}

func TestTNTExplosionWithUnavailableSourceRemainsShieldBlockable(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()
	source := world.EntitySpawnOpts{}.New(tntTestEntityType{}, tntTestEntityType{})

	<-w.Exec(func(tx *world.Tx) {
		conf := tntExplosionConfig(tx, source, true)
		if conf.UnblockableByShield {
			t.Fatal("expected source-ignited TNT to remain shield blockable even if its source entity is unavailable")
		}
		if conf.Source != nil {
			t.Fatal("expected unavailable source entity not to be attached to the explosion config")
		}
	})
}

func TestTNTExplosionWithoutSourceIsUnblockableByShield(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	<-w.Exec(func(tx *world.Tx) {
		conf := tntExplosionConfig(tx, nil, false)
		if !conf.UnblockableByShield {
			t.Fatal("expected source-less TNT to be unblockable by shields")
		}
	})
}
