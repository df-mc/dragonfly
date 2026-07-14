package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestHandlePlayerActionSameTargetContinuationPreservesPrivateBreak(t *testing.T) {
	s := &Session{}
	c := &breakActionControllable{privateVisible: true}
	pos := protocol.BlockPos{1, 2, 3}

	if err := handlePlayerAction(protocol.PlayerActionStartBreak, int32(cube.FaceUp), pos, selfEntityRuntimeID, s, c); err != nil {
		t.Fatal(err)
	}
	c.privateVisible = false
	if err := handlePlayerAction(protocol.PlayerActionContinueDestroyBlock, int32(cube.FaceUp), pos, selfEntityRuntimeID, s, c); err != nil {
		t.Fatal(err)
	}
	if err := handlePlayerAction(protocol.PlayerActionPredictDestroyBlock, int32(cube.FaceUp), pos, selfEntityRuntimeID, s, c); err != nil {
		t.Fatal(err)
	}

	if c.publicBroken {
		t.Fatal("same-target continuation replaced the retained private break with the public block")
	}
	if c.starts != 1 {
		t.Fatalf("expected one break start, got %d", c.starts)
	}
}

type breakActionControllable struct {
	Controllable

	privateVisible bool
	targetPrivate  bool
	publicBroken   bool
	starts         int
}

func (c *breakActionControllable) StartBreaking(cube.Pos, cube.Face) {
	c.starts++
	c.targetPrivate = c.privateVisible
}

func (c *breakActionControllable) FinishBreaking() {
	c.publicBroken = !c.targetPrivate
}
