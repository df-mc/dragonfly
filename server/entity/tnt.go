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
	return newTNTWithSourceHandle(opts, fuse, nil, true)
}

// NewTNTWithSource creates a new primed TNT entity with the entity handle that caused it to ignite.
// The source is runtime-only and is not persisted through NBT reloads.
func NewTNTWithSource(opts world.EntitySpawnOpts, fuse time.Duration, source *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
	return newTNTWithSourceHandle(opts, fuse, source, blockableByShield)
}

func newTNTWithSourceHandle(opts world.EntitySpawnOpts, fuse time.Duration, source *world.EntityHandle, blockableByShield bool) *world.EntityHandle {
	if opts.Velocity.Len() == 0 {
		angle := rand.Float64() * math.Pi * 2
		opts.Velocity = mgl64.Vec3{-math.Sin(angle) * 0.02, 0.1, -math.Cos(angle) * 0.02}
	}
	return opts.New(TNTType, tntBehaviourConfig{Fuse: fuse, Source: source, UnblockableByShield: !blockableByShield})
}

var tntConf = PassiveBehaviourConfig{
	Gravity: 0.04,
	Drag:    0.02,
	Expire: func(e *Ent, tx *world.Tx) {
		explodeTNT(e, tx, nil, true)
	},
}

type tntBehaviourConfig struct {
	Fuse                time.Duration
	Source              *world.EntityHandle
	UnblockableByShield bool
}

func (conf tntBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

func (conf tntBehaviourConfig) New() *tntBehaviour {
	b := &tntBehaviour{source: conf.Source, blockableByShield: !conf.UnblockableByShield}
	confPassive := tntConf
	confPassive.ExistenceDuration = conf.Fuse
	confPassive.Expire = func(e *Ent, tx *world.Tx) {
		explodeTNT(e, tx, b.source, b.blockableByShield)
	}
	b.PassiveBehaviour = confPassive.New()
	return b
}

type tntBehaviour struct {
	*PassiveBehaviour
	source            *world.EntityHandle
	blockableByShield bool
}

// explodeTNT creates an explosion at the position of e.
func explodeTNT(e *Ent, tx *world.Tx, source *world.EntityHandle, blockableByShield bool) {
	tntExplosionConfig(tx, source, blockableByShield).Explode(tx, e.Position())
}

func tntExplosionConfig(tx *world.Tx, source *world.EntityHandle, blockableByShield bool) block.ExplosionConfig {
	var sourceEntity world.Entity
	if source != nil {
		sourceEntity, _ = source.Entity(tx)
	}
	return block.ExplosionConfig{ItemDropChance: 1, Source: sourceEntity, UnblockableByShield: !blockableByShield}
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
	data.Data = tntBehaviourConfig{
		Fuse:                nbtconv.TickDuration[uint8](m, "Fuse"),
		UnblockableByShield: nbtconv.Bool(m, "DragonflyUnblockableByShield"),
	}.New()
}

func (tntType) EncodeNBT(data *world.EntityData) map[string]any {
	fuse, blockableByShield := tntFuseAndBlockability(data.Data)
	m := map[string]any{"Fuse": uint8(fuse.Milliseconds() / 50)}
	if !blockableByShield {
		m["DragonflyUnblockableByShield"] = uint8(1)
	}
	return m
}

func tntFuseAndBlockability(data any) (time.Duration, bool) {
	switch b := data.(type) {
	case *tntBehaviour:
		return b.Fuse(), b.blockableByShield
	case *PassiveBehaviour:
		return b.Fuse(), true
	default:
		panic("invalid TNT behaviour type")
	}
}
