package world

import (
	"math/rand/v2"
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
func (minimalRedstoneTestHandler) HandleClose(*Tx) {}

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

func TestRedstoneCompileRegionIncludesCustomRelayerNeighbours(t *testing.T) {
	pos, step := cube.Pos{0, 64, 0}, cube.Pos{1, 65, 0}
	w := Config{Blocks: redstoneCustomRelayerTestRegistry()}.New()
	defer w.Close()

	var nodes []redstoneNode
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(pos, redstoneStepRelayer{}, nil)
		tx.SetBlock(step, redstoneStepRelayer{}, nil)

		tx.World().redstone.compileRegion(tx, pos, make(map[cube.Pos]struct{}), &nodes)
	})
	if !redstoneNodeTestContains(nodes, step) {
		t.Fatalf("compiled region nodes = %v, want custom neighbour %v included", nodes, step)
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

func TestRedstoneStrongPowerConductorExcludesRedstoneBlock(t *testing.T) {
	pos := cube.Pos{0, 64, 0}
	if !redstoneStrongPowerConductor(pos, redstoneNamedSolidBlock{name: "minecraft:stone"}, nil, cube.FaceWest) {
		t.Fatal("stone was not treated as a strong-power conductor")
	}
	if redstoneStrongPowerConductor(pos, redstoneNamedSolidBlock{name: "minecraft:redstone_block"}, nil, cube.FaceWest) {
		t.Fatal("redstone block was treated as a strong-power conductor")
	}
}

func TestRedstoneTorchBurnoutState(t *testing.T) {
	engine := newRedstoneEngine(0)
	pos := cube.Pos{0, 64, 0}
	for i := range redstoneTorchBurnoutThreshold - 1 {
		if engine.recordRedstoneTorchToggle(pos, int64(i)) {
			t.Fatalf("torch burned out after %d toggles, want below threshold", i+1)
		}
	}
	if !engine.recordRedstoneTorchToggle(pos, redstoneTorchBurnoutThreshold-1) {
		t.Fatal("torch did not burn out at threshold")
	}
	burnedOut, recoverable := engine.redstoneTorchBurnoutStatus(pos, redstoneTorchBurnoutThreshold)
	if !burnedOut || recoverable {
		t.Fatalf("burnout status = burnedOut %t recoverable %t, want true false", burnedOut, recoverable)
	}
	burnedOut, recoverable = engine.redstoneTorchBurnoutStatus(pos, redstoneTorchBurnoutWindowTicks+redstoneTorchBurnoutThreshold)
	if !burnedOut || !recoverable {
		t.Fatalf("expired burnout status = burnedOut %t recoverable %t, want true true", burnedOut, recoverable)
	}
	engine.clearRedstoneTorchBurnout(pos)
	burnedOut, recoverable = engine.redstoneTorchBurnoutStatus(pos, redstoneTorchBurnoutWindowTicks+redstoneTorchBurnoutThreshold)
	if burnedOut || recoverable {
		t.Fatalf("cleared burnout status = burnedOut %t recoverable %t, want false false", burnedOut, recoverable)
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

func TestRedstoneCancelledSourceDoesNotPropagate(t *testing.T) {
	sourcePos, sinkPos := cube.Pos{0, 64, 0}, cube.Pos{1, 64, 0}
	w := Config{Blocks: redstoneCancellationTestRegistry()}.New()
	defer w.Close()

	w.Handle(&redstoneCancellationHandler{cancel: map[cube.Pos]struct{}{sourcePos: {}}})
	var sinkPowered bool
	var sourceOutput int
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(sourcePos, redstoneCancellationSource{Power: 15}, nil)
		tx.SetBlock(sinkPos, redstoneCancellationConsumer{}, nil)
		tx.World().redstone.tick(tx, 1)

		sinkPowered = tx.Block(sinkPos).(redstoneCancellationConsumer).Powered
		sourceOutput = tx.World().redstone.output[sourcePos]
	})
	if sinkPowered {
		t.Fatalf("sink powered after cancelling source update")
	}
	if sourceOutput != 0 {
		t.Fatalf("stored source output = %d, want 0", sourceOutput)
	}
}

func TestRedstoneCancelledConsumerDoesNotUpdate(t *testing.T) {
	sourcePos, sinkPos := cube.Pos{0, 64, 0}, cube.Pos{1, 64, 0}
	w := Config{Blocks: redstoneCancellationTestRegistry()}.New()
	defer w.Close()

	w.Handle(&redstoneCancellationHandler{cancel: map[cube.Pos]struct{}{sinkPos: {}}})
	var sinkPowered bool
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(sourcePos, redstoneCancellationSource{Power: 15}, nil)
		tx.SetBlock(sinkPos, redstoneCancellationConsumer{}, nil)
		tx.World().redstone.tick(tx, 1)

		sinkPowered = tx.Block(sinkPos).(redstoneCancellationConsumer).Powered
	})
	if sinkPowered {
		t.Fatalf("sink powered after cancelling consumer update")
	}
}

