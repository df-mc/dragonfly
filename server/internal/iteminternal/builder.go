package iteminternal

import (
	"github.com/df-mc/dragonfly/server/item/category"
	"golang.org/x/exp/maps"
	"strings"
)

// ComponentBuilder represents a builder that can be used to construct an item components map to be sent to a client.
type ComponentBuilder struct {
	name       string
	identifier string
	category   category.Category

	properties map[string]any
	components map[string]any
}

// NewComponentBuilder returns a new component builder with the provided item data.
func NewComponentBuilder(name, identifier string, category category.Category) *ComponentBuilder {
	return &ComponentBuilder{
		name:       name,
		identifier: identifier,
		category:   category,

		properties: make(map[string]any),
		components: make(map[string]any),
	}
}

// AddProperty adds the provided property to the builder.
func (builder *ComponentBuilder) AddProperty(name string, value any) {
	builder.properties[name] = value
}

// AddComponent adds the provided component to the builder.
func (builder *ComponentBuilder) AddComponent(name string, value any) {
	builder.components[name] = value
}

// Empty returns if there are no components or item properties in the builder.
func (builder *ComponentBuilder) Empty() bool {
	return len(builder.properties) == 0 && len(builder.components) == 0
}

// Construct constructs the final item components map and returns it. It also applies the default properties required
// for the item to work without modifying the original maps in the builder.
func (builder *ComponentBuilder) Construct() map[string]any {
	properties := maps.Clone(builder.properties)
	components := maps.Clone(builder.components)
	builder.applyDefaultProperties(properties)
	builder.applyDefaultComponents(components, properties)
	return map[string]any{"components": components}
}

// applyDefaultProperties applies the default properties to the provided map. It is important that this method does
// not modify the builder's properties map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultProperties(x map[string]any) {
	x["minecraft:icon"] = map[string]any{
		"texture": strings.Split(builder.identifier, ":")[1],
	}
	x["creative_group"] = builder.category.String()
	x["creative_category"] = int32(builder.category.Uint8())
	if _, ok := x["max_stack_size"]; !ok {
		x["max_stack_size"] = int32(64)
	}
}

// applyDefaultComponents applies the default components to the provided map. It is important that this method does not
// modify the builder's components map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultComponents(x, properties map[string]any) {
	x["item_properties"] = properties
	x["minecraft:display_name"] = map[string]any{
		"value": builder.name,
	}
}
