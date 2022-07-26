package blockinternal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/segmentio/fasthash/fnv1"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"
)

// ComponentBuilder represents a builder that can be used to construct a block components map to be sent to a client.
type ComponentBuilder struct {
	permutations map[string]map[string]any
	properties   []map[string]any
	components   map[string]any

	identifier string
	group      []world.CustomBlock
}

// NewComponentBuilder returns a new component builder with the provided block data.
func NewComponentBuilder(identifier string, group []world.CustomBlock) *ComponentBuilder {
	return &ComponentBuilder{
		permutations: make(map[string]map[string]any),
		components:   make(map[string]any),

		identifier: identifier,
		group:      group,
	}
}

// AddProperty adds the provided property to the builder.
func (builder *ComponentBuilder) AddProperty(value map[string]any) {
	builder.properties = append(builder.properties, value)
}

// AddComponent adds the provided component to the builder.
func (builder *ComponentBuilder) AddComponent(name string, value any) {
	builder.components[name] = value
}

// AddPermutation adds a permutation to the builder.
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

// Construct constructs the final block components map and returns it. It also applies the default properties required
// for the block to work without modifying the original maps in the builder.
func (builder *ComponentBuilder) Construct() map[string]any {
	properties := slices.Clone(builder.properties)
	components := maps.Clone(builder.components)
	builder.applyDefaultProperties(&properties)
	builder.applyDefaultComponents(components)

	result := map[string]any{"components": components}
	if len(properties) > 0 {
		result["properties"] = properties
	}

	permutations := maps.Clone(builder.permutations)
	if len(permutations) > 0 {
		result["molangVersion"] = int32(0)
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

// applyDefaultProperties applies the default properties to the provided map. It is important that this method does
// not modify the builder's properties map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultProperties(x *[]map[string]any) {
	traits := make(map[string][]any)
	for _, b := range builder.group {
		_, properties := b.EncodeBlock()
		for trait, value := range properties {
			if _, ok := traits[trait]; !ok {
				traits[trait] = []any{}
			}
			if slices.IndexFunc(traits[trait], func(i any) bool {
				return i == value
			}) >= 0 {
				// Already exists, skip.
				continue
			}
			traits[trait] = append(traits[trait], value)
		}
	}
	for trait, values := range traits {
		*x = append(*x, map[string]any{"enum": values, "name": trait})
	}
}

// applyDefaultComponents applies the default components to the provided map. It is important that this method does not
// modify the builder's components map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultComponents(x map[string]any) {
	base := builder.group[0]
	name := strings.Split(builder.identifier, ":")[1]

	geometry, permutationGeometries, _ := base.Geometries()
	generalModel := customblock.NewModel(geometry)

	textures, permutationTextures, method := base.Textures()
	for target := range textures {
		generalModel = generalModel.WithMaterial(target, customblock.NewMaterial(fmt.Sprintf("%v_%v", name, target.Name()), method))
	}

	permutationModels := make(map[string]customblock.Model)
	for permutation, permutationSpecificGeometry := range permutationGeometries {
		permutationModels[permutation] = customblock.NewModel(permutationSpecificGeometry)
	}
	for permutation, permutationSpecificTextures := range permutationTextures {
		h := fnv1.HashString64(permutation)
		for target := range permutationSpecificTextures {
			if _, ok := permutationModels[permutation]; !ok {
				// If we don't have a model for this permutation, re-use the base geometry and create a new model.
				permutationModels[permutation] = customblock.NewModel(geometry)
			}
			permutationModel := permutationModels[permutation]
			permutationModels[permutation] = permutationModel.WithMaterial(target, customblock.NewMaterial(fmt.Sprintf("%s_%s_%x", name, target.Name(), h), method))
		}
	}
	for permutation, model := range permutationModels {
		builder.AddPermutation(permutation, model.Encode())
	}
	for key, value := range generalModel.Encode() {
		x[key] = value
	}
}
