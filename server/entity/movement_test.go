package entity

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestStationaryEntityActivatesPressurePlate(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()

	w.Do(func(tx *world.Tx) {
		pos := cube.Pos{0, 64, 0}
		tx.SetBlock(pos, block.PressurePlate{Type: block.LightWeightedPressurePlate()}, nil)
		handle := world.EntitySpawnOpts{Position: mgl64.Vec3{0.5, 64.05, 0.5}}.
			New(stationaryPressurePlateTestType{}, StationaryBehaviourConfig{})
		tx.AddEntity(handle)
		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("stationary entity was not added")
		}
		e.(*Ent).Tick(tx, 0)
		if power := tx.Block(pos).(block.PressurePlate).Power; power != 1 {
			t.Fatalf("pressure plate power = %v, want 1", power)
		}
	}).Wait(context.Background())
}

type stationaryPressurePlateTestType struct{}

func (stationaryPressurePlateTestType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (stationaryPressurePlateTestType) EncodeEntity() string { return "test:stationary" }
func (stationaryPressurePlateTestType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.25, 0, -0.25, 0.25, 0.5, 0.25)
}
func (stationaryPressurePlateTestType) DecodeNBT(map[string]any, *world.EntityData) {}
func (stationaryPressurePlateTestType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
