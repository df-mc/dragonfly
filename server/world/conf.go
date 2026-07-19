package world

import (
	"log/slog"
	"math/rand/v2"
	"time"
)

type blockRegistrySetter interface {
	// SetBlockRegistry updates the registry used by the provider to encode and decode blocks.
	// Config.New calls it with Config.Blocks after applying the default registry and finalizing it.
	SetBlockRegistry(BlockRegistry)
}

// Config may be used to create a new World. It holds a variety of fields that
// influence the World.
type Config struct {
	// Log is the Logger that will be used to log errors and debug messages to.
	// If set to nil, slog.Default() is set.
	Log *slog.Logger
	// Dim is the Dimension of the World. If set to nil, the World will use
	// Overworld as its dimension. The dimension set here influences, among
	// others, the sky colour, weather/time and liquid behaviour in that World.
	Dim Dimension
	// PortalDestination is a function that returns the destination World for a
	// portal of a specific Dimension type. If set to nil, no portals will
	// function. If the function returns a nil world for a Dimension, only
	// portals of that specific Dimension type will not function.
	PortalDestination func(dim Dimension) *World
	// Provider is the Provider implementation used to read and write World
	// data. If set to nil, the Provider used will be NopProvider, which does
	// not store any data to disk.
	Provider Provider
	// Generator is the Generator implementation used to generate new areas of
	// the World. If set to nil, the Generator used will be NopGenerator, which
	// generates completely empty chunks.
	Generator Generator
	// ReadOnly specifies if the World should be read-only, meaning no new data
	// will be written to the Provider.
	ReadOnly bool
	// SaveInterval specifies how often a World should be automatically saved to
	// disk. This includes chunks, entities and level.dat data. If ReadOnly is
	// set to false, changing SaveInterval will have no effect.
	// By default, SaveInterval is set to 10 minutes. Setting SaveInterval to
	// a negative number disables automatic saving entirely.
	SaveInterval time.Duration
	// ChunkUnloadInterval specifies how often unused chunks should be unloaded
	// from memory when no longer in use. By default, this is set to 2 minutes.
	// ChunkUnloadInterval should not be used to prevent chunks from unloading
	// altogether. This should be done using a Loader with a custom Viewer.
	ChunkUnloadInterval time.Duration
	// ChunkLoadWorkers is the number of background workers that load and
	// generate chunks, defaulting to 1. Higher values are faster, but require
	// the Generator to be safe for concurrent use.
	ChunkLoadWorkers int
	// RandomTickSpeed specifies the rate at which blocks should be ticked in
	// the World. By default, each sub chunk has 3 blocks randomly ticked per
	// sub chunk, so the default value is 3. Setting this value to -1 or lower
	// will stop random ticking altogether, while setting it higher results in
	// faster ticking.
	RandomTickSpeed int
	// RandSource is the rand.Source used for generation of random numbers in a
	// World, such as when selecting blocks to tick or when deciding where to
	// strike lightning. If set to nil, RandSource defaults to a `rand.PCG`
	// source seeded with `time.Now().UnixNano()`. PCG is significantly faster
	// than `rand.ChaCha8` on 64-bit systems at the expense of poorer
	// statistical distribution, which is acceptable here.
	// See https://go.dev/blog/chacha8rand.
	RandSource rand.Source
	// Entities is an EntityRegistry with all Entity types registered that may
	// be added to the World.
	Entities EntityRegistry

	// Blocks is the BlockRegistry used by the World.
	// If left nil, DefaultBlockRegistry is used. For a non-default registry,
	// use NewBlockRegistry(), register blocks/states, and call Finalize().
	Blocks BlockRegistry

	// Synchronous removes the World's own background goroutines. Immediate tasks
	// from World.Do and Call run on the calling goroutine, the World is not saved
	// or unloaded automatically, and time only passes on explicit
	// World.AdvanceTick calls. World.DoAfter and entity work scheduled before an
	// entity enters a world still use background goroutines and wall-clock
	// delays; callers must synchronise on the returned Task. This makes
	// Synchronous Worlds well suited to unit tests that need a World to interact
	// with.
	// A Synchronous World must be driven from one goroutine. Do, Call and
	// AdvanceTick are not safe to call concurrently, including from delayed
	// item or death callbacks.
	Synchronous bool
}

// New creates a new World using the Config conf. The World returned will start
// ticking as soon as a viewer is added to it and is otherwise ready for use.
func (conf Config) New() *World {
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	if conf.Dim == nil {
		conf.Dim = Overworld
	}
	if conf.SaveInterval == 0 {
		conf.SaveInterval = time.Minute * 10
	}
	if conf.ChunkUnloadInterval <= 0 {
		conf.ChunkUnloadInterval = time.Minute * 2
	}
	if conf.ChunkLoadWorkers <= 0 {
		conf.ChunkLoadWorkers = defaultChunkLoadWorkers
	}
	if conf.Generator == nil {
		conf.Generator = NopGenerator{}
	}
	if conf.Provider == nil {
		// If no provider is set, use the default settings and the default spawn position from the generator.
		s := defaultSettings()
		s.Spawn = conf.Generator.DefaultSpawn(conf.Dim)
		conf.Provider = NopProvider{Set: s}
	}
	if conf.RandomTickSpeed == 0 {
		conf.RandomTickSpeed = 3
	}
	if conf.Blocks == nil {
		conf.Blocks = DefaultBlockRegistry
	}

	// Initialize the passed block registry and also initialize the default block registry which
	// is used in some vanilla paths.
	conf.Blocks.Finalize()
	DefaultBlockRegistry.Finalize()
	if provider, ok := conf.Provider.(blockRegistrySetter); ok {
		provider.SetBlockRegistry(conf.Blocks)
	}

	if conf.RandSource == nil {
		t := uint64(time.Now().UnixNano())
		conf.RandSource = rand.NewPCG(t, t)
	}
	s := conf.Provider.Settings()

	// The Provider is shared between the owner and the chunk load workers, so
	// serialise calls to it. A single chunk load worker also keeps the
	// Generator serialised.
	conf.Provider = &lockedProvider{p: conf.Provider}
	if conf.ChunkLoadWorkers == 1 {
		conf.Generator = &lockedGenerator{g: conf.Generator}
	}
	w := &World{
		scheduledUpdates: newScheduledTickQueue(s.CurrentTick),
		redstone:         newRedstoneEngine(s.CurrentTick),
		entities:         make(map[*EntityHandle]ChunkPos),
		viewers:          make(map[*Loader]Viewer),
		chunks:           make(map[ChunkPos]*Column),
		chunkRequests:    make(map[ChunkPos]*chunkRequest),
		queueClosing:     make(chan struct{}),
		closeStarted:     make(chan struct{}),
		closing:          make(chan struct{}),
		queue:            make(chan transaction, 128),
		r:                rand.New(conf.RandSource),
		advance:          s.ref.Add(1) == 1,
		conf:             conf,
		ra:               conf.Dim.Range(),
		set:              s,
	}
	w.chunkWorkers = newChunkWorkerPool(w)
	w.weather = weather{w: w}
	var h Handler = NopHandler{}
	w.handler.Store(&h)

	t := ticker{interval: time.Second / 20}
	if !conf.Synchronous {
		w.queueing.Add(1)
		w.running.Add(2)

		go t.tickLoop(w)
		go w.autoSave()
		go w.handleTransactions()
		w.chunkWorkers.wg.Add(conf.ChunkLoadWorkers)
		for range conf.ChunkLoadWorkers {
			go w.chunkWorkers.handle()
		}
	}

	<-w.exec(t.tick)
	return w
}
