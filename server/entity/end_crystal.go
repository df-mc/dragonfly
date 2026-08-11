package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
)

// NewEndCrystal creates a new End crystal entity.
func NewEndCrystal(opts world.EntitySpawnOpts) *world.EntityHandle {
	return EndCrystalConfig{}.New(opts)
}

// EndCrystalConfig holds configuration for an End crystal entity.
type EndCrystalConfig struct {
	// ExplosionSize is the size of the explosion created when the End crystal is destroyed. It defaults to 6.
	ExplosionSize float64
	// ShowBase specifies whether the End crystal renders its bottom base.
	ShowBase bool
}

// New creates an End crystal entity with the configuration c.
func (c EndCrystalConfig) New(opts world.EntitySpawnOpts) *world.EntityHandle {
	return opts.New(EndCrystalType, c)
}

// Apply applies the End crystal configuration to data.
func (c EndCrystalConfig) Apply(data *world.EntityData) {
	if c.ExplosionSize == 0 {
		c.ExplosionSize = 6
	}
	data.Data = endCrystalBehaviour{
		showBase:      c.ShowBase,
		explosionSize: c.ExplosionSize,
	}
}

// EndCrystalType is a world.EntityType implementation for End crystals.
var EndCrystalType endCrystalType

type endCrystalType struct{}

func (endCrystalType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return Open(tx, handle, data)
}

func (endCrystalType) EncodeEntity() string {
	return "minecraft:ender_crystal"
}

func (endCrystalType) BBox(world.Entity) cube.BBox {
	return cube.Box(-1, 0, -1, 1, 2, 1)
}

func (endCrystalType) DecodeNBT(m map[string]any, data *world.EntityData) {
	b := endCrystalBehaviour{
		showBase:      nbtconv.Bool(m, "ShowBottom"),
		explosionSize: nbtconv.Float64(m, "ExplosionSize"),
	}
	if b.explosionSize == 0 {
		b.explosionSize = 6
	}
	x, hasX := m["BlockTargetX"].(int32)
	y, hasY := m["BlockTargetY"].(int32)
	z, hasZ := m["BlockTargetZ"].(int32)
	if hasX && hasY && hasZ {
		b.beamTarget = cube.Pos{int(x), int(y), int(z)}
		b.hasBeamTarget = true
	}
	b.Apply(data)
}

func (endCrystalType) EncodeNBT(data *world.EntityData) map[string]any {
	b := data.Data.(endCrystalBehaviour)
	m := map[string]any{
		"ShowBottom":    boolByte(b.showBase),
		"ExplosionSize": b.explosionSize,
	}
	if b.hasBeamTarget {
		m["BlockTargetX"] = int32(b.beamTarget[0])
		m["BlockTargetY"] = int32(b.beamTarget[1])
		m["BlockTargetZ"] = int32(b.beamTarget[2])
	}
	return m
}
