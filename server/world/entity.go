package world

import (
	"encoding/binary"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"io"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// EntityType is the type of Entity. It specifies the name, encoded Entity
// ID and bounding box of an Entity.
type EntityType interface {
	// Open returns an Entity implementation in the context of a transaction.
	Open(tx *Tx, handle *EntityHandle, data *EntityData) Entity

	// EncodeEntity converts the Entity to its encoded representation: It
	// returns the type of the Minecraft Entity, for example
	// 'minecraft:falling_block'.
	EncodeEntity() string
	// BBox returns the bounding box of an Entity with this EntityType.
	BBox(e Entity) cube.BBox
	// DecodeNBT reads the fields from the NBT data map passed and converts it
	// to an Entity of the same EntityType.
	DecodeNBT(m map[string]any, data *EntityData)
	// EncodeNBT encodes the Entity of the same EntityType passed to a map of
	// properties that can be encoded to NBT.
	EncodeNBT(data *EntityData) map[string]any
}

// EntityConfig is used to configure the initial settings of an Entity upon
// creation using NewEntity.
type EntityConfig interface {
	Apply(data *EntityData)
}

// EntityHandle is a persistent identifier of an entity. It holds data of the
// entity that can be transformed into an Entity implementation in the context
// of a transaction.
type EntityHandle struct {
	id uuid.UUID
	t  EntityType

	cond         *sync.Cond
	worldless    *atomic.Bool
	weakTxActive bool
	w            *World

	data EntityData

	// TODO Handler? Handle world change here?
}

// EntitySpawnOpts holds spawning related options for entities created.
type EntitySpawnOpts struct {
	// Position is the position that an Entity should be spawned at.
	Position mgl64.Vec3
	// Rotation is the rotation that an Entity should be spawned with.
	Rotation cube.Rotation
	// Velocity specifies the initial velocity of the Entity.
	Velocity mgl64.Vec3
	// ID specifies the UUID of an entity. This field should usually be left
	// empty, as a valid UUID is generated when not set. Non-player entities
	// only have the last 8 bytes of the UUID set.
	ID uuid.UUID
	// NameTag is the name tag that the entity is spawned with.
	NameTag string
}

// New creates an EntityHandle using an EntityType and EntityConfig passed. The
// EntityHandle may be added to a world by calling Tx.AddEntity().
// The spawn conditions depend on the options set in opts.
func (opts EntitySpawnOpts) New(t EntityType, conf EntityConfig) *EntityHandle {
	if opts.ID == uuid.Nil {
		// Generate a new UUID with only the upper 8 bytes filled. This UUID
		// needs to be translatable to an int64.
		opts.ID = uuid.New()
		clear(opts.ID[:8])
	}
	handle := &EntityHandle{id: opts.ID, t: t, cond: sync.NewCond(&sync.Mutex{}), worldless: &atomic.Bool{}}
	handle.worldless.Store(true)
	handle.data.Pos, handle.data.Rot, handle.data.Vel = opts.Position, opts.Rotation, opts.Velocity
	handle.data.Name = opts.NameTag
	conf.Apply(&handle.data)
	return handle
}

// NewEntity creates an EntityHandle using an EntityType and EntityConfig
// passed. The EntityHandle may be added to a world by calling Tx.AddEntity().
// NewEntity uses the zero value for EntitySpawnOpts.
func NewEntity(t EntityType, conf EntityConfig) *EntityHandle {
	var opts EntitySpawnOpts
	return opts.New(t, conf)
}

// entityFromData reads an entity from the decoded NBT data passed and returns
// an EntityHandle.
func entityFromData(t EntityType, id int64, data map[string]any) *EntityHandle {
	handle := &EntityHandle{t: t, cond: sync.NewCond(&sync.Mutex{}), worldless: &atomic.Bool{}}
	binary.LittleEndian.PutUint64(handle.id[8:], uint64(id))
	handle.decodeNBT(data)
	t.DecodeNBT(data, &handle.data)
	return handle
}

// Type returns the EntityType of the EntityHandle.
func (e *EntityHandle) Type() EntityType {
	return e.t
}

// Entity attempts to convert an EntityHandle to an Entity using the Tx passed.
// A non-nil Entity is returned only if the entity's world matches the world of
// the Tx. If they do not match, false is returned.
func (e *EntityHandle) Entity(tx *Tx) (Entity, bool) {
	if e == nil || e.w != tx.World() {
		return nil, false
	}
	return e.t.Open(tx, e, &e.data), true
}

// mustEntity calls Entity but panics if the worlds do not match.
func (e *EntityHandle) mustEntity(tx *Tx) Entity {
	if ent, ok := e.Entity(tx); ok {
		return ent
	}
	panic("can't load entity with Tx of different world")
}

// UUID returns the identifier of the EntityHandle.
func (e *EntityHandle) UUID() uuid.UUID {
	return e.id
}

// Close closes the EntityHandle. Any subsequent call to ExecWorld will return
// immediately without the transaction function being called. Close always
// returns nil.
func (e *EntityHandle) Close() error {
	e.setAndUnlockWorld(closeWorld)
	return nil
}

// ExecWorld obtains the EntityHandle's World in a thread-safe way and opens a
// transaction in it when it does. If the EntityHandle has not been added to a
// world, ExecWorld will block until the EntityHandle is added to a World and
// run the transaction function once it is. If the Entity is closed before
// ExecWorld is called, ExecWorld will return false immediately without running
// the transaction function.
func (e *EntityHandle) ExecWorld(f func(tx *Tx, e Entity)) bool {
	return e.execWorld(f, false)
}

// execWorld uses a sync.Cond to synchronise access to the handler's world. We
// are dealing with a rather complicated synchronisation pattern here. The goal
// for ExecWorld is to block until e.w becomes accessible. Meanwhile, World.Exec
// may also affect e.w, which execWorld needs to deal with.
func (e *EntityHandle) execWorld(f func(tx *Tx, e Entity), weak bool) bool {
	e.cond.L.Lock()
	for e.w == nil || (!weak && e.weakTxActive) {
		// Wait suspends the current goroutine and unlocks e.cond.L, until
		// e.cond.Broadcast() is called. After this, one of the goroutines
		// waiting will acquire a lock of e.cond.L again. This means that only
		// one goroutine will run the code after this simultaneously.
		e.cond.Wait()
	}
	// If a goroutine manages to exit the for loop, it will have acquired a lock
	// on e.cond.L. This also means that e.w can be assumed to not be nil here.
	// Because of the lock on e.cond.L, no other transaction will be able to
	// change e.w until we finish. e.worldless is set to true in
	// e.unsetAndLockWorld(), where the entity's world is removed.
	e.worldless.Store(false)
	if e.w == closeWorld {
		// EntityHandle was closed. No need to continue.
		e.cond.L.Unlock()
		return false
	}
	// We now arrive at the more complicated part. When we call e.w.Exec(), our
	// transaction must await earlier transactions in the world. If one of those
	// earlier transactions tries to change e.w (through e.unsetAndLockWorld()
	// or e.setAndUnlockWorld()), it must lock e.cond.L. This would lead to a
	// deadlock, because we already have e.cond.L locked here.
	// We work around this with so-called "weak transactions". This is a
	// transaction that may be invalidated before it is executed. In this case,
	// this invalidation happens by setting e.worldless to true. If the
	// transaction turns out to be invalidated (ret == false), we simply try
	// again, this time with e.execWorld(f, true) to make this goroutine bypass
	// any goroutines still awaiting e.cond.
	ret := e.weakExec(func(tx *Tx) { f(tx, e.mustEntity(tx)) })
	e.cond.L.Unlock()

	if !ret {
		// Our weak transaction was suspended. We try again, this time with
		// e.execWorld(f, true) to make this goroutine bypass any goroutines
		// still awaiting e.cond.
		return e.execWorld(f, true)
	}
	return true
}

// weakExec performs a "weak transaction". It adds a transaction to the world
// that is invalidated when e.worldless is set to true. In this case, weakExec
// returns false. If the weak transaction is successfully executed, it returns
// true, and any calls to ExecWorld waiting on e.cond are awakened. The goal of
// weakExec is to suspend the current goroutine and unlock e.cond.L while
// waiting for previous transactions to finish.
func (e *EntityHandle) weakExec(f ExecFunc) bool {
	e.weakTxActive = true

	// We create a weak transaction and start a for loop to listen for the
	// length of the channel. This might look weird, but the crucial part here
	// is the call to e.cond.Wait(), which unlocks e.cond.L. This is required
	// to prevent a deadlock if an earlier transaction tries to change e.w.
	c := e.w.weakExec(e.worldless, e.cond, f)
	for len(c) == 0 && e.w != closeWorld {
		// Calling e.cond.Wait() here will free the lock on e.cond.L until our
		// transaction finishes. e.w.weakExec() ensures that e.cond.Broadcast()
		// is called once the transaction finished/is suspended, so we can
		// continue after that.
		e.cond.Wait()
	}
	// If the EntityHandle was closed (e.w == closeWorld), we treat the
	// transaction as successful, because all transactions must be cancelled.
	if e.w != closeWorld && !<-c {
		// Weak transaction was suspended. Return false and try again.
		return false
	}
	// After setting e.weakTxActive back to false, we must Broadcast to make
	// sure any goroutines waiting in e.execWorld as a result of the
	// e.weakTxActive condition can continue.
	e.weakTxActive = false
	e.cond.Broadcast()
	return true
}

var closeWorld = &World{}

// unsetAndLockWorld sets e.w to nil, causing any subsequent calls to ExecWorld
// to block until e.w is set to a non-nil value.
func (e *EntityHandle) unsetAndLockWorld() {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()

	e.worldless.Store(true)
	e.w = nil
}

// setAndUnlockWorld sets e.w to a World passed and broadcasts e.cond, so that
// any goroutines waiting for a non-nil world are awoken.
func (e *EntityHandle) setAndUnlockWorld(w *World) {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()

	if e.w != nil {
		panic("cannot add entity to new world before removing from old world")
	}
	e.w = w
	e.cond.Broadcast()
}

// decodeNBT decodes the position, velocity, rotation, age, on-fire duration and
// name tag of an entity.
func (e *EntityHandle) decodeNBT(m map[string]any) {
	e.data.Pos = readVec3(m, "Pos")
	e.data.Vel = readVec3(m, "Motion")
	e.data.Rot = readRotation(m)
	e.data.Age = time.Duration(readInt16(m, "Age")) * (time.Second / 20)
	e.data.FireDuration = time.Duration(readInt16(m, "Fire")) * time.Second / 20
	e.data.Name, _ = m["NameTag"].(string)
}

// encodeNBT encodes the position, velocity, rotation, age, on-fire duration and
// name tag of an entity.
func (e *EntityHandle) encodeNBT() map[string]any {
	return map[string]any{
		"Pos":     []float32{float32(e.data.Pos[0]), float32(e.data.Pos[1]), float32(e.data.Pos[2])},
		"Motion":  []float32{float32(e.data.Vel[0]), float32(e.data.Vel[1]), float32(e.data.Vel[2])},
		"Yaw":     float32(e.data.Rot[0]),
		"Pitch":   float32(e.data.Rot[1]),
		"Fire":    int16(e.data.FireDuration.Seconds() * 20),
		"Age":     int16(e.data.Age / (time.Second * 20)),
		"NameTag": e.data.Name,
	}
}

// EntityData holds data shared by every entity. It is kept in an EntityHandle.
type EntityData struct {
	Pos, Vel     mgl64.Vec3
	Rot          cube.Rotation
	Name         string
	FireDuration time.Duration
	Age          time.Duration

	Data any
}

// Entity represents an Entity in the world, typically an object that may be moved around and can be
// interacted with by other entities.
// Viewers of a world may view an Entity when near it.
type Entity interface {
	io.Closer
	// H returns the EntityHandle that points to the entity.
	H() *EntityHandle
	// Position returns the current position of the Entity in the world.
	Position() mgl64.Vec3
	// Rotation returns the yaw (horizontal rotation) and pitch (vertical
	// rotation) of the entity in degrees.
	Rotation() cube.Rotation
}

// TickerEntity represents an Entity that has a Tick method which should be called every time the Entity is
// ticked every 20th of a second.
type TickerEntity interface {
	Entity
	// Tick ticks the Entity with the current World and tick passed.
	Tick(tx *Tx, current int64)
}

// EntityAction represents an action that may be performed by an Entity. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type EntityAction interface {
	EntityAction()
}

