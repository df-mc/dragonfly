package blockinternal

import (
	"github.com/df-mc/dragonfly/server/item/category"
	"golang.org/x/exp/maps"
	"slices"
)

// ComponentBuilder represents a builder that can be used to construct a block components map to be sent to a client.
type ComponentBuilder struct {
	permutations map[string]map[string]any
	properties   []map[string]any
	components   map[string]any
	blockID      int32

	identifier   string
	menuCategory category.Category
}

// NewComponentBuilder returns a new component builder with the provided block data, using the provided components map
// as a base.
func NewComponentBuilder(identifier string, components map[string]any, blockID int32) *ComponentBuilder {
	if components == nil {
		components = map[string]any{}
	}
	return &ComponentBuilder{
		permutations: make(map[string]map[string]any),
		components:   components,
		blockID:      blockID,

		identifier:   identifier,
		menuCategory: category.Construction(),
	}
}

// AddProperty adds the provided block property to the builder.
func (builder *ComponentBuilder) AddProperty(name string, values []any) {
	builder.properties = append(builder.properties, map[string]any{
		"name": name,
		"enum": values,
	})
}

// AddComponent adds the provided component to the builder. If the component already exists, it will be overwritten.
func (builder *ComponentBuilder) AddComponent(name string, value any) {
	builder.components[name] = value
}

// AddPermutation adds a permutation to the builder. If there is already an existing permutation for the provided
// condition, the new components will be added to the existing permutation.
func (builder *ComponentBuilder) AddPermutation(condition string, components map[string]any) {
	if len(builder.permutations) == 0 {
		// This trigger really does not matter at all, the component just needs to be set for custom block placements to
		// function as expected client-side, when permutations are applied.
		builder.AddComponent("minecraft:on_player_placing", map[string]any{
			"triggerType": "placement_trigger",
		})
	}
	if builder.permutations[condition] == nil {
		builder.permutations[condition] = map[string]any{}
	}
	for key, value := range components {
		builder.permutations[condition][key] = value
	}
}

// SetMenuCategory sets the creative category for the current block.
func (builder *ComponentBuilder) SetMenuCategory(category category.Category) {
	builder.menuCategory = category
}

// Construct constructs the final block components map that is ready to be sent to the client.
func (builder *ComponentBuilder) Construct() map[string]any {
	properties := slices.Clone(builder.properties)
	components := maps.Clone(builder.components)

	result := map[string]any{
		"components":    components,
		"molangVersion": int32(10),
		"menu_category": map[string]any{
			"category": builder.menuCategory.String(),
			"group":    builder.menuCategory.Group(),
		},
		"vanilla_block_data": map[string]any{
			"block_id": builder.blockID,
		},
	}
	if len(properties) > 0 {
		result["properties"] = properties
	}

	permutations := maps.Clone(builder.permutations)
	if len(permutations) > 0 {
		result["permutations"] = []map[string]any{}
		for condition, values := range permutations {
			result["permutations"] = append(result["permutations"].([]map[string]any), map[string]any{
				"condition":  condition,
				"components": values,
			})
		}
	}
	return result
}
