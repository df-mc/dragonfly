package world

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"io"
	"log/slog"
	"slices"
	"sync"
)

// TestSynchronousWorldDo verifies that Do on a synchronous World runs the task
// on the calling goroutine and returns a completed task.
func TestSynchronousWorldDo(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	var ran bool
	task := w.Do(func(tx *Tx) { ran = true })
	if !ran {
		t.Fatal("expected task to have run when Do returned")
	}
	select {
	case <-task.Done():
	default:
		t.Fatal("expected task returned by Do to be done when Do returned")
	}
}

// TestSynchronousWorldAdvanceTick verifies that a synchronous World does not
// tick on its own and that AdvanceTick advances the current tick exactly once
// per call, even without any viewers.
func TestSynchronousWorldAdvanceTick(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	current := func() int64 {
		w.set.Lock()
		defer w.set.Unlock()
		return w.set.CurrentTick
	}
	start := current()
	time.Sleep(time.Second / 10)
	if got := current(); got != start {
		t.Fatalf("expected no automatic ticking, tick advanced from %v to %v", start, got)
	}
	for range 5 {
		w.AdvanceTick()
	}
	if got := current(); got != start+5 {
		t.Fatalf("expected current tick %v after 5 AdvanceTick calls, got %v", start+5, got)
	}
}

