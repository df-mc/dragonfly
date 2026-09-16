package component

//go:generate go run ../../../cmd/generate componentgen -output ../../../server/item/component -vanilla-nbt ../../../server/world/vanilla_items.nbt

// Component is an item component that can be encoded to NBT for the client.
type Component interface {
	// ComponentName returns the namespaced component identifier (e.g., "minecraft:wearable").
	ComponentName() string
	// Encode returns the component data as a map for NBT serialization.
	Encode() (map[string]any, error)
}

// ComponentItem is implemented by custom items that provide typed components.
// This replaces the monolithic type-assertion approach in iteminternal.
type ComponentItem interface {
	ItemComponents() []Component
}

// RawComponent is an escape hatch for components not yet modeled in the typed API.
type RawComponent struct {
	Name string
	Data map[string]any
}

func (r RawComponent) ComponentName() string           { return r.Name }
func (r RawComponent) Encode() (map[string]any, error) { return r.Data, nil }
