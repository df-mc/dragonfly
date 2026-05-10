package world

import (
	"slices"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

var _ Handler = minimalRedstoneTestHandler{}

type minimalRedstoneTestHandler struct{}

func (minimalRedstoneTestHandler) HandleLiquidFlow(*Context, cube.Pos, cube.Pos, Liquid, Block) {}
func (minimalRedstoneTestHandler) HandleLiquidDecay(*Context, cube.Pos, Liquid, Liquid)         {}
func (minimalRedstoneTestHandler) HandleLiquidHarden(*Context, cube.Pos, Block, Block, Block)   {}
func (minimalRedstoneTestHandler) HandleSound(*Context, Sound, mgl64.Vec3)                      {}
func (minimalRedstoneTestHandler) HandleFireSpread(*Context, cube.Pos, cube.Pos)                {}
func (minimalRedstoneTestHandler) HandleBlockBurn(*Context, cube.Pos)                           {}
func (minimalRedstoneTestHandler) HandleCropTrample(*Context, cube.Pos)                         {}
func (minimalRedstoneTestHandler) HandleLeavesDecay(*Context, cube.Pos)                         {}
func (minimalRedstoneTestHandler) HandleEntitySpawn(*Tx, Entity)                                {}
func (minimalRedstoneTestHandler) HandleEntityDespawn(*Tx, Entity)                              {}
func (minimalRedstoneTestHandler) HandleExplosion(*Context, mgl64.Vec3, *[]Entity, *[]cube.Pos, *float64, *bool) {
}
func (minimalRedstoneTestHandler) HandleRedstoneUpdate(*Context, RedstoneUpdate) {}
func (minimalRedstoneTestHandler) HandleClose(*Tx)                               {}

func TestClampRedstonePower(t *testing.T) {
	tests := []struct {
		name  string
		power int
		want  int
	}{
		{name: "negative", power: -1, want: 0},
		{name: "zero", power: 0, want: 0},
		{name: "middle", power: 8, want: 8},
		{name: "maximum", power: 15, want: 15},
		{name: "over maximum", power: 16, want: 15},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := clampRedstonePower(test.power); got != test.want {
				t.Fatalf("clampRedstonePower(%d) = %d, want %d", test.power, got, test.want)
			}
		})
	}
}

func TestCompareBlockPosSortOrder(t *testing.T) {
	positions := []cube.Pos{
		{3, 2, 1},
		{2, 1, 2},
		{1, 1, 1},
		{0, 1, 1},
		{0, 0, 9},
		{9, 0, -1},
	}
	want := []cube.Pos{
		{9, 0, -1},
		{0, 0, 9},
		{0, 1, 1},
		{1, 1, 1},
		{2, 1, 2},
		{3, 2, 1},
	}

	slices.SortFunc(positions, compareBlockPos)
	if !slices.Equal(positions, want) {
		t.Fatalf("sorted positions = %v, want %v", positions, want)
	}
	if got := compareBlockPos(cube.Pos{1, 2, 3}, cube.Pos{1, 2, 3}); got != 0 {
		t.Fatalf("compareBlockPos(equal positions) = %d, want 0", got)
	}
}

func TestRedstoneRelayerNeighbourPositionsAreDeterministic(t *testing.T) {
	engine := newRedstoneEngine(0)
	pos := cube.Pos{0, 64, 0}
	got := engine.redstoneRelayerNeighbourPositions(nil, pos, redstoneNeighbourOrderTestBlock{neighbours: []cube.Pos{
		{1, 64, 0},
		{0, 63, 0},
		{0, 64, -1},
		{-1, 64, 0},
		{0, 65, 0},
		{0, 64, 1},
	}})
	want := []cube.Pos{
		{0, 63, 0},
		{0, 64, -1},
		{-1, 64, 0},
		{1, 64, 0},
		{0, 64, 1},
		{0, 65, 0},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("redstone relayer neighbours = %v, want %v", got, want)
	}
}

func TestRedstoneGraphID(t *testing.T) {
	if got := redstoneGraphID(nil, nil); got != 0 {
		t.Fatalf("redstoneGraphID(nil) = %d, want 0", got)
	}
	if got := redstoneGraphID([]redstoneNode{}, []redstoneEdge{{from: 0, to: 1, weight: 1}}); got != 0 {
		t.Fatalf("redstoneGraphID(empty) = %d, want 0", got)
	}

	nodes := []redstoneNode{
		{pos: cube.Pos{0, 0, 0}, power: 0, source: true},
		{pos: cube.Pos{1, -4, 3}, power: 7, sink: true},
		{pos: cube.Pos{-8, 12, 2}, power: 15, source: true, sink: true},
	}
	edges := []redstoneEdge{
		{from: 0, to: 1, weight: 1},
		{from: 1, to: 2, weight: 2},
	}
	id := redstoneGraphID(nodes, edges)
	if id == 0 {
		t.Fatalf("redstoneGraphID(non-empty nodes) = 0, want non-zero")
	}
	if got := redstoneGraphID(slices.Clone(nodes), slices.Clone(edges)); got != id {
		t.Fatalf("redstoneGraphID(cloned nodes) = %d, want %d", got, id)
	}

	changedPower := slices.Clone(nodes)
	changedPower[1].power++
	if got := redstoneGraphID(changedPower, edges); got != id {
		t.Fatalf("redstoneGraphID(nodes with changed power) = %d, want topology ID %d", got, id)
	}

	movedNode := slices.Clone(nodes)
	movedNode[2].pos[0]++
	if got := redstoneGraphID(movedNode, edges); got == id {
		t.Fatalf("redstoneGraphID(nodes with moved position) = %d, want value different from %d", got, id)
	}

	changedEdge := slices.Clone(edges)
	changedEdge[0].weight++
	if got := redstoneGraphID(nodes, changedEdge); got == id {
		t.Fatalf("redstoneGraphID(nodes with changed edge) = %d, want value different from %d", got, id)
	}
}

