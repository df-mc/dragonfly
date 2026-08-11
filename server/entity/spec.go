package entity

import (
	"fmt"
	"strings"
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Spec describes a component-based entity type.
type Spec struct {
	// Name is the namespaced ID used for registration and saves.
	Name string
	// NetworkID is the client-side entity ID. It defaults to Name.
	NetworkID string
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

var specTypes sync.Map

// RegisterType registers a Spec. Register its component types first.
// RegisterType panics if the Spec is invalid or already registered.
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
	if t.spec.NetworkID != "" {
		return t.spec.NetworkID
	}
	return t.spec.Name
}

// BBox returns the Spec's bounding box.
func (t *Type) BBox(world.Entity) cube.BBox { return t.spec.Box }

// DecodeNBT adds default components missing from the saved entity.
func (t *Type) DecodeNBT(_ map[string]any, data *world.EntityData) {
	applyDefaults(t, data, nil)
}

// EncodeNBT returns no extra data. Components save their own state.
func (t *Type) EncodeNBT(*world.EntityData) map[string]any { return nil }

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
