package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func TestEndPortalRegistered(t *testing.T) {
	world.New().Close()

	b, ok := world.BlockByName("minecraft:end_portal", nil)
	if !ok {
		t.Fatal("minecraft:end_portal not registered")
	}
	if _, ok := b.(block.EndPortal); !ok {
		t.Fatalf("minecraft:end_portal returned %T, want block.EndPortal", b)
	}
}

func TestEndPortalFrameRegistered(t *testing.T) {
	world.New().Close()

	for _, dir := range cube.Directions() {
		for _, eye := range []bool{false, true} {
			b, ok := world.BlockByName("minecraft:end_portal_frame", map[string]any{
				"minecraft:cardinal_direction": dir.String(),
				"end_portal_eye_bit":           eye,
			})
			if !ok {
				t.Fatalf("minecraft:end_portal_frame not registered for dir=%s eye=%v", dir, eye)
			}
			f, ok := b.(block.EndPortalFrame)
			if !ok {
				t.Fatalf("minecraft:end_portal_frame for dir=%s eye=%v returned %T, want block.EndPortalFrame", dir, eye, b)
			}
			if f.Facing != dir || f.Eye != eye {
				t.Fatalf("minecraft:end_portal_frame returned %+v, want Facing=%s Eye=%v", f, dir, eye)
			}
		}
	}
}
