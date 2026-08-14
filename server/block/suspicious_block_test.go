package block_test

import (
	"context"
	"slices"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// TestSuspiciousBlockBrushing verifies that brushing a suspicious block ten times shows the correct amount of
// dust after every brushing action, turns the block into the block it is made of and drops the item it holds.
func TestSuspiciousBlockBrushing(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos, loot := cube.Pos{0, 1, 0}, item.NewStack(item.Diamond{}, 1)
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, block.Stone{}, nil)
		tx.SetBlock(pos, block.SuspiciousBlock{Item: loot}, nil)
	})

	dust := []int{1, 1, 2, 2, 2, 3, 3, 3, 3}
	for i, expected := range dust {
		completed, err := world.Call(context.Background(), w, func(tx *world.Tx) (bool, error) {
			return tx.Block(pos).(block.Brushable).Brush(pos, tx, cube.FaceUp), nil
		})
		if err != nil {
			t.Fatalf("brush block: %v", err)
		}
		if completed {
			t.Fatalf("brushing completed after %v brushing actions, expected %v", i+1, len(dust)+1)
		}
		b, err := world.Call(context.Background(), w, func(tx *world.Tx) (world.Block, error) {
			return tx.Block(pos), nil
		})
		if err != nil {
			t.Fatalf("read block: %v", err)
		}
		if s, ok := b.(block.SuspiciousBlock); !ok || s.Dust != expected {
			t.Fatalf("expected %v dust after %v brushing actions, got %v", expected, i+1, b)
		}
	}

	completed, err := world.Call(context.Background(), w, func(tx *world.Tx) (bool, error) {
		return tx.Block(pos).(block.Brushable).Brush(pos, tx, cube.FaceUp), nil
	})
	if err != nil {
		t.Fatalf("brush block: %v", err)
	}
	if !completed {
		t.Fatal("expected brushing to complete after ten brushing actions")
	}

	b, err := world.Call(context.Background(), w, func(tx *world.Tx) (world.Block, error) {
		return tx.Block(pos), nil
	})
	if err != nil {
		t.Fatalf("read block: %v", err)
	}
	if b != (block.Sand{}) {
		t.Errorf("expected suspicious sand to turn into sand once brushed, got %v", b)
	}

	dropped, err := world.Call(context.Background(), w, func(tx *world.Tx) ([]item.Stack, error) {
		var stacks []item.Stack
		for e := range tx.Entities() {
			if ent, ok := e.(*entity.Ent); ok {
				if behaviour, ok := ent.Behaviour().(*entity.ItemBehaviour); ok {
					stacks = append(stacks, behaviour.Item())
				}
			}
		}
		return stacks, nil
	})
	if err != nil {
		t.Fatalf("read entities: %v", err)
	}
	if !slices.ContainsFunc(dropped, loot.Equal) {
		t.Errorf("expected %v to be dropped once brushing completed, got %v", loot, dropped)
	}
}

// TestSuspiciousBlockBrushingResets verifies that a suspicious block loses its brushing progress again after
// it has not been brushed for two seconds.
func TestSuspiciousBlockBrushingResets(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 1, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, block.Stone{}, nil)
		tx.SetBlock(pos, block.SuspiciousBlock{Gravel: true}, nil)
		for range 6 {
			tx.Block(pos).(block.Brushable).Brush(pos, tx, cube.FaceUp)
		}
	})
	for range 60 {
		w.AdvanceTick()
	}

	b, err := world.Call(context.Background(), w, func(tx *world.Tx) (world.Block, error) {
		return tx.Block(pos), nil
	})
	if err != nil {
		t.Fatalf("read block: %v", err)
	}
	if s, ok := b.(block.SuspiciousBlock); !ok || s.Dust != 0 {
		t.Errorf("expected suspicious gravel to lose its brushing progress, got %v", b)
	}
}

// TestSuspiciousBlockNBT verifies that the brushing progress and the item held by a suspicious block survive
// a block entity data round trip.
func TestSuspiciousBlockNBT(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos, loot := cube.Pos{0, 1, 0}, item.NewStack(item.Emerald{}, 1)
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{0, 0, 0}, block.Stone{}, nil)
		tx.SetBlock(pos, block.SuspiciousBlock{Item: loot}, nil)
		for range 4 {
			tx.Block(pos).(block.Brushable).Brush(pos, tx, cube.FaceEast)
		}
	})

	data, err := world.Call(context.Background(), w, func(tx *world.Tx) (map[string]any, error) {
		return tx.Block(pos).(world.NBTer).EncodeNBT(), nil
	})
	if err != nil {
		t.Fatalf("encode block entity data: %v", err)
	}
	for k, expected := range map[string]any{
		"id":              "BrushableBlock",
		"type":            "minecraft:suspicious_sand",
		"brush_count":     int32(4),
		"brush_direction": byte(cube.FaceEast),
	} {
		if data[k] != expected {
			t.Errorf("expected %v to be %v, got %v", k, expected, data[k])
		}
	}

	decoded := block.SuspiciousBlock{Dust: 2}.DecodeNBT(data).(block.SuspiciousBlock)
	if !decoded.Item.Equal(loot) {
		t.Errorf("expected decoded block to hold %v, got %v", loot, decoded.Item)
	}
	if decoded.Dust != 2 {
		t.Errorf("expected decoded block to keep its dust, got %v", decoded.Dust)
	}
}

// TestSuspiciousBlockBreaksOnLanding verifies that a suspicious block that falls breaks without dropping
// anything instead of being placed back into the world.
func TestSuspiciousBlockBreaksOnLanding(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	ground, pos := cube.Pos{0, 0, 0}, cube.Pos{0, 2, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(ground, block.Stone{}, nil)
		tx.SetBlock(pos, block.SuspiciousBlock{}, nil)
	})
	for range 20 {
		w.AdvanceTick()
	}

	blocks, err := world.Call(context.Background(), w, func(tx *world.Tx) ([]world.Block, error) {
		return []world.Block{tx.Block(pos), tx.Block(cube.Pos{0, 1, 0})}, nil
	})
	if err != nil {
		t.Fatalf("read blocks: %v", err)
	}
	for _, b := range blocks {
		if b != (block.Air{}) {
			t.Errorf("expected suspicious sand to break when it falls, got %v", b)
		}
	}
}
