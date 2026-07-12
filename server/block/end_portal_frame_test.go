package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestEndPortalFrameLightEmission(t *testing.T) {
	emitter, ok := any(block.EndPortalFrame{}).(interface{ LightEmissionLevel() uint8 })
	if !ok {
		t.Fatal("EndPortalFrame does not emit light, want light level 1")
	}
	if got := emitter.LightEmissionLevel(); got != 1 {
		t.Fatalf("LightEmissionLevel() = %d, want 1", got)
	}
}

func TestEndPortalFrameCollisionIncludesEye(t *testing.T) {
	withoutEye := (block.EndPortalFrame{}).Model().BBox(cube.Pos{}, nil)
	withEye := (block.EndPortalFrame{Eye: true}).Model().BBox(cube.Pos{}, nil)

	if len(withoutEye) != 1 {
		t.Fatalf("frame without eye has %d collision boxes, want 1", len(withoutEye))
	}
	if len(withEye) != 2 {
		t.Fatalf("frame with eye has %d collision boxes, want 2", len(withEye))
	}
	wantEye := cube.Box(0.3125, 0.8125, 0.3125, 0.6875, 1, 0.6875)
	if withEye[1] != wantEye {
		t.Fatalf("eye collision box = %v, want %v", withEye[1], wantEye)
	}
}
