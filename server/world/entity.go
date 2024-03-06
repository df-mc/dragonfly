package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/maps"
	"io"
	"time"
)

// Entity represents an entity in the world, typically an object that may be moved around and can be
// interacted with by other entities.
// Viewers of a world may view an entity when near it.
type Entity interface {
	io.Closer

	// Type returns the EntityType of the Entity.
	Type() EntityType

	// Position returns the current position of the entity in the world.
	Position() mgl64.Vec3
	// Rotation returns the yaw and pitch of the entity in degrees. Yaw is horizontal rotation (rotation around the
	// vertical axis, 0 when facing forward), pitch is vertical rotation (rotation around the horizontal axis, also 0
	// when facing forward).
	Rotation() cube.Rotation
	// World returns the current world of the entity. This is always the world that the entity can actually be
	// found in.
	World() *World
}

// EntityType is the type of Entity. It specifies the name, encoded entity
// ID and bounding box of an Entity.
type EntityType interface {
	// EncodeEntity converts the entity to its encoded representation: It
	// returns the type of the Minecraft entity, for example
	// 'minecraft:falling_block'.
	EncodeEntity() string
	// BBox returns the bounding box of an Entity with this EntityType.
	BBox(e Entity) cube.BBox
}

// SaveableEntityType is an EntityType that may be saved to disk by decoding
// and encoding from/to NBT.
type SaveableEntityType interface {
	EntityType
	// DecodeNBT reads the fields from the NBT data map passed and converts it
	// to an Entity of the same EntityType.
	DecodeNBT(m map[string]any) Entity
	// EncodeNBT encodes the Entity of the same EntityType passed to a map of
	// properties that can be encoded to NBT.
	EncodeNBT(e Entity) map[string]any
}

// TickerEntity represents an entity that has a Tick method which should be called every time the entity is
// ticked every 20th of a second.
type TickerEntity interface {
	Entity
	// Tick ticks the entity with the current World and tick passed.
	Tick(w *World, current int64)
}

// EntityAction represents an action that may be performed by an entity. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type EntityAction interface {
	EntityAction()
}

// DamageSource represents the source of the damage dealt to an entity. This
// source may be passed to the Hurt() method of an entity in order to deal
// damage to an entity with a specific source.
type DamageSource interface {
	// ReducedByArmour checks if the source of damage may be reduced if the
	// receiver of the damage is wearing armour.
	ReducedByArmour() bool
	// ReducedByResistance specifies if the Source is affected by the resistance
	// effect. If false, damage dealt to an entity with this source will not be
	// lowered if the entity has the resistance effect.
	ReducedByResistance() bool
	// Fire specifies if the Source is fire related and should be ignored when
	// an entity has the fire resistance effect.
	Fire() bool
}

// HealingSource represents a source of healing for an entity. This source may
// be passed to the Heal() method of a living entity.
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
	Item               func(it any, pos, vel mgl64.Vec3) Entity
	FallingBlock       func(bl Block, pos mgl64.Vec3) Entity
	TNT                func(pos mgl64.Vec3, fuse time.Duration) Entity
	BottleOfEnchanting func(pos, vel mgl64.Vec3, owner Entity) Entity
	Arrow              func(pos, vel mgl64.Vec3, rot cube.Rotation, damage float64, owner Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip any) Entity
	Egg                func(pos, vel mgl64.Vec3, owner Entity) Entity
	EnderPearl         func(pos, vel mgl64.Vec3, owner Entity) Entity
	Firework           func(pos mgl64.Vec3, rot cube.Rotation, attached bool, firework Item, owner Entity) Entity
	LingeringPotion    func(pos, vel mgl64.Vec3, t any, owner Entity) Entity
	Snowball           func(pos, vel mgl64.Vec3, owner Entity) Entity
	SplashPotion       func(pos, vel mgl64.Vec3, t any, owner Entity) Entity
	Lightning          func(pos mgl64.Vec3) Entity
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
	return maps.Values(reg.ent)
}
