package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestLiquidLoadedUsesWorldBlockRegistry(t *testing.T) {
	br := NewBlockRegistry()
	liquid := customLiquidTestBlock{}
	br.RegisterBlockState(BlockState{Name: "test:liquid", Properties: map[string]any{}})
	br.RegisterBlock(liquid)

	w := Config{Blocks: br}.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	pos := cube.Pos{0, 64, 0}
	var (
		got Liquid
		ok  bool
	)
	<-w.Exec(func(tx *Tx) {
		c := tx.World().chunk(chunkPosFromBlockPos(pos))
		c.SetBlock(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0, tx.World().conf.Blocks.BlockRuntimeID(liquid))
		c.modified = true

		got, ok = tx.LiquidLoaded(pos)
	})
	if !ok {
		t.Fatal("LiquidLoaded returned ok=false, want true")
	}
	if got != liquid {
		t.Fatalf("LiquidLoaded returned %#v, want %#v", got, liquid)
	}
}

func TestBlockLoadedDoesNotCreateMissingNBTBlockEntity(t *testing.T) {
	br := NewBlockRegistry()
	block := customNBTTestBlock{}
	br.RegisterBlockState(BlockState{Name: "test:nbt_block", Properties: map[string]any{}})
	br.RegisterBlock(block)

	w := Config{Blocks: br}.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	pos := cube.Pos{0, 64, 0}
	var (
		got         Block
		ok, created bool
	)
	<-w.Exec(func(tx *Tx) {
		c := tx.World().chunk(chunkPosFromBlockPos(pos))
		c.SetBlock(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0, tx.World().conf.Blocks.BlockRuntimeID(block))
		c.modified = true

		got, ok = tx.BlockLoaded(pos)
		_, created = c.BlockEntities[pos]
	})
	if !ok {
		t.Fatal("BlockLoaded returned ok=false, want true")
	}
	if got != block {
		t.Fatalf("BlockLoaded returned %#v, want %#v", got, block)
	}
	if created {
		t.Fatal("BlockLoaded created missing block entity data")
	}
}

type customLiquidTestBlock struct{}

func (customLiquidTestBlock) EncodeBlock() (string, map[string]any) {
	return "test:liquid", nil
}
func (customLiquidTestBlock) Hash() (uint64, uint64) { return 1 << 45, 0 }
func (customLiquidTestBlock) Model() BlockModel      { return redstoneCancellationModel{} }
func (customLiquidTestBlock) LiquidDepth() int       { return 0 }
func (customLiquidTestBlock) SpreadDecay() int       { return 1 }
func (customLiquidTestBlock) WithDepth(int, bool) Liquid {
	return customLiquidTestBlock{}
}
func (customLiquidTestBlock) LiquidFalling() bool      { return false }
func (customLiquidTestBlock) BlastResistance() float64 { return 100 }
func (customLiquidTestBlock) LiquidType() string       { return "test" }
func (customLiquidTestBlock) Harden(cube.Pos, *Tx, *cube.Pos) bool {
	return false
}
func (customLiquidTestBlock) LiquidRemoveBlock(cube.Pos, *Tx, Block) {}

type customNBTTestBlock struct {
	Decoded bool
}

func (customNBTTestBlock) EncodeBlock() (string, map[string]any) {
	return "test:nbt_block", nil
}
func (customNBTTestBlock) Hash() (uint64, uint64) { return 1 << 46, 0 }
func (customNBTTestBlock) Model() BlockModel      { return redstoneCancellationModel{} }
func (customNBTTestBlock) EncodeNBT() map[string]any {
	return nil
}
func (customNBTTestBlock) DecodeNBT(map[string]any) any {
	return customNBTTestBlock{Decoded: true}
}
