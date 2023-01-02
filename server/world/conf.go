package world

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

// Config may be used to create a new World. It holds a variety of fields that influence the World.
type Config struct {
	// Log is the Logger that will be used to log errors and debug messages to. If set to nil, a Logrus logger will be
	// used.
	Log Logger
	// Dim is the Dimension of the World. If set to nil, the World will use Overworld as its dimension. The dimension
	// set here influences, among others, the sky colour, weather/time and liquid behaviour in that World.
	Dim Dimension
	// PortalDestination is a function that returns the destination World for a portal of a specific Dimension type. If
	// set to nil, no portals will function. If the function returns a nil world for a Dimension, only portals of that
	// specific Dimension type will not function.
	PortalDestination func(dim Dimension) *World
	// Provider is the Provider implementation used to read and write World data. If set to nil, the Provider used will
	// be NopProvider, which does not store any data to disk.
	Provider Provider
	// Generator is the Generator implementation used to generate new areas of the World. If set to nil, the Provider
	// used will be NopGenerator, which generates completely empty chunks.
	Generator Generator
	// ReadOnly specifies if the World should be read-only, meaning no new data will be written to the Provider.
	ReadOnly bool
	// RandomTickSpeed specifies the rate at which blocks should be ticked in the World. By default, each sub chunk has
	// 3 blocks randomly ticked per sub chunk, so the default value is 3. Setting this value to -1 or lower will stop
	// random ticking altogether, while setting it higher results in faster ticking.
	RandomTickSpeed int
	// RandSource is the rand.Source used for generation of random numbers in a World, such as when selecting blocks to
	// tick or when deciding where to strike lightning. If set to nil, `rand.NewSource(time.Now().Unix())` will be used
	// to generate a new source.
	RandSource rand.Source
}

// Logger is a logger implementation that may be passed to the Log field of Config. World will send errors and debug
// messages to this Logger when appropriate.
type Logger interface {
	Errorf(format string, a ...any)
	Debugf(format string, a ...any)
}

// New creates a new World using the Config conf. The World returned will start ticking as soon as a viewer is added
// to it and is otherwise ready for use.
func (conf Config) New() *World {
	if conf.Log == nil {
		conf.Log = logrus.New()
	}
	if conf.Dim == nil {
		conf.Dim = Overworld
	}
	if conf.Provider == nil {
		conf.Provider = NopProvider{}
	}
	if conf.Generator == nil {
		conf.Generator = NopGenerator{}
	}
	if conf.RandomTickSpeed == 0 {
		conf.RandomTickSpeed = 3
	}
	if conf.RandSource == nil {
		conf.RandSource = rand.NewSource(time.Now().Unix())
	}
	s := conf.Provider.Settings()
	w := &World{
		scheduledUpdates: make(map[cube.Pos]int64),
		entities:         make(map[Entity]ChunkPos),
		viewers:          make(map[*Loader]Viewer),
		chunks:           make(map[ChunkPos]*chunkData),
		closing:          make(chan struct{}),
		handler:          *atomic.NewValue[Handler](NopHandler{}),
		r:                rand.New(conf.RandSource),
		advance:          s.ref.Inc() == 1,
		conf:             conf,
		ra:               conf.Dim.Range(),
		set:              s,
	}
	w.weather, w.ticker = weather{w: w}, ticker{w: w}

	go w.tickLoop()
	go w.chunkCacheJanitor()
	return w
}