func TestRedstoneCancelledActionDoesNotRun(t *testing.T) {
	sourcePos, actionPos := cube.Pos{0, 64, 0}, cube.Pos{1, 64, 0}
	w := Config{Blocks: redstoneCancellationTestRegistry()}.New()
	defer w.Close()

	actions := 0
	redstoneCancellationActions = &actions
	t.Cleanup(func() {
		redstoneCancellationActions = nil
	})
	w.Handle(&redstoneCancellationHandler{cancel: map[cube.Pos]struct{}{actionPos: {}}})
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(sourcePos, redstoneCancellationSource{Power: 15}, nil)
		tx.SetBlock(actionPos, redstoneCancellationAction{}, nil)
		tx.World().redstone.tick(tx, 1)
	})
	if actions != 0 {
		t.Fatalf("actions = %d, want 0", actions)
	}
}

func TestRedstoneRelayerToSinkDoesNotLosePower(t *testing.T) {
	sourcePos, relayerPos, sinkPos := cube.Pos{0, 64, 0}, cube.Pos{1, 64, 0}, cube.Pos{2, 64, 0}
	w := Config{Blocks: redstoneSignalLossTestRegistry()}.New()
	defer w.Close()

	var directPower, sinkPower int
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(sourcePos, redstoneLossSource{Power: 15}, nil)
		tx.SetBlock(relayerPos, redstoneLossRelayer{}, nil)
		tx.SetBlock(sinkPos, redstoneLossConsumer{}, nil)

		directPower = tx.RedstonePower(sinkPos)
		tx.World().redstone.tick(tx, 1)
		if sink, ok := tx.Block(sinkPos).(redstoneLossConsumer); ok {
			sinkPower = sink.Power
		}
	})
	if directPower != 15 {
		t.Fatalf("powerFrom through relayer into sink = %d, want 15", directPower)
	}
	if sinkPower != 15 {
		t.Fatalf("graph power through relayer into sink = %d, want 15", sinkPower)
	}
}

func TestRedstoneUpdateIncludesAfterForConsumerStateChange(t *testing.T) {
	sourcePos, sinkPos := cube.Pos{0, 64, 0}, cube.Pos{1, 64, 0}
	w := Config{Blocks: redstoneCancellationTestRegistry()}.New()
	defer w.Close()

	handler := &redstoneRecordingHandler{}
	w.Handle(handler)
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(sourcePos, redstoneCancellationSource{Power: 15}, nil)
		tx.SetBlock(sinkPos, redstoneCancellationConsumer{}, nil)
		tx.World().redstone.tick(tx, 1)
	})

	for _, update := range handler.updates {
		if update.Pos != sinkPos {
			continue
		}
		after, ok := update.After.(redstoneCancellationConsumer)
		if !ok {
			t.Fatalf("consumer update After = %#v, want redstoneCancellationConsumer", update.After)
		}
		if !after.Powered {
			t.Fatalf("consumer update After.Powered = false, want true")
		}
		return
	}
	t.Fatalf("no redstone update recorded for sink %v; updates=%v", sinkPos, handler.updates)
}

func TestScheduledTickQueueKeepsLatestTickForSameBlock(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}

	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/20)
	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/10)

	index := scheduledTickIndex{pos: pos, hash: DefaultBlockRegistry.BlockHash(b)}
	if got, want := queue.furthestTicks[index], int64(102); got != want {
		t.Fatalf("furthest tick = %d, want %d", got, want)
	}
	ticks := queue.fromChunk(chunkPosFromBlockPos(pos))
	if len(ticks) != 1 {
		t.Fatalf("active ticks = %v, want one latest tick", ticks)
	}
	if got, want := ticks[0].t, int64(102); got != want {
		t.Fatalf("fromChunk tick = %d, want %d", got, want)
	}
}

