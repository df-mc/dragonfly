package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

func TestOwnedProjectileConstructorsRejectNilOwner(t *testing.T) {
	tests := map[string]func(){
		"ender pearl": func() { NewEnderPearl(world.EntitySpawnOpts{}, nil) },
		"attached firework": func() {
			NewFireworkAttached(world.EntitySpawnOpts{}, item.Firework{}, nil)
		},
	}
	for name, construct := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("constructor accepted a nil owner")
				}
			}()
			construct()
		})
	}
}
