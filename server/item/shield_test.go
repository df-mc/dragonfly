package item

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestShieldProperties(t *testing.T) {
	shield := Shield{}
	if shield.MaxCount() != 1 {
		t.Fatalf("expected shield max count 1, got %v", shield.MaxCount())
	}
	if !shield.OffHand() {
		t.Fatal("expected shield to be valid in the off hand")
	}

	info := shield.DurabilityInfo()
	if info.MaxDurability != 337 {
		t.Fatalf("expected shield max durability 337, got %v", info.MaxDurability)
	}

	name, meta := shield.EncodeItem()
	if name != "minecraft:shield" || meta != 0 {
		t.Fatalf("expected minecraft:shield/0 encoding, got %v/%v", name, meta)
	}
}

func TestShieldRegistered(t *testing.T) {
	it, ok := world.ItemByName("minecraft:shield", 0)
	if !ok {
		t.Fatal("expected minecraft:shield to be registered")
	}
	if _, ok := it.(Shield); !ok {
		t.Fatalf("expected registered item to be Shield, got %T", it)
	}
}
