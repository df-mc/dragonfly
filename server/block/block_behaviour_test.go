package block_test

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
)

// TestTorchBreaksWithoutSupport verifies that a torch is broken by a neighbour
// update on the tick after its supporting block is removed, using a
// synchronous World to make the tick deterministic.
func TestTorchBreaksWithoutSupport(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	support, torch := cube.Pos{0, 0, 0}, cube.Pos{0, 1, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(support, block.Stone{}, nil)
		tx.SetBlock(torch, block.Torch{Facing: cube.FaceDown}, nil)
		tx.SetBlock(support, block.Air{}, nil)
	})
	w.AdvanceTick()

	b, err := world.Call(context.Background(), w, func(tx *world.Tx) (world.Block, error) {
		return tx.Block(torch), nil
	})
	if err != nil {
		t.Fatalf("read torch block: %v", err)
	}
	if b != (block.Air{}) {
		t.Errorf("expected torch to break after removing its support, got %v", b)
	}
}

// TestStrawBedKeepsSleeper verifies that the sleeper stored on a straw bed's head survives being written
// to and read back from the world, which its runtime ID alone cannot carry.
func TestStrawBedKeepsSleeper(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	foot, head := cube.Pos{0, 1, 0}, cube.Pos{0, 1, 1}
	h := world.EntitySpawnOpts{}.New(sleeperTestEntityType{}, sleeperTestEntityConfig{})
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(foot, block.StrawBed{Facing: cube.South}, nil)
		tx.SetBlock(head, block.StrawBed{Facing: cube.South, Head: true}, nil)
		tx.Block(head).(block.StrawBed).StartSleeping(head, tx, h)
	})
	w.AdvanceTick()

	got, err := world.Call(context.Background(), w, func(tx *world.Tx) (block.StrawBed, error) {
		b, _ := tx.Block(head).(block.StrawBed)
		return b, nil
	})
	if err != nil {
		t.Fatalf("read straw bed: %v", err)
	}
	if !got.Occupied {
		t.Errorf("expected the head to be occupied after StartSleeping")
	}
	if got.SleepingEntity() != h {
		t.Errorf("head sleeper = %v, want the entity that started sleeping", got.SleepingEntity())
	}
}

// sleeperTestEntityType is never opened: the test only needs a handle to store.
type sleeperTestEntityType struct{}

func (sleeperTestEntityType) Open(*world.Tx, *world.EntityHandle, *world.EntityData) world.Entity {
	return nil
}
func (sleeperTestEntityType) EncodeEntity() string                        { return "test:sleeper" }
func (sleeperTestEntityType) BBox(world.Entity) cube.BBox                 { return cube.BBox{} }
func (sleeperTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (sleeperTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type sleeperTestEntityConfig struct{}

func (sleeperTestEntityConfig) Apply(*world.EntityData) {}
