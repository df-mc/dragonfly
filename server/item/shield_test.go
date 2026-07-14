package item

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestShieldRegistered(t *testing.T) {
	it, ok := world.ItemByName("minecraft:shield", 0)
	if !ok {
		t.Fatal("expected minecraft:shield to be registered")
	}
	if _, ok := it.(Shield); !ok {
		t.Fatalf("expected registered item to be Shield, got %T", it)
	}
}
