package session

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/player/debug"
	"github.com/df-mc/dragonfly/server/world"
)

// TestDebugShapeTiming verifies that the lifetime and render distance of a debug shape reach the packet. Without them
// a shape stays until it is removed explicitly and is drawn at any distance, which is what every shape used to do.
func TestDebugShapeTiming(t *testing.T) {
	shapes := []debug.Shape{
		&debug.Arrow{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Box{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Circle{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Line{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Sphere{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Text{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Cylinder{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Pyramid{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Ellipsoid{Duration: time.Second * 5, MaxRenderDistance: 32},
		&debug.Cone{Duration: time.Second * 5, MaxRenderDistance: 32},
	}
	for _, shape := range shapes {
		ps := debugShapeToProtocol(shape, world.Overworld, 0)

		if got, ok := ps.TotalTimeLeft.Value(); !ok || got != 5 {
			t.Errorf("%T: TotalTimeLeft = %v (set %v), want 5", shape, got, ok)
		}
		if got, ok := ps.MaxRenderDistance.Value(); !ok || got != 32 {
			t.Errorf("%T: MaxRenderDistance = %v (set %v), want 32", shape, got, ok)
		}
	}
}

// TestDebugShapeTimingUnset verifies that a shape with neither set leaves both out of the packet, so that it behaves
// as it did before they existed.
func TestDebugShapeTimingUnset(t *testing.T) {
	ps := debugShapeToProtocol(&debug.Box{}, world.Overworld, 0)
	if _, ok := ps.TotalTimeLeft.Value(); ok {
		t.Error("TotalTimeLeft was set, want it left out")
	}
	if _, ok := ps.MaxRenderDistance.Value(); ok {
		t.Error("MaxRenderDistance was set, want it left out")
	}
}
