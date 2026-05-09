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

func TestTNTExplosionConfigHonoursBlockabilityInput(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	<-w.Exec(func(tx *world.Tx) {
		conf := tntExplosionConfig(tx, nil, false)
		if !conf.UnblockableByShield {
			t.Fatal("expected TNT configured as shield-unblockable to remain unblockable")
		}
	})
}

func TestTNTExplosionConfigDefaultsToShieldBlockable(t *testing.T) {
	w := world.New()
	defer func() {
		_ = w.Close()
	}()

	<-w.Exec(func(tx *world.Tx) {
		conf := tntExplosionConfig(tx, nil, true)
		if conf.UnblockableByShield {
			t.Fatal("expected default TNT explosions to be shield blockable")
		}
	})
}

func TestTNTNBTPreservesUnblockableShieldConfig(t *testing.T) {
	var data world.EntityData
	TNTType.DecodeNBT(map[string]any{
		"Fuse":                         uint8(5),
		"DragonflyUnblockableByShield": uint8(1),
	}, &data)

	encoded := TNTType.EncodeNBT(&data)
	if encoded["DragonflyUnblockableByShield"] != uint8(1) {
		t.Fatalf("expected saved TNT shield blockability to survive NBT round trip, got %#v", encoded["DragonflyUnblockableByShield"])
	}
}

func TestTNTNBTDefaultsToShieldBlockable(t *testing.T) {
	var data world.EntityData
	TNTType.DecodeNBT(map[string]any{"Fuse": uint8(5)}, &data)

	encoded := TNTType.EncodeNBT(&data)
	if _, ok := encoded["DragonflyUnblockableByShield"]; ok {
		t.Fatal("expected default decoded TNT to stay shield blockable")
	}
}

func TestExplosionDamageSourceFromNilConfigIsBlockable(t *testing.T) {
	src := ExplosionDamageSourceFromConfig(cube.Pos{}.Vec3Centre(), nil)

	if !src.HasOrigin {
		t.Fatal("expected nil-config explosion damage source to keep origin")
	}
	if !src.BlockableByShield {
		t.Fatal("expected nil-config explosion damage source to default to shield blockable")
	}
	if src.Source != nil {
		t.Fatalf("expected nil-config explosion damage source not to have a source, got %T", src.Source)
	}
}