func TestRedstoneEngineInvalidateAround(t *testing.T) {
	var nilEngine *redstoneEngine
	nilEngine.invalidateAround(cube.Pos{0, 0, 0}, cube.Pos{0, 0, 0}, RedstoneUpdateCauseBlockUpdate, cube.Range{0, 0})

	engine := newRedstoneEngine(42)
	pos, changed := cube.Pos{8, 0, 8}, cube.Pos{9, 0, 8}
	engine.invalidateAround(pos, changed, RedstoneUpdateCauseBlockUpdate, cube.Range{0, 1})

	want := map[cube.Pos]redstoneDirty{
		{8, 0, 8}: redstoneDirty{changed: changed, cause: RedstoneUpdateCauseBlockUpdate},
		{9, 0, 8}: redstoneDirty{changed: changed, cause: RedstoneUpdateCauseBlockUpdate},
		{7, 0, 8}: redstoneDirty{changed: changed, cause: RedstoneUpdateCauseBlockUpdate},
		{8, 1, 8}: redstoneDirty{changed: changed, cause: RedstoneUpdateCauseBlockUpdate},
		{8, 0, 9}: redstoneDirty{changed: changed, cause: RedstoneUpdateCauseBlockUpdate},
		{8, 0, 7}: redstoneDirty{changed: changed, cause: RedstoneUpdateCauseBlockUpdate},
	}
	if len(engine.dirty) != len(want) {
		t.Fatalf("dirty positions = %v, want %v", engine.dirty, want)
	}
	for pos, dirty := range want {
		if got, ok := engine.dirty[pos]; !ok || got != dirty {
			t.Fatalf("dirty[%v] = %v, %t; want %v, true", pos, got, ok, dirty)
		}
	}

	engine.invalidateAround(cube.Pos{0, -1, 0}, changed, RedstoneUpdateCauseScheduledTick, cube.Range{0, 1})
	if len(engine.dirty) != len(want) {
		t.Fatalf("out-of-bounds invalidation changed dirty positions to %v, want %v", engine.dirty, want)
	}
}

func TestScheduledTickQueueReschedulesEarlierTick(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}

	queue.schedule(DefaultBlockRegistry, pos, b, time.Second)
	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/10)
	queue.schedule(DefaultBlockRegistry, pos, b, time.Second*2)

	index := scheduledTickIndex{pos: pos, hash: BlockHash(b)}
	if got, want := queue.scheduledTicks[index], int64(102); got != want {
		t.Fatalf("scheduled tick = %d, want %d", got, want)
	}
	ticks := queue.fromChunk(chunkPosFromBlockPos(pos))
	if len(ticks) != 1 {
		t.Fatalf("active ticks = %v, want one tick", ticks)
	}
	if got, want := ticks[0].t, int64(102); got != want {
		t.Fatalf("fromChunk tick = %d, want %d", got, want)
	}
}

func TestScheduledTickQueueRemoveChunkClearsSchedule(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}

	queue.schedule(DefaultBlockRegistry, pos, b, time.Second)
	queue.removeChunk(chunkPosFromBlockPos(pos))

	if len(queue.ticks) != 0 {
		t.Fatalf("ticks after removeChunk = %v, want empty", queue.ticks)
	}
	if len(queue.scheduledTicks) != 0 {
		t.Fatalf("scheduled ticks after removeChunk = %v, want empty", queue.scheduledTicks)
	}
}

func TestScheduledTickQueueCanRescheduleWhileCurrentTickIsDue(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}
	index := scheduledTickIndex{pos: pos, hash: BlockHash(b)}
	queue.scheduledTicks[index] = 100

	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/2)
	if got, want := queue.scheduledTicks[index], int64(110); got != want {
		t.Fatalf("rescheduled tick = %d, want %d", got, want)
	}
}

type scheduledTickTestBlock struct{}

func (scheduledTickTestBlock) EncodeBlock() (string, map[string]any) {
	return "test:scheduled_tick", nil
}
func (scheduledTickTestBlock) Hash() (uint64, uint64) { return 1 << 40, 0 }
func (scheduledTickTestBlock) Model() BlockModel      { return nil }

type redstoneNeighbourOrderTestBlock struct {
	neighbours []cube.Pos
}

func (b redstoneNeighbourOrderTestBlock) RedstoneRelayerNeighbours(cube.Pos, *Tx) []cube.Pos {
	return slices.Clone(b.neighbours)
}
func (redstoneNeighbourOrderTestBlock) EncodeBlock() (string, map[string]any) {
	return "test:redstone_neighbour_order", nil
}
func (redstoneNeighbourOrderTestBlock) Hash() (uint64, uint64) { return 1 << 41, 0 }
func (redstoneNeighbourOrderTestBlock) Model() BlockModel      { return nil }