func TestSynchronousEntityDoCanRemoveEntity(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	h := EntitySpawnOpts{Position: mgl64.Vec3{0, 4, 0}}.New(testEntityType{}, testEntityConfig{})
	<-w.exec(func(tx *Tx) {
		tx.AddEntity(h)
	})

	task := h.Do(func(tx *Tx, e Entity) {
		tx.RemoveEntity(e)
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := task.Wait(ctx); err != nil {
		t.Fatalf("entity Do self-removal did not complete: %v", err)
	}
}

func TestSynchronousEntityDoWaitsForAddEntityToFinish(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	state := &blockingOpenState{
		firstOpen:  make(chan struct{}),
		secondOpen: make(chan struct{}),
		release:    make(chan struct{}),
	}
	h := EntitySpawnOpts{}.New(blockingOpenType{}, blockingOpenConfig{state: state})
	task := h.Do(func(*Tx, Entity) {})
	added := make(chan struct{})
	go func() {
		w.Do(func(tx *Tx) { tx.AddEntity(h) })
		close(added)
	}()
	<-state.firstOpen

	premature := false
	select {
	case <-state.secondOpen:
		premature = true
	case <-time.After(time.Millisecond * 50):
	}
	close(state.release)
	<-added
	if err := task.Wait(context.Background()); err != nil {
		t.Fatalf("entity Do failed: %v", err)
	}
	if premature {
		t.Fatal("entity callback opened before AddEntity completed")
	}
}

func TestSynchronousAdvanceTickTicksViewerlessEntities(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	h := EntitySpawnOpts{Position: mgl64.Vec3{0, 4, 0}}.New(testEntityType{}, testEntityConfig{})
	<-w.exec(func(tx *Tx) {
		tx.AddEntity(h)
	})

	start := h.data.Pos
	for range 3 {
		w.AdvanceTick()
	}
	if got := h.data.Pos; got == start {
		t.Fatalf("expected entity position to change after ticking, got %v", got)
	}
}

func TestSynchronousAdvanceTickTicksViewerlessBlockEntities(t *testing.T) {
	w := Config{Synchronous: true}.New()
	defer w.Close()

	pos := cube.Pos{0, 4, 0}
	tb := &testTickerBlock{}
	<-w.exec(func(tx *Tx) {
		col := tx.chunk(chunkPosFromBlockPos(pos))
		chest, ok := tx.World().conf.Blocks.BlockByName("minecraft:chest", map[string]any{"minecraft:cardinal_direction": "north"})
		if !ok {
			t.Fatal("expected chest block to be registered")
		}
		col.SetBlock(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0, tx.World().conf.Blocks.BlockRuntimeID(chest))
		col.BlockEntities[pos] = tb
	})

	w.AdvanceTick()
	if tb.ticks == 0 {
		t.Fatal("expected block entity to tick")
	}
}

type testEntityConfig struct{}

func (testEntityConfig) Apply(*EntityData) {}

type testEntityType struct{}

func (testEntityType) Open(_ *Tx, handle *EntityHandle, data *EntityData) Entity {
	return &testEntity{handle: handle, data: data}
}

func (testEntityType) EncodeEntity() string {
	return "dragonfly:test_entity"
}

func (testEntityType) BBox(Entity) cube.BBox {
	return cube.Box(0, 0, 0, 1, 1, 1)
}

func (testEntityType) DecodeNBT(map[string]any, *EntityData) {}

func (testEntityType) EncodeNBT(*EntityData) map[string]any {
	return nil
}

type testEntity struct {
	handle *EntityHandle
	data   *EntityData
}

func (e *testEntity) Close() error {
	return nil
}

func (e *testEntity) H() *EntityHandle {
	return e.handle
}

func (e *testEntity) Position() mgl64.Vec3 {
	return e.data.Pos
}

func (e *testEntity) Rotation() cube.Rotation {
	return e.data.Rot
}

func (e *testEntity) Tick(*Tx, int64) {
	e.data.Pos = e.data.Pos.Add(mgl64.Vec3{0, -0.1, 0})
}

type testTickerBlock struct {
	ticks int
}

type blockingOpenState struct {
	opens      atomic.Int32
	firstOpen  chan struct{}
	secondOpen chan struct{}
	release    chan struct{}
}

type blockingOpenConfig struct {
	state *blockingOpenState
}

func (c blockingOpenConfig) Apply(data *EntityData) { data.Data = c.state }

type blockingOpenType struct{}

func (blockingOpenType) Open(_ *Tx, handle *EntityHandle, data *EntityData) Entity {
	state := data.Data.(*blockingOpenState)
	switch state.opens.Add(1) {
	case 1:
		close(state.firstOpen)
		<-state.release
	case 2:
		close(state.secondOpen)
	}
	return &testEntity{handle: handle, data: data}
}

func (blockingOpenType) EncodeEntity() string { return "dragonfly:blocking_open" }

func (blockingOpenType) BBox(Entity) cube.BBox { return cube.BBox{} }

func (blockingOpenType) DecodeNBT(map[string]any, *EntityData) {}

func (blockingOpenType) EncodeNBT(*EntityData) map[string]any { return nil }

func (*testTickerBlock) EncodeBlock() (string, map[string]any) {
	return "dragonfly:test_ticker", nil
}

func (*testTickerBlock) Hash() (uint64, uint64) {
	return 1<<32 - 1, 0
}

func (*testTickerBlock) Model() BlockModel {
	return unknownModel{}
}

func (*testTickerBlock) DecodeNBT(map[string]any) any {
	return &testTickerBlock{}
}

func (*testTickerBlock) EncodeNBT() map[string]any {
	return nil
}

func (b *testTickerBlock) Tick(int64, cube.Pos, *Tx) {
	b.ticks++
}

// saveChest mimics block.Chest: the block value in Column.BlockEntities holds a
// pointer to its contents (block.Chest holds a *inventory.Inventory), so item
// moves mutate the container in place without ever going back through
// Tx.SetBlock or Tx.SetBlockEntity.
type saveChest struct {
	items *int
}

func (saveChest) EncodeBlock() (string, map[string]any) {
	return "minecraft:chest", map[string]any{"minecraft:cardinal_direction": "north"}
}
func (saveChest) Hash() (uint64, uint64) { return 1 << 40, 0 }
func (saveChest) Model() BlockModel      { return unknownModel{} }

func (c saveChest) EncodeNBT() map[string]any {
	return map[string]any{"Items": int32(*c.items)}
}

func (saveChest) DecodeNBT(data map[string]any) any {
	n := int(data["Items"].(int32))
	return saveChest{items: &n}
}

// saveChestRegistry builds a registry in which minecraft:chest is backed by
// saveChest, so the world treats it as an NBT block entity.
func saveChestRegistry() BlockRegistry {
	br := NewBlockRegistry()
	n := 0
	br.(*BasicBlockRegistry).RegisterBlock(saveChest{items: &n})
	br.Finalize()
	return br
}

// TestSaveChunkStoresContainerContents reproduces the "chunk save
// stating" family: emptying a container in a chunk that is otherwise untouched
// never sets Column.modified, so saveChunk skips the chunk and the on-disk copy
// keeps the items the player already walked away with.
func TestSaveChunkStoresContainerContents(t *testing.T) {
	br := saveChestRegistry()
	p := newSaveProvider(Overworld.Range())
	p.blocks = br

	w := Config{
		Log:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		Synchronous: true,
		Provider:    p,
		Entities:    saveTestRegistry(),
		Blocks:      br,
	}.New()

	chestPos := cube.Pos{8, 40, 8}
	cp := chunkPosFromBlockPos(chestPos)

	// Session 1: a chest with 64 items is placed and saved.
	stored := 64
	runWorld(w, func(tx *Tx) {
		tx.SetBlock(chestPos, saveChest{items: &stored}, nil)
	})
	w.Save()
	unloadAll(w)
	t.Logf("on disk after placing: chest holds %d items", diskChestItems(t, p, cp, chestPos))

	// Session 2: the chunk is reloaded and is therefore unmodified. A player
	// walks in from a neighbouring chunk and empties the chest. Walking does
	// not mark a chunk modified (tickEntities never sets the flag), and neither
	// does mutating the container's inventory.
	var taken int
	runWorld(w, func(tx *Tx) {
		c := tx.chunk(cp)
		if c.modified {
			t.Fatal("expected reloaded chunk to be unmodified")
		}
		ch := c.BlockEntities[chestPos].(saveChest)
		if *ch.items != 64 {
			t.Fatalf("expected 64 items in reloaded chest, got %d", *ch.items)
		}
		// Player takes everything out; this is exactly what
		// inventory.Inventory does to block.Chest's shared inventory pointer.
		taken, *ch.items = *ch.items, 0
		t.Logf("player took %d items, chest now holds %d, chunk.modified=%v", taken, *ch.items, c.modified)
	})

	w.Save()
	unloadAll(w)

	onDisk := diskChestItems(t, p, cp, chestPos)
	t.Logf("on disk after emptying the chest: chest holds %d items", onDisk)

	// Session 3: reload and count what actually exists.
	var reloaded int
	runWorld(w, func(tx *Tx) {
		reloaded = *tx.chunk(cp).BlockEntities[chestPos].(saveChest).items
	})
	total := taken + reloaded
	t.Logf("items in player inventory: %d, items back in chest after reload: %d, total: %d (expected 64)", taken, reloaded, total)
	_ = w.Close()

	if total != 64 {
		t.Fatalf("container contents duplicated: player kept %d items and the chest still holds %d after reload (total %d, expected 64)", taken, reloaded, total)
	}
}

// diskChestItems reads the item count the provider currently has stored for the
// chest at pos.
func diskChestItems(t *testing.T, p *saveProvider, cp ChunkPos, pos cube.Pos) int {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, be := range p.cols[cp].blockEntities {
		if be.Pos == pos {
			return int(be.Data["Items"].(int32))
		}
	}
	return -1
}

var _ = mgl64.Vec3{}

// saveStoredColumn is a fully serialised column, mirroring what mcdb writes to
// LevelDB. Nothing here aliases live world state.
type saveStoredColumn struct {
	data          chunk.SerialisedData
	entities      []chunk.Entity
	blockEntities []chunk.BlockEntity
	scheduled     []chunk.ScheduledBlockUpdate
	tick          int64
}

// saveProvider is an in-memory Provider that serialises columns the exact same
// way mcdb does (chunk.Encode + NBT round trip), so it behaves like a real disk
// provider without the import cycle that using mcdb from package world would
// cause.
type saveProvider struct {
	mu      sync.Mutex
	set     *Settings
	cols    map[ChunkPos]saveStoredColumn
	spawns  map[uuid.UUID]cube.Pos
	blocks  BlockRegistry
	rng     cube.Range
	stores  int
	closedN int
}

func newSaveProvider(r cube.Range) *saveProvider {
	br := DefaultBlockRegistry
	br.Finalize()
	return &saveProvider{
		set:    defaultSettings(),
		cols:   make(map[ChunkPos]saveStoredColumn),
		spawns: make(map[uuid.UUID]cube.Pos),
		blocks: br,
		rng:    r,
	}
}

func (p *saveProvider) Settings() *Settings      { return p.set }
func (p *saveProvider) SaveSettings(s *Settings) {}
func (p *saveProvider) LoadPlayerSpawnPosition(id uuid.UUID) (cube.Pos, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pos, ok := p.spawns[id]
	return pos, ok, nil
}
func (p *saveProvider) SavePlayerSpawnPosition(id uuid.UUID, pos cube.Pos) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spawns[id] = pos
	return nil
}
func (p *saveProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closedN++
	return nil
}

// copyNBT round trips an NBT map through the same encoding mcdb uses.
func copyNBT(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	b, err := nbt.MarshalEncoding(m, nbt.LittleEndian)
	if err != nil {
		panic(err)
	}
	out := make(map[string]any)
	if err := nbt.UnmarshalEncoding(b, &out, nbt.LittleEndian); err != nil {
		panic(err)
	}
	return out
}

func (p *saveProvider) StoreColumn(pos ChunkPos, _ Dimension, col *chunk.Column) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stores++
	s := saveStoredColumn{
		data:      chunk.Encode(col.Chunk, chunk.DiskEncoding),
		tick:      col.Tick,
		scheduled: slices.Clone(col.ScheduledBlocks),
	}
	for _, e := range col.Entities {
		s.entities = append(s.entities, chunk.Entity{ID: e.ID, Data: copyNBT(e.Data)})
	}
	for _, be := range col.BlockEntities {
		s.blockEntities = append(s.blockEntities, chunk.BlockEntity{Pos: be.Pos, Data: copyNBT(be.Data)})
	}
	p.cols[pos] = s
	return nil
}