// DamageSource represents the source of the damage dealt to an Entity. This
// source may be passed to the Hurt() method of an Entity in order to deal
// damage to an Entity with a specific source.
type DamageSource interface {
	// ReducedByArmour checks if the source of damage may be reduced if the
	// receiver of the damage is wearing armour.
	ReducedByArmour() bool
	// ReducedByResistance specifies if the Source is affected by the resistance
	// effect. If false, damage dealt to an Entity with this source will not be
	// lowered if the Entity has the resistance effect.
	ReducedByResistance() bool
	// Fire specifies if the Source is fire related and should be ignored when
	// an Entity has the fire resistance effect.
	Fire() bool
	// IgnoreTotem specifies whether the totem will be ignored if the damage is lethal.
	IgnoreTotem() bool
}

// HealingSource represents a source of healing for an Entity. This source may
// be passed to the Heal() method of a living Entity.
type HealingSource interface {
	HealingSource()
}

// EntityRegistry is a mapping that EntityTypes may be registered to. It is used
// for loading entities from disk in a World's Provider.
type EntityRegistry struct {
	conf EntityRegistryConfig
	ent  map[string]EntityType
}

// EntityRegistryConfig holds functions used by the block and item packages to
// create entities as a result of their behaviour. ALL functions of
// EntityRegistryConfig must be filled out for the behaviour of these blocks and
// items not to fail.
type EntityRegistryConfig struct {
	Item               func(opts EntitySpawnOpts, it any) *EntityHandle
	FallingBlock       func(opts EntitySpawnOpts, bl Block) *EntityHandle
	TNT                func(opts EntitySpawnOpts, fuse time.Duration) *EntityHandle
	BottleOfEnchanting func(opts EntitySpawnOpts, owner Entity) *EntityHandle
	Arrow              func(opts EntitySpawnOpts, damage float64, owner Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip any) *EntityHandle
	Egg                func(opts EntitySpawnOpts, owner Entity) *EntityHandle
	EnderPearl         func(opts EntitySpawnOpts, owner Entity) *EntityHandle
	Firework           func(opts EntitySpawnOpts, firework Item, owner Entity, sidewaysVelocityMultiplier, upwardsAcceleration float64, attached bool) *EntityHandle
	LingeringPotion    func(opts EntitySpawnOpts, t any, owner Entity) *EntityHandle
	Snowball           func(opts EntitySpawnOpts, owner Entity) *EntityHandle
	SplashPotion       func(opts EntitySpawnOpts, t any, owner Entity) *EntityHandle
	Lightning          func(opts EntitySpawnOpts) *EntityHandle
}

