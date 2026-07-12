package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

func TestFlintAndSteelDoesNotIgniteBesideWaterloggedCandle(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	handler := &candleSoundHandler{}
	w.Handle(handler)
	pos := cube.Pos{0, 64, 0}
	var used bool
	runWorld(w, func(tx *world.Tx) {
		tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
		tx.SetBlock(pos, Candle{}, nil)
		tx.SetLiquid(pos, Water{Depth: 8})

		used = (item.FlintAndSteel{}).UseOnBlock(pos, cube.FaceUp, mgl64.Vec3{}, tx, nil, &item.UseContext{})
	})

	if used {
		t.Fatal("flint and steel reported igniting a waterlogged candle")
	}
	if handler.igniteSounds != 0 {
		t.Fatalf("ignite sounds = %d, want 0", handler.igniteSounds)
	}
}

type candleSoundHandler struct {
	world.NopHandler
	igniteSounds int
}

func (h *candleSoundHandler) HandleSound(_ *world.Context, s world.Sound, _ mgl64.Vec3) {
	if _, ok := s.(sound.Ignite); ok {
		h.igniteSounds++
	}
}
