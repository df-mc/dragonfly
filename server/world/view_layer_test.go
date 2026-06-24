package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestViewLayerScopesBlockOverridesByWorld(t *testing.T) {
	v := NewViewLayer(nil)
	w1, w2 := &World{}, &World{}
	pos := cube.Pos{1, 64, 1}
	b1, b2 := viewLayerTestBlock(1), viewLayerTestBlock(2)

	v.ViewBlock(w1, pos, b1)
	if b, ok := v.Block(w1, pos); !ok || b != b1 {
		t.Fatalf("expected block %v in first world, got %v, %v", b1, b, ok)
	}

	if _, ok := v.Block(w2, pos); ok {
		t.Fatal("expected no block override in second world")
	}
	v.ViewBlock(w2, pos, b2)
	if b, ok := v.Block(w2, pos); !ok || b != b2 {
		t.Fatalf("expected block %v in second world, got %v, %v", b2, b, ok)
	}

	if b, ok := v.Block(w1, pos); !ok || b != b1 {
		t.Fatalf("expected block %v in first world after switch back, got %v, %v", b1, b, ok)
	}
}

type viewLayerTestBlock uint64

func (b viewLayerTestBlock) EncodeBlock() (string, map[string]any) { return "test:block", nil }

func (b viewLayerTestBlock) Hash() (uint64, uint64) { return uint64(b), 0 }

func (viewLayerTestBlock) Model() BlockModel { return unknownModel{} }