func TestScheduledTickQueueFromChunkOmitsSupersededTicks(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}
	hash := DefaultBlockRegistry.BlockHash(b)
	queue.ticks = []scheduledTick{
		{pos: pos, t: 101, b: b, bhash: hash},
		{pos: pos, t: 102, b: b, bhash: hash},
	}
	queue.furthestTicks[scheduledTickIndex{pos: pos, hash: hash}] = 102

	ticks := queue.fromChunk(chunkPosFromBlockPos(pos))
	if len(ticks) != 1 || ticks[0].t != 102 {
		t.Fatalf("fromChunk active ticks = %v, want only tick 102", ticks)
	}
}

func TestScheduledTickQueueIgnoresEarlierTickBehindLaterTick(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}

	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/10)
	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/20)

	index := scheduledTickIndex{pos: pos, hash: DefaultBlockRegistry.BlockHash(b)}
	if got, want := queue.furthestTicks[index], int64(102); got != want {
		t.Fatalf("furthest tick = %d, want %d", got, want)
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
	if len(queue.furthestTicks) != 0 {
		t.Fatalf("furthest ticks after removeChunk = %v, want empty", queue.furthestTicks)
	}
}

func TestScheduledTickQueueCanRescheduleWhileCurrentTickIsDue(t *testing.T) {
	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}
	index := scheduledTickIndex{pos: pos, hash: DefaultBlockRegistry.BlockHash(b)}
	queue.furthestTicks[index] = 100

	queue.schedule(DefaultBlockRegistry, pos, b, time.Second/2)
	if got, want := queue.furthestTicks[index], int64(110); got != want {
		t.Fatalf("rescheduled tick = %d, want %d", got, want)
	}
}

func TestScheduledTickQueueSkipsStaleDueTickBehindLaterTick(t *testing.T) {
	registry := scheduledTickTestRegistry()
	w := Config{Blocks: registry}.New()
	defer w.Close()

	queue := newScheduledTickQueue(100)
	pos := cube.Pos{8, 64, 8}
	b := scheduledTickTestBlock{}
	index := scheduledTickIndex{pos: pos, hash: registry.BlockHash(b)}

	ticks := 0
	scheduledTickTestBlockTicks = &ticks
	t.Cleanup(func() {
		scheduledTickTestBlockTicks = nil
	})

	var (
		ticksAfterFirst, ticksAfterSecond   int
		activeAfterFirst, activeAfterSecond []scheduledTick
		furthestAfterFirst                  int64
		hasFurthestAfterFirst               bool
	)
	<-w.Exec(func(tx *Tx) {
		tx.SetBlock(pos, b, nil)
		queue.schedule(registry, pos, b, time.Second/20)
		queue.schedule(registry, pos, b, time.Second/10)

		queue.tick(tx, 101)
		ticksAfterFirst = ticks
		activeAfterFirst = queue.fromChunk(chunkPosFromBlockPos(pos))
		furthestAfterFirst, hasFurthestAfterFirst = queue.furthestTicks[index]

		queue.tick(tx, 102)
		ticksAfterSecond = ticks
		activeAfterSecond = queue.fromChunk(chunkPosFromBlockPos(pos))
	})
	if ticksAfterFirst != 0 {
		t.Fatalf("stale due tick executed %d time(s), want 0", ticksAfterFirst)
	}
	if !hasFurthestAfterFirst || furthestAfterFirst != 102 {
		t.Fatalf("furthest tick after stale tick = %d, %t; want 102, true", furthestAfterFirst, hasFurthestAfterFirst)
	}
	if len(activeAfterFirst) != 1 || activeAfterFirst[0].t != 102 {
		t.Fatalf("active ticks after stale tick = %v, want only tick 102", activeAfterFirst)
	}
	if ticksAfterSecond != 1 {
		t.Fatalf("later scheduled tick executed %d time(s), want 1", ticksAfterSecond)
	}
	if len(activeAfterSecond) != 0 {
		t.Fatalf("active ticks after later tick = %v, want empty", activeAfterSecond)
	}
}

type scheduledTickTestBlock struct{}

var scheduledTickTestBlockTicks *int

func (scheduledTickTestBlock) ScheduledTick(cube.Pos, *Tx, *rand.Rand) {
	if scheduledTickTestBlockTicks != nil {
		(*scheduledTickTestBlockTicks)++
	}
}
func (scheduledTickTestBlock) EncodeBlock() (string, map[string]any) {
	return "test:scheduled_tick", nil
}
func (scheduledTickTestBlock) Hash() (uint64, uint64) { return 1 << 40, 0 }
func (scheduledTickTestBlock) Model() BlockModel      { return nil }

