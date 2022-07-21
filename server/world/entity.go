package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
	"io"
)

// Entity represents an entity in the world, typically an object that may be moved around and can be
// interacted with by other entities.
// Viewers of a world may view an entity when near it.
type Entity interface {
	io.Closer

	// Name returns a human-readable name for the entity. This is not unique for an entity, but generally
	// unique for an entity type.
	Name() string
	// EncodeEntity converts the entity to its encoded representation: It returns the type of the Minecraft
	// entity, for example 'minecraft:falling_block'.
	EncodeEntity() string

	// BBox returns the bounding box of the Entity.
	BBox() cube.BBox

	// Position returns the current position of the entity in the world.
	Position() mgl64.Vec3
	// Rotation returns the yaw and pitch of the entity in degrees. Yaw is horizontal rotation (rotation around the
	// vertical axis, 0 when facing forward), pitch is vertical rotation (rotation around the horizontal axis, also 0
	// when facing forward).
	Rotation() (yaw, pitch float64)
	// World returns the current world of the entity. This is always the world that the entity can actually be
	// found in.
	World() *World
}

// TickerEntity represents an entity that has a Tick method which should be called every time the entity is
// ticked every 20th of a second.
type TickerEntity interface {
	// Tick ticks the entity with the current World and tick passed.
	Tick(w *World, current int64)
}

// SaveableEntity is an Entity that can be saved and loaded with the World it was added to. These entities can be
// registered on startup using RegisterEntity to allow loading them in a World.
type SaveableEntity interface {
	Entity
	NBTer
}

// entities holds a map of name => SaveableEntity to be used for looking up the entity by a string ID. It is registered
// to when calling RegisterEntity.
var entities = map[string]SaveableEntity{}

// RegisterEntity registers a SaveableEntity to the map so that it can be saved and loaded with the world.
func RegisterEntity(e SaveableEntity) {
	name := e.EncodeEntity()
	if _, ok := entities[name]; ok {
		panic("cannot register the same entity (" + name + ") twice")
	}
	entities[name] = e
}

// EntityByName looks up a SaveableEntity by the name (for example, 'minecraft:slime') and returns it if found.
// EntityByName can only return entities previously registered using RegisterEntity. If not found, the bool returned is
// false.
func EntityByName(name string) (SaveableEntity, bool) {
	e, ok := entities[name]
	return e, ok
}

// Entities returns all registered entities.
func Entities() []SaveableEntity {
	es := make([]SaveableEntity, 0, len(entities))
	for _, e := range entities {
		es = append(es, e)
	}
	return es
}

// EntityAction represents an action that may be performed by an entity. Typically, these actions are sent to
// viewers in a world so that they can see these actions.
type EntityAction interface {
	EntityAction()
}