func (p *saveProvider) LoadColumn(pos ChunkPos, _ Dimension) (*chunk.Column, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.cols[pos]
	if !ok {
		return nil, leveldb.ErrNotFound
	}
	c, err := chunk.DiskDecode(p.blocks, s.data, p.rng)
	if err != nil {
		return nil, err
	}
	col := &chunk.Column{Chunk: c, Tick: s.tick, ScheduledBlocks: slices.Clone(s.scheduled)}
	for _, e := range s.entities {
		col.Entities = append(col.Entities, chunk.Entity{ID: e.ID, Data: copyNBT(e.Data)})
	}
	for _, be := range s.blockEntities {
		col.BlockEntities = append(col.BlockEntities, chunk.BlockEntity{Pos: be.Pos, Data: copyNBT(be.Data)})
	}
	return col, nil
}

// storedEntities returns the number of entities stored on "disk" for pos.
func (p *saveProvider) storedEntities(pos ChunkPos) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.cols[pos].entities)
}

// saveTestRegistry is an EntityRegistry with the test entity type registered so
// entities survive a save/load round trip.
func saveTestRegistry() EntityRegistry {
	return EntityRegistryConfig{}.New([]EntityType{testEntityType{}})
}

func saveSyncWorld(t *testing.T, p *saveProvider) *World {
	t.Helper()
	return Config{
		Log:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		Synchronous: true,
		Provider:    p,
		Entities:    saveTestRegistry(),
	}.New()
}

