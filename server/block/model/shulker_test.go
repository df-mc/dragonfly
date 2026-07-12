package model

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestShulkerBBoxDoesNotOverlapFacingBlock(t *testing.T) {
	for _, face := range cube.Faces() {
		boxes := (Shulker{Facing: face, Progress: 10}).BBox(cube.Pos{}, nil)
		if len(boxes) != 1 {
			t.Fatalf("face %v: got %d boxes, want 1", face, len(boxes))
		}
		box := boxes[0]
		if box.Min().X() < 0 || box.Min().Y() < 0 || box.Min().Z() < 0 ||
			box.Max().X() > 1 || box.Max().Y() > 1 || box.Max().Z() > 1 {
			t.Fatalf("face %v: BBox %v..%v overlaps an adjacent block", face, box.Min(), box.Max())
		}
	}
}
