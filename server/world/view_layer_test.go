package world

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/stretchr/testify/require"
)

func TestViewLayerScopesBlockOverridesByWorld(t *testing.T) {
	v := NewViewLayer(nil)
	w1 := Config{Synchronous: true}.New()
	w2 := Config{Synchronous: true}.New()
	defer w1.Close()
	defer w2.Close()

	pos := cube.Pos{1, 64, 1}
	b1, b2 := viewLayerTestBlock(1), viewLayerTestBlock(2)

	err := w1.Do(func(tx *Tx) {
		v.ViewBlock(tx, pos, b1)
	}).Wait(context.Background())
	require.NoError(t, err)
	if b, ok := v.Block(w1, pos); !ok || b != b1 {
		t.Fatalf("expected block %v in first world, got %v, %v", b1, b, ok)
	}

	if _, ok := v.Block(w2, pos); ok {
		t.Fatal("expected no block override in second world")
	}
	err = w2.Do(func(tx *Tx) {
		v.ViewBlock(tx, pos, b2)
	}).Wait(context.Background())
	require.NoError(t, err)
	if b, ok := v.Block(w2, pos); !ok || b != b2 {
		t.Fatalf("expected block %v in second world, got %v, %v", b2, b, ok)
	}

	if b, ok := v.Block(w1, pos); !ok || b != b1 {
		t.Fatalf("expected block %v in first world after switch back, got %v, %v", b1, b, ok)
	}
}

func TestTxPublicBlockViewersFiltersWithoutMutatingWorldViewers(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	privateViewer := &viewLayerTestViewer{viewLayer: NewViewLayer(nil)}
	publicViewer := &viewLayerTestViewer{viewLayer: NewViewLayer(nil)}
	pos := cube.Pos{0, 64, 0}

	var before, visible, after []Viewer
	err := w.Do(func(tx *Tx) {
		privateViewer.viewLayer.ViewBlock(tx, pos, viewLayerTestBlock(1))

		privateLoader := NewLoader(1, w, privateViewer)
		publicLoader := NewLoader(1, w, publicViewer)
		privateLoader.Load(tx, 1)
		publicLoader.Load(tx, 1)
		defer privateLoader.Close(tx)
		defer publicLoader.Close(tx)

		before = append([]Viewer(nil), tx.Viewers(pos.Vec3())...)
		visible = tx.PublicBlockViewers(pos)
		after = append([]Viewer(nil), tx.Viewers(pos.Vec3())...)
	}).Wait(context.Background())
	require.NoError(t, err)
	require.Equal(t, []Viewer{publicViewer}, visible)
	require.Equal(t, []Viewer{privateViewer, publicViewer}, before)
	require.Equal(t, before, after)
}

func TestNewViewLayerAcceptsEntityOnlyUpdater(t *testing.T) {
	require.NotNil(t, NewViewLayer(viewLayerEntityUpdater{}))
}

func TestViewLayerBlockUpdaterReceivesTransaction(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	updater := &viewLayerBlockUpdater{}
	v := NewViewLayer(updater)
	pos := cube.Pos{1, 64, 1}
	err := w.Do(func(tx *Tx) {
		v.ViewBlock(tx, pos, viewLayerTestBlock(1))
		require.Same(t, tx, updater.tx)
	}).Wait(context.Background())
	require.NoError(t, err)
	require.Equal(t, pos, updater.pos)
}

type viewLayerTestBlock uint64

func (b viewLayerTestBlock) EncodeBlock() (string, map[string]any) { return "test:block", nil }

func (b viewLayerTestBlock) Hash() (uint64, uint64) { return uint64(b), 0 }

func (viewLayerTestBlock) Model() BlockModel { return unknownModel{} }

type viewLayerTestViewer struct {
	NopViewer
	viewLayer *ViewLayer
}

func (v *viewLayerTestViewer) ViewLayer() *ViewLayer {
	return v.viewLayer
}

type viewLayerEntityUpdater struct{}

func (viewLayerEntityUpdater) ViewLayerEntityChanged(Entity) {}

type viewLayerBlockUpdater struct {
	viewLayerEntityUpdater
	tx  *Tx
	pos cube.Pos
}

func (u *viewLayerBlockUpdater) ViewLayerBlockChanged(tx *Tx, pos cube.Pos) {
	u.tx = tx
	u.pos = pos
}