// unloadAll force-unloads every chunk currently in memory, exactly as the
// chunk unload ticker does.
func unloadAll(w *World) {
	runWorld(w, func(tx *Tx) {
		w.closeUnusedChunks(tx)
	})
}

// TestSaveChunkStoresEntityAfterItMoves reproduces the classic
// "dupe principle" case: the entity is recorded in its new chunk before the old
// chunk's on-disk copy is cleared. tickEntities moves the handle between
// Column.Entities slices without setting Column.modified on either chunk, so
// the source chunk is never re-saved and keeps its stale copy of the entity,
// while the destination chunk is saved with the same entity in it.
func TestSaveChunkStoresEntityAfterItMoves(t *testing.T) {
	p := newSaveProvider(Overworld.Range())
	w := saveSyncWorld(t, p)

	src, dst := ChunkPos{0, 0}, ChunkPos{1, 0}

	// Session 1: spawn one entity inside the source chunk and persist it.
	h := EntitySpawnOpts{Position: mgl64.Vec3{8, 40, 8}}.New(testEntityType{}, testEntityConfig{})
	runWorld(w, func(tx *Tx) { tx.AddEntity(h) })
	w.Save()
	unloadAll(w)

	t.Logf("after initial save: src=%d dst=%d entities on disk", p.storedEntities(src), p.storedEntities(dst))

	// Session 2: reload the source chunk from disk. The freshly loaded Column
	// has modified == false.
	var moved *EntityHandle
	runWorld(w, func(tx *Tx) {
		c := tx.chunk(src)
		if len(c.Entities) != 1 {
			t.Fatalf("expected 1 entity in reloaded source chunk, got %d", len(c.Entities))
		}
		if c.modified {
			t.Fatal("expected freshly loaded chunk to be unmodified")
		}
		moved = c.Entities[0]

		// Anything happening in the destination chunk marks it modified. Here,
		// a second item simply drops there.
		tx.AddEntity(EntitySpawnOpts{Position: mgl64.Vec3{20, 40, 8}}.New(testEntityType{}, testEntityConfig{}))

		// Walk the entity over the chunk border.
		moved.data.Pos = mgl64.Vec3{20, 40, 8}
	})

	// A single tick is all it takes for tickEntities to move the handle.
	w.AdvanceTick()

	runWorld(w, func(tx *Tx) {
		if got := w.entities[moved]; got != dst {
			t.Fatalf("expected entity to be tracked in %v, got %v", dst, got)
		}
		if c := w.chunks[src]; len(c.Entities) != 0 {
			t.Fatalf("expected source chunk to have 0 entities in memory, got %d", len(c.Entities))
		}
		if c := w.chunks[dst]; len(c.Entities) != 2 {
			t.Fatalf("expected destination chunk to have 2 entities in memory, got %d", len(c.Entities))
		}
		t.Logf("in memory after move: src=%d dst=%d, src.modified=%v dst.modified=%v",
			len(w.chunks[src].Entities), len(w.chunks[dst].Entities),
			w.chunks[src].modified, w.chunks[dst].modified)
	})

	w.Save()
	unloadAll(w)

	srcOnDisk, dstOnDisk := p.storedEntities(src), p.storedEntities(dst)
	t.Logf("after move+save: src=%d dst=%d entities on disk (total %d)", srcOnDisk, dstOnDisk, srcOnDisk+dstOnDisk)

	// Session 3: reload both chunks and count the entities they actually
	// contain. Counting Column.Entities rather than World.entities keeps this
	// independent of the known handle leak in closeChunk.
	var total int
	runWorld(w, func(tx *Tx) {
		total = len(tx.chunk(src).Entities) + len(tx.chunk(dst).Entities)
	})
	t.Logf("entities alive after reload: %d (expected 2)", total)
	_ = w.Close()

	if total != 2 {
		t.Fatalf("entity duplicated across chunk border: expected 2 entities after reload, got %d (src chunk on disk: %d, dst chunk on disk: %d)", total, srcOnDisk, dstOnDisk)
	}
}
