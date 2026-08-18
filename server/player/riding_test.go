package player

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
)

func TestControllerStateChangedWhenControllerMovesSeat(t *testing.T) {
	handle := &world.EntityHandle{}
	before := ridingController{handle: handle, seat: 0}
	after := ridingController{handle: handle, seat: 1}

	if !controllerStateChanged(before, after) {
		t.Fatal("expected a controller moving seats to refresh controller metadata")
	}
}