func scheduledTickTestRegistry() BlockRegistry {
	registry := NewBlockRegistry()
	registry.RegisterBlockState(BlockState{Name: "test:scheduled_tick", Properties: map[string]any{}})
	registry.RegisterBlock(scheduledTickTestBlock{})
	return registry
}

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

type redstoneStepRelayer struct{}

func (redstoneStepRelayer) RedstoneRelayerNeighbours(pos cube.Pos, _ *Tx) []cube.Pos {
	return []cube.Pos{
		{pos[0] + 1, pos[1] + 1, pos[2]},
		{pos[0] - 1, pos[1] - 1, pos[2]},
	}
}
func (redstoneStepRelayer) RedstoneSignalLoss(cube.Pos, *Tx, cube.Face, cube.Face) int {
	return 1
}
func (redstoneStepRelayer) EncodeBlock() (string, map[string]any) {
	return "test:redstone_step_relayer", nil
}
func (redstoneStepRelayer) Hash() (uint64, uint64) { return 1 << 42, 0 }
func (redstoneStepRelayer) Model() BlockModel      { return redstoneCancellationModel{} }

func redstoneCustomRelayerTestRegistry() BlockRegistry {
	registry := NewBlockRegistry()
	registry.RegisterBlockState(BlockState{Name: "test:redstone_step_relayer", Properties: map[string]any{}})
	registry.RegisterBlock(redstoneStepRelayer{})
	return registry
}

func redstoneNodeTestContains(nodes []redstoneNode, pos cube.Pos) bool {
	for _, node := range nodes {
		if node.pos == pos {
			return true
		}
	}
	return false
}

type redstoneCancellationHandler struct {
	NopHandler
	cancel map[cube.Pos]struct{}
}

func (h *redstoneCancellationHandler) HandleRedstoneUpdate(ctx *Context, update RedstoneUpdate) {
	if _, ok := h.cancel[update.Pos]; ok {
		ctx.Cancel()
	}
}

type redstoneRecordingHandler struct {
	NopHandler
	updates []RedstoneUpdate
}

func (h *redstoneRecordingHandler) HandleRedstoneUpdate(_ *Context, update RedstoneUpdate) {
	h.updates = append(h.updates, update)
}

var (
	redstoneCancellationConsumerUpdates *int
	redstoneCancellationActions         *int
)

func redstoneCancellationTestRegistry() BlockRegistry {
	registry := NewBlockRegistry()
	for _, power := range []int32{0, 15} {
		registry.RegisterBlockState(BlockState{Name: "test:redstone_source", Properties: map[string]any{"power": power}})
		registry.RegisterBlock(redstoneCancellationSource{Power: int(power)})
	}
	for _, powered := range []bool{false, true} {
		registry.RegisterBlockState(BlockState{Name: "test:redstone_consumer", Properties: map[string]any{"powered": powered}})
		registry.RegisterBlock(redstoneCancellationConsumer{Powered: powered})
	}
	registry.RegisterBlockState(BlockState{Name: "test:redstone_action", Properties: map[string]any{}})
	registry.RegisterBlock(redstoneCancellationAction{})
	return registry
}

type redstoneCancellationSource struct {
	Power int
}

func (b redstoneCancellationSource) RedstonePower(cube.Pos, *Tx, cube.Face) int {
	return b.Power
}
func (b redstoneCancellationSource) EncodeBlock() (string, map[string]any) {
	return "test:redstone_source", map[string]any{"power": int32(b.Power)}
}
func (b redstoneCancellationSource) Hash() (uint64, uint64) {
	return 1 << 43, uint64(b.Power)
}
func (redstoneCancellationSource) Model() BlockModel { return redstoneCancellationModel{} }

type redstoneCancellationConsumer struct {
	Powered bool
}

func (b redstoneCancellationConsumer) RedstonePowerUpdate(_ cube.Pos, _ *Tx, power int) (Block, bool) {
	if redstoneCancellationConsumerUpdates != nil {
		(*redstoneCancellationConsumerUpdates)++
	}
	powered := power > 0
	if b.Powered == powered {
		return b, false
	}
	b.Powered = powered
	return b, true
}
func (b redstoneCancellationConsumer) EncodeBlock() (string, map[string]any) {
	return "test:redstone_consumer", map[string]any{"powered": b.Powered}
}
func (b redstoneCancellationConsumer) Hash() (uint64, uint64) {
	if b.Powered {
		return 1 << 44, 1
	}
	return 1 << 44, 0
}
func (redstoneCancellationConsumer) Model() BlockModel { return redstoneCancellationModel{} }

