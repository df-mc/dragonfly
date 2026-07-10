package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

func TestShulkerPushDelta(t *testing.T) {
	lid := cube.Box(0, 0, 0, 1, 1, 1)
	tests := []struct {
		face   cube.Face
		entity cube.BBox
		want   mgl64.Vec3
	}{
		{cube.FaceDown, cube.Box(0, -0.75, 0, 1, 0.25, 1), mgl64.Vec3{0, -0.25, 0}},
		{cube.FaceUp, cube.Box(0, 0.75, 0, 1, 1.75, 1), mgl64.Vec3{0, 0.25, 0}},
		{cube.FaceNorth, cube.Box(0, 0, -0.75, 1, 1, 0.25), mgl64.Vec3{0, 0, -0.25}},
		{cube.FaceSouth, cube.Box(0, 0, 0.75, 1, 1, 1.75), mgl64.Vec3{0, 0, 0.25}},
		{cube.FaceWest, cube.Box(-0.75, 0, 0, 0.25, 1, 1), mgl64.Vec3{-0.25, 0, 0}},
		{cube.FaceEast, cube.Box(0.75, 0, 0, 1.75, 1, 1), mgl64.Vec3{0.25, 0, 0}},
	}
	for _, test := range tests {
		t.Run(test.face.String(), func(t *testing.T) {
			if got := shulkerPushDelta(test.face, lid, test.entity); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}
