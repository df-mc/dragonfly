package entity

import (
	"fmt"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Spec declaratively describes an entity type, replacing hand-written
// world.EntityType implementations. Types created from a Spec are composed of
// components; see world.RegisterComponent.
type Spec struct {
	// Name is the namespaced identifier the entity is registered and saved
	// under, e.g. "myplugin:wisp".
	Name string
	// NetworkID is the identifier sent to clients, allowing an entity to
	// render as an existing client-side type, e.g. "minecraft:vex". It
	// defaults to Name.
	NetworkID string
	// Box is the bounding box of entities of this type.
	Box cube.BBox
	// Components returns the components attached to entities of this type at
	// spawn. It must return fresh values on every call: pointer components
	// returned from a shared value would alias state between entities. A
	// component implementing Behaviour becomes the entity's main behaviour;
	// it is bridged at runtime only and is not persisted with the entity,
	// even if it implements world.NBTSaver.
	Components func() []any
}

// Type is a world.EntityType created from a Spec. Entities of a Type are
// spawned with Type.New or Tx.SpawnEntity and must be registered in the
// World's EntityRegistry (world.Config.Entities) to load from the world save.
type Type struct {
	spec Spec
}

// specTypes tracks registered Specs by name to reject duplicates early, at
// registration rather than at EntityRegistry construction.
var specTypes sync.Map

// RegisterType registers a Spec and returns the Type for it. The Spec is
// validated eagerly: its Components are checked for unregistered types,
// duplicates and multiple Behaviours, so mistakes surface at init instead of
// at first spawn. Component types must therefore be registered before
// RegisterType runs: within a single package, declare component keys above
// spec variables. RegisterType is intended to be called from package-level
// variable initialisation and panics if the Spec is invalid or its name is
// already registered.
func RegisterType(s Spec) *Type {
	if !strings.Contains(s.Name, ":") {
		panic("entity.RegisterType: Spec.Name must be namespaced, e.g. 'myplugin:wisp', got " + s.Name)
	}
	if err := validateSpecComponents(s); err != nil {
		panic("entity.RegisterType: " + s.Name + ": " + err.Error())
	}
	t := &Type{spec: s}
	if _, loaded := specTypes.LoadOrStore(s.Name, t); loaded {
		panic("entity.RegisterType: name " + s.Name + " registered twice")
	}
	return t
}

// validateSpecComponents checks a Spec's component list for unregistered
// component types, duplicates and multiple Behaviours.
func validateSpecComponents(s Spec) error {
	if s.Components == nil {
		return nil
	}
	var comps []any
	behaviours := 0
	for _, v := range s.Components() {
		if _, ok := v.(Behaviour); ok {
			if behaviours++; behaviours > 1 {
				return fmt.Errorf("more than one component implements Behaviour")
			}
			continue
		}
		comps = append(comps, v)
	}
	return world.ValidateComponents(comps...)
}

// New creates an entity handle of this Type with the Spec's default
// components and any extra components passed attached. The handle may be
// added to a world using Tx.AddEntity. New panics if an extra component's
// type was not registered with world.RegisterComponent.
func (t *Type) New(opts world.EntitySpawnOpts, components ...any) *world.EntityHandle {
	extras := make([]any, 0, len(components))
	for _, v := range components {
		if _, ok := v.(Behaviour); ok {
			continue
		}
		extras = append(extras, v)
	}
	if err := world.ValidateComponents(extras...); err != nil {
		panic("entity.Type.New: " + t.spec.Name + ": " + err.Error())
	}
	return opts.New(t, specConfig{t: t, extra: components})
}

// Open returns an Ent for the handle. Entities of a Type run their component
// logic and, if a Behaviour component is present, that behaviour.
func (t *Type) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return Open(tx, handle, data)
}

// EncodeEntity returns the Spec's name.
func (t *Type) EncodeEntity() string { return t.spec.Name }

// NetworkEncodeEntity returns the identifier the entity renders as
// client-side: the Spec's NetworkID, or its Name if unset.
func (t *Type) NetworkEncodeEntity() string {
	if t.spec.NetworkID != "" {
		return t.spec.NetworkID
	}
	return t.spec.Name
}

// BBox returns the Spec's bounding box.
func (t *Type) BBox(world.Entity) cube.BBox { return t.spec.Box }

// DecodeNBT attaches the Spec's default components where the saved data,
// decoded before this call, did not already restore them.
func (t *Type) DecodeNBT(_ map[string]any, data *world.EntityData) {
	applyDefaults(t, data, nil)
}

// EncodeNBT returns no extra data: component state is persisted with the
// entity itself through world.NBTSaver. A Behaviour among the Spec's
// components is not persisted; it is recreated from the Spec on load.
func (t *Type) EncodeNBT(*world.EntityData) map[string]any { return nil }

// specConfig is the world.EntityConfig used by Type.New.
type specConfig struct {
	t     *Type
	extra []any
}

// Apply attaches the Spec's default components followed by the extra
// components passed to New.
func (c specConfig) Apply(data *world.EntityData) {
	applyDefaults(c.t, data, c.extra)
}

// applyDefaults attaches a Type's default components to data without
// replacing components already present, then attaches extra components,
// which do replace. A component implementing Behaviour is set as the
// entity's main behaviour instead.
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
