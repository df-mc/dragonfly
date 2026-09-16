package block

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// TestWeightedPressurePlateUpdatesWhilePressed checks immediate strength changes without extra clicks or delayed release.
func TestWeightedPressurePlateUpdatesWhilePressed(t *testing.T) {
	for _, test := range []struct {
		plateType  PressurePlateType
		additional int
	}{
		{LightWeightedPressurePlate(), 1},
		{HeavyWeightedPressurePlate(), 10},
	} {
		t.Run(test.plateType.Name(), func(t *testing.T) {
			w := world.Config{Synchronous: true}.New()
			defer w.Close()
			h := &pressurePlateSoundTestHandler{}
			w.Handle(h)
			pos := cube.Pos{0, 64, 0}
			var first *world.EntityHandle
			runWorld(w, func(tx *world.Tx) {
				tx.SetBlock(pos.Side(cube.FaceDown), Stone{}, nil)
				plate := PressurePlate{Type: test.plateType}
				tx.SetBlock(pos, plate, nil)
				first = world.EntitySpawnOpts{}.New(redstoneArrowTestEntityType{}, redstoneTNTTestEntityConfig{})
				tx.AddEntityAt(first, mgl64.Vec3{0.5, 64.05, 0.5})
				e, _ := first.Entity(tx)
				plate.EntityInside(pos, tx, e)
				if got := tx.Block(pos).(PressurePlate).Power; got != 1 {
					t.Errorf("first entity power=%d, want 1", got)
				}
			})
			for range 4 {
				w.AdvanceTick()
			}
			runWorld(w, func(tx *world.Tx) {
				var added []world.Entity
				for range test.additional {
					handle := world.EntitySpawnOpts{}.New(redstoneArrowTestEntityType{}, redstoneTNTTestEntityConfig{})
					tx.AddEntityAt(handle, mgl64.Vec3{0.5, 64.05, 0.5})
					e, _ := handle.Entity(tx)
					added = append(added, e)
				}
				tx.Block(pos).(PressurePlate).EntityInside(pos, tx, added[0])
				if got := tx.Block(pos).(PressurePlate).Power; got != 2 {
					t.Errorf("additional entities power=%d, want 2", got)
				}
				for _, e := range added {
					_ = tx.RemoveEntity(e).Close()
				}
				e, _ := first.Entity(tx)
				tx.Block(pos).(PressurePlate).EntityInside(pos, tx, e)
				if got := tx.Block(pos).(PressurePlate).Power; got != 1 {
					t.Errorf("remaining entity power=%d, want 1", got)
				}
				_ = tx.RemoveEntity(e).Close()
			})
			for range 5 {
				w.AdvanceTick()
			}
			runWorld(w, func(tx *world.Tx) {
				if got := tx.Block(pos).(PressurePlate).Power; got != 1 {
					t.Errorf("power before original release tick=%d, want 1", got)
				}
			})
			w.AdvanceTick()
			runWorld(w, func(tx *world.Tx) {
				if got := tx.Block(pos).(PressurePlate).Power; got != 0 {
					t.Errorf("power at original release tick=%d, want 0", got)
				}
			})
			if h.on != 1 || h.off != 1 {
				t.Errorf("plate clicks on=%d off=%d, want one each", h.on, h.off)
			}
		})
	}
}

// pressurePlateSoundTestHandler records pressure plate transitions.
type pressurePlateSoundTestHandler struct {
	world.NopHandler
	on, off int
}

// HandleSound counts activation and release sounds.
func (h *pressurePlateSoundTestHandler) HandleSound(_ *world.Context, s world.Sound, _ mgl64.Vec3) {
	switch s.(type) {
	case sound.PressurePlateClickOn:
		h.on++
	case sound.PressurePlateClickOff:
		h.off++
	}
}