type redstoneCancellationAction struct{}

func (redstoneCancellationAction) RedstonePowerAction(cube.Pos, *Tx, int, int) bool {
	if redstoneCancellationActions != nil {
		(*redstoneCancellationActions)++
	}
	return true
}
func (redstoneCancellationAction) EncodeBlock() (string, map[string]any) {
	return "test:redstone_action", nil
}
func (redstoneCancellationAction) Hash() (uint64, uint64) { return 1 << 45, 0 }
func (redstoneCancellationAction) Model() BlockModel      { return redstoneCancellationModel{} }

type redstoneCancellationModel struct{}

func (redstoneCancellationModel) BBox(cube.Pos, BlockSource) []cube.BBox { return nil }
func (redstoneCancellationModel) FaceSolid(cube.Pos, cube.Face, BlockSource) bool {
	return false
}

func redstoneSignalLossTestRegistry() BlockRegistry {
	registry := NewBlockRegistry()
	registry.RegisterBlockState(BlockState{Name: "test:redstone_loss_source", Properties: map[string]any{"power": int32(15)}})
	registry.RegisterBlock(redstoneLossSource{Power: 15})
	registry.RegisterBlockState(BlockState{Name: "test:redstone_loss_relayer", Properties: map[string]any{}})
	registry.RegisterBlock(redstoneLossRelayer{})
	for power := int32(0); power <= 15; power++ {
		registry.RegisterBlockState(BlockState{Name: "test:redstone_loss_consumer", Properties: map[string]any{"power": power}})
		registry.RegisterBlock(redstoneLossConsumer{Power: int(power)})
	}
	return registry
}

type redstoneLossSource struct {
	Power int
}

func (b redstoneLossSource) RedstonePower(cube.Pos, *Tx, cube.Face) int {
	return b.Power
}
func (b redstoneLossSource) EncodeBlock() (string, map[string]any) {
	return "test:redstone_loss_source", map[string]any{"power": int32(b.Power)}
}
func (b redstoneLossSource) Hash() (uint64, uint64) {
	return 1 << 45, uint64(b.Power)
}
func (redstoneLossSource) Model() BlockModel { return redstoneCancellationModel{} }

type redstoneLossRelayer struct{}

func (redstoneLossRelayer) RedstoneSignalLoss(cube.Pos, *Tx, cube.Face, cube.Face) int {
	return 1
}
func (redstoneLossRelayer) EncodeBlock() (string, map[string]any) {
	return "test:redstone_loss_relayer", nil
}
func (redstoneLossRelayer) Hash() (uint64, uint64) { return 1 << 46, 0 }
func (redstoneLossRelayer) Model() BlockModel      { return redstoneCancellationModel{} }

type redstoneLossConsumer struct {
	Power int
}

func (b redstoneLossConsumer) RedstonePowerUpdate(_ cube.Pos, _ *Tx, power int) (Block, bool) {
	if b.Power == power {
		return b, false
	}
	b.Power = power
	return b, true
}
func (b redstoneLossConsumer) EncodeBlock() (string, map[string]any) {
	return "test:redstone_loss_consumer", map[string]any{"power": int32(b.Power)}
}
func (b redstoneLossConsumer) Hash() (uint64, uint64) {
	return 1 << 47, uint64(b.Power)
}
func (redstoneLossConsumer) Model() BlockModel { return redstoneCancellationModel{} }

type redstoneNamedSolidBlock struct {
	name string
}

func (b redstoneNamedSolidBlock) EncodeBlock() (string, map[string]any) {
	return b.name, nil
}
func (redstoneNamedSolidBlock) Hash() (uint64, uint64) { return 1 << 48, 0 }
func (redstoneNamedSolidBlock) Model() BlockModel      { return redstoneSolidModel{} }

type redstoneSolidModel struct{}

func (redstoneSolidModel) BBox(cube.Pos, BlockSource) []cube.BBox { return nil }
func (redstoneSolidModel) FaceSolid(cube.Pos, cube.Face, BlockSource) bool {
	return true
}
