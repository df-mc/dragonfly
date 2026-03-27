package item_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

func TestEyeOfEnderUseThrowsSignal(t *testing.T) {
	finalisePortalItemBlocksOnce.Do(worldFinaliseBlockRegistry)

	w := world.Config{
		Dim:      world.Overworld,
		Provider: world.NopProvider{},
		Entities: entity.DefaultRegistry,
	}.New()
	defer closePortalItemWorld(t, w)

	var (
		used         bool
		ctx          item.UseContext
		spawnedTypes []string
	)
	<-w.Exec(func(tx *world.Tx) {
		user := tx.AddEntity(world.NewEntity(player.Type, player.Config{
			Name: "eye-thrower",
		})).(*player.Player)

		used = (item.EyeOfEnder{}).Use(tx, user, &ctx)
		for e := range tx.Entities() {
			if e == user {
				continue
			}
			spawnedTypes = append(spawnedTypes, e.H().Type().EncodeEntity())
		}
	})

	if !used {
		t.Fatal("expected eye of ender use to throw a signal entity")
	}
	if ctx.CountSub != 1 {
		t.Fatalf("expected eye of ender use to subtract one item, got %d", ctx.CountSub)
	}
	var eyeSignals int
	for _, typ := range spawnedTypes {
		if typ == "minecraft:eye_of_ender_signal" {
			eyeSignals++
		}
	}
	if eyeSignals != 1 {
		t.Fatalf("expected exactly one thrown eye signal entity, got %d (%v)", eyeSignals, spawnedTypes)
	}
}
