package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand/v2"
	"time"
)

// NewTNT creates a new primed TNT entity.
func NewTNT(opts world.EntitySpawnOpts, fuse time.Duration) *world.EntityHandle {
	return newTNTWithSourceHandle(opts, fuse, nil)
}

// NewTNTWithSource creates a new primed TNT entity with the entity that caused it to ignite.
func NewTNTWithSource(opts world.EntitySpawnOpts, fuse time.Duration, source world.Entity) *world.EntityHandle {
	return newTNTWithSourceHandle(opts, fuse, entityHandle(source))
}

func newTNTWithSourceHandle(opts world.EntitySpawnOpts, fuse time.Duration, source *world.EntityHandle) *world.EntityHandle {
	conf := tntConf
	conf.ExistenceDuration = fuse
	conf.Expire = func(e *Ent, tx *world.Tx) {
		explodeTNT(e, tx, source)
	}
	if opts.Velocity.Len() == 0 {
		angle := rand.Float64() * math.Pi * 2
		opts.Velocity = mgl64.Vec3{-math.Sin(angle) * 0.02, 0.1, -math.Cos(angle) * 0.02}
	}
	return opts.New(TNTType, conf)
}

func entityHandle(e world.Entity) *world.EntityHandle {
	if e == nil {
		return nil
	}
	return e.H()
}

var tntConf = PassiveBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
	Expire: func(e *Ent, tx *world.Tx) {
		explodeTNT(e, tx, nil)
	},
}

// explodeTNT creates an explosion at the position of e.
func explodeTNT(e *Ent, tx *world.Tx, source *world.EntityHandle) {
	sourceEntity, ok := source.Entity(tx)
	block.ExplosionConfig{ItemDropChance: 1, UnblockableByShield: !ok, Source: sourceEntity}.Explode(tx, e.Position())
}

// TNTType is a world.EntityType implementation for TNT.
var TNTType tntType

type tntType struct{}

func (t tntType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (tntType) EncodeEntity() string   { return "minecraft:tnt" }
func (tntType) NetworkOffset() float64 { return 0.49 }
func (tntType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.49, 0, -0.49, 0.49, 0.98, 0.49)
}

func (t tntType) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := tntConf
	conf.ExistenceDuration = nbtconv.TickDuration[uint8](m, "Fuse")
	data.Data = conf.New()
}

func (tntType) EncodeNBT(data *world.EntityData) map[string]any {
	return map[string]any{"Fuse": uint8(data.Data.(*PassiveBehaviour).Fuse().Milliseconds() / 50)}
}