// New creates an EntityRegistry using conf and the EntityTypes passed.
func (conf EntityRegistryConfig) New(ent []EntityType) EntityRegistry {
	m := make(map[string]EntityType, len(ent))
	for _, e := range ent {
		name := e.EncodeEntity()
		if _, ok := m[name]; ok {
			panic("cannot register the same entity (" + name + ") twice")
		}
		m[name] = e
	}
	return EntityRegistry{conf: conf, ent: m}
}

// Config returns the EntityRegistryConfig that was used to create the
// EntityRegistry.
func (reg EntityRegistry) Config() EntityRegistryConfig {
	return reg.conf
}

// Lookup looks up an EntityType by its name. If found, the EntityType is
// returned and the bool is true. The bool is false otherwise.
func (reg EntityRegistry) Lookup(name string) (EntityType, bool) {
	t, ok := reg.ent[name]
	return t, ok
}

// Types returns all EntityTypes passed upon construction of the EntityRegistry.
func (reg EntityRegistry) Types() []EntityType {
	return slices.Collect(maps.Values(reg.ent))
}

func readVec3(x map[string]any, k string) mgl64.Vec3 {
	if i, ok := x[k].([]any); ok {
		if len(i) != 3 {
			return mgl64.Vec3{}
		}
		var v mgl64.Vec3
		for index, f := range i {
			f32, _ := f.(float32)
			v[index] = float64(f32)
		}
		return v
	} else if i, ok := x[k].([]float32); ok {
		if len(i) != 3 {
			return mgl64.Vec3{}
		}
		return mgl64.Vec3{float64(i[0]), float64(i[1]), float64(i[2])}
	}
	return mgl64.Vec3{}
}

func readFloat32(m map[string]any, k string) float32 {
	v, _ := m[k].(float32)
	return v
}

func readRotation(m map[string]any) cube.Rotation {
	return cube.Rotation{float64(readFloat32(m, "Yaw")), float64(readFloat32(m, "Pitch"))}
}

func readInt16(m map[string]any, k string) int16 {
	v, _ := m[k].(int16)
	return v
}
