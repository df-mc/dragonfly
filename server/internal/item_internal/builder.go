package item_internal

import (
	"github.com/df-mc/dragonfly/server/item/category"
	"strings"
)

// ComponentBuilder represents a builder that can be used to construct an item components map to be sent to a client.
type ComponentBuilder struct {
	name       string
	identifier string
	category   category.Category

	itemProperties map[string]interface{}
	components     map[string]interface{}
}

// NewComponentBuilder returns a new component builder with the provided item data.
func NewComponentBuilder(name, identifier string, category category.Category) *ComponentBuilder {
	return &ComponentBuilder{
		name:       name,
		identifier: identifier,
		category:   category,

		itemProperties: make(map[string]interface{}),
		components:     make(map[string]interface{}),
	}
}

// AddItemProperty adds the provided item property to the builder.
func (builder *ComponentBuilder) AddItemProperty(name string, value interface{}) {
	builder.itemProperties[name] = value
}

// AddComponent adds the provided component to the builder.
func (builder *ComponentBuilder) AddComponent(name string, value interface{}) {
	builder.components[name] = value
}

// Empty returns if there are no components or item properties in the builder.
func (builder *ComponentBuilder) Empty() bool {
	return len(builder.itemProperties) == 0 && len(builder.components) == 0
}

// Construct constructs the final item components map and returns it. It also applies the default properties required
// for the item to work without modifying the original maps in the builder.
func (builder *ComponentBuilder) Construct() map[string]interface{} {
	itemProperties := builder.copyMap(builder.itemProperties)
	components := builder.copyMap(builder.components)
	builder.applyDefaultItemProperties(itemProperties)
	builder.applyDefaultComponents(components, itemProperties)
	return map[string]interface{}{"components": components}
}

// applyDefaultItemProperties applies the default itemProperties to the provided map. It is important that this method does
// not modify the builder's itemProperties map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultItemProperties(x map[string]interface{}) {
	x["minecraft:icon"] = map[string]interface{}{
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
func (builder *ComponentBuilder) applyDefaultComponents(x, itemProperties map[string]interface{}) {
	x["item_properties"] = itemProperties
	x["minecraft:display_name"] = map[string]interface{}{
		"value": builder.name,
	}
}

// copyMap copies the keys and values of the provided map into a new map and returns it.
func (builder *ComponentBuilder) copyMap(x map[string]interface{}) map[string]interface{} {
	y := make(map[string]interface{})
	for k, v := range x {
		y[k] = v
	}
	return y
}
