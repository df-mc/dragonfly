package entity

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

const behaviourNBTKey = "DragonflyBehaviour"

// PersistentBehaviour is a Behaviour with versioned state that survives world
// saves. LoadBehaviourNBT must accept every version the implementation has
// previously saved. Stateless behaviours need not implement this interface;
// mutable state may alternatively live in persistent world components. Call
// Ent.MarkBehaviourDirty after mutating a persistent behaviour.
type PersistentBehaviour interface {
	Behaviour
	SaveBehaviourNBT() (version int32, data map[string]any)
	LoadBehaviourNBT(version int32, data map[string]any)
}

// ActorSpec describes how a Spec entity is spawned to Bedrock clients. Spec
// entities use AddActor; player and item actor packet families are not
// supported. Client metadata is supplied by MetaSyncer components.
type ActorSpec struct {
	// ID is the client actor identifier. It defaults to Spec.Name.
	ID string
	// Offset is added to the entity's Y coordinate in movement packets.
	Offset float64
}

// Spec describes a component-based entity type.
type Spec struct {
	// Name is the namespaced ID used for registration and saves.
	Name string
	// Actor describes the AddActor representation sent to clients.
	Actor ActorSpec
	// Box is the entity's bounding box.
	Box cube.BBox
	// Components returns fresh components for each entity. A Behaviour becomes
	// the entity's main behaviour and is not persisted as a component.
	Components func() []any
}

// Type implements world.EntityType for a Spec. Register it in the world's
// EntityRegistry to load saved entities of this type.
type Type struct {
	spec Spec
}

// RegisterType validates a Spec and adds its type to registry. Register its
// component types first, and call RegisterType before passing registry to a
// world. It panics if the Spec is invalid or its save ID is already present in
// registry.
func RegisterType(registry *world.EntityRegistry, s Spec) *Type {
	if registry == nil {
		panic("entity.RegisterType: nil entity registry")
	}
	if !strings.Contains(s.Name, ":") {
		panic("entity.RegisterType: Spec.Name must be namespaced, e.g. 'myplugin:wisp', got " + s.Name)
	}
	if s.Actor.ID != "" && !strings.Contains(s.Actor.ID, ":") {
		panic("entity.RegisterType: Spec.Actor.ID must be namespaced, e.g. 'minecraft:allay', got " + s.Actor.ID)
	}
	if err := validateSpecComponents(s); err != nil {
		panic("entity.RegisterType: " + s.Name + ": " + err.Error())
	}
	t := &Type{spec: s}
	registry.Register(t)
	return t
}

func validateSpecComponents(s Spec) error {
	if s.Components == nil {
		return nil
	}
	return validateComponents(s.Components())
}

func validateComponents(values []any) error {
	comps := make([]any, 0, len(values))
	behaviours := 0
	for _, v := range values {
		if behaviour, ok := v.(Behaviour); ok {
			if behaviours++; behaviours > 1 {
				return fmt.Errorf("more than one component implements Behaviour")
			}
			if err := validateBehaviourPersistence(behaviour); err != nil {
				return err
			}
			continue
		}
		comps = append(comps, v)
	}
	return world.ValidateComponents(comps...)
}

// validateBehaviourPersistence rejects field-bearing behaviours whose state
// would otherwise silently reset when an entity is loaded.
func validateBehaviourPersistence(behaviour Behaviour) error {
	if _, ok := behaviour.(PersistentBehaviour); ok {
		return nil
	}
	typ := reflect.TypeOf(behaviour)
	if typ == nil {
		return fmt.Errorf("nil Behaviour")
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Size() != 0 {
		return fmt.Errorf("stateful Behaviour %v must implement entity.PersistentBehaviour or keep mutable state in a persistent component", reflect.TypeOf(behaviour))
	}
	return nil
}

// New creates an entity with the Spec's components and any extras passed.
// It panics if the extra components are invalid.
func (t *Type) New(opts world.EntitySpawnOpts, components ...any) *world.EntityHandle {
	if err := validateComponents(components); err != nil {
		panic("entity.Type.New: " + t.spec.Name + ": " + err.Error())
	}
	return opts.New(t, specConfig{t: t, extra: components})
}

// Open opens an entity of this type.
func (t *Type) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return Open(tx, handle, data)
}

// EncodeEntity returns the Spec's name.
func (t *Type) EncodeEntity() string { return t.spec.Name }

// NetworkEncodeEntity returns the client-side entity ID.
func (t *Type) NetworkEncodeEntity() string {
	if t.spec.Actor.ID != "" {
		return t.spec.Actor.ID
	}
	return t.spec.Name
}

// NetworkOffset returns the vertical offset applied to client movement.
func (t *Type) NetworkOffset() float64 { return t.spec.Actor.Offset }

// BBox returns the Spec's bounding box.
func (t *Type) BBox(world.Entity) cube.BBox { return t.spec.Box }

// DecodeNBT adds default components missing from the saved entity and restores
// versioned behaviour state.
func (t *Type) DecodeNBT(m map[string]any, data *world.EntityData) {
	applyDefaults(t, data, nil)
	behaviour, ok := data.Data.(PersistentBehaviour)
	if !ok {
		return
	}
	saved, ok := m[behaviourNBTKey].(map[string]any)
	if !ok {
		return
	}
	version, versionOK := saved["Version"].(int32)
	state, stateOK := saved["Data"].(map[string]any)
	if versionOK && stateOK {
		behaviour.LoadBehaviourNBT(version, state)
	}
}

// EncodeNBT returns versioned behaviour state. Components save their own
// state through world.NBTSaver.
func (t *Type) EncodeNBT(data *world.EntityData) map[string]any {
	behaviour, ok := data.Data.(PersistentBehaviour)
	if !ok {
		return nil
	}
	version, state := behaviour.SaveBehaviourNBT()
	if state == nil {
		state = map[string]any{}
	}
	return map[string]any{behaviourNBTKey: map[string]any{
		"Version": version,
		"Data":    state,
	}}
}

type specConfig struct {
	t     *Type
	extra []any
}

// Apply adds the Spec's components followed by the extras.
func (c specConfig) Apply(data *world.EntityData) {
	applyDefaults(c.t, data, c.extra)
}

func applyDefaults(t *Type, data *world.EntityData, extra []any) {
	if t.spec.Components != nil {
		for _, comp := range t.spec.Components() {
			if b, ok := comp.(Behaviour); ok {
				if data.Data == nil {
					data.Data = b
				}
				continue
			}
			world.AttachComponentIfAbsent(data, comp)
		}
	}
	for _, comp := range extra {
		if b, ok := comp.(Behaviour); ok {
			data.Data = b
			continue
		}
		world.AttachComponent(data, comp)
	}
}
