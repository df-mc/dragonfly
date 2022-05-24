package blockinternal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"
)

// ComponentBuilder represents a builder that can be used to construct a block components map to be sent to a client.
type ComponentBuilder struct {
	permutations []map[string]any
	properties   []map[string]any
	components   map[string]any
	events       map[string]any

	identifier string
	group      []world.CustomBlock
	traits     map[string][]any
}

// NewComponentBuilder returns a new component builder with the provided block data.
func NewComponentBuilder(identifier string, group []world.CustomBlock) *ComponentBuilder {
	traits := make(map[string][]any)
	for _, b := range group {
		_, properties := b.EncodeBlock()
		for trait, value := range properties {
			if _, ok := traits[trait]; !ok {
				traits[trait] = []any{}
			}
			traits[trait] = append(traits[trait], value)
		}
	}
	return &ComponentBuilder{
		properties: make([]map[string]any, 0),
		components: make(map[string]any),
		events:     make(map[string]any),

		identifier: identifier,
		traits:     traits,
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

// AddEvent adds an event to the builder.
func (builder *ComponentBuilder) AddEvent(name, event, property, query string) {
	builder.events[name] = map[string]any{
		event: map[string]any{
			property: query,
		},
	}
}

// AddDirectionPermutation adds a direction rotation permutation to the builder.
func (builder *ComponentBuilder) AddDirectionPermutation(property string, target cube.Direction, rotation mgl32.Vec3) {
	builder.AddRotationPermutation(fmt.Sprintf("query.block_property('%s') == %v", property, int32(target.Face())), rotation)
}

// AddAxisPermutation adds a permutation that rotates the block around the axis.
func (builder *ComponentBuilder) AddAxisPermutation(property string, target cube.Axis, rotation mgl32.Vec3) {
	builder.AddRotationPermutation(fmt.Sprintf("query.block_property('%s') == %v", property, int32(target)), rotation)
}

// AddRotationPermutation adds a rotation permutation to the builder.
func (builder *ComponentBuilder) AddRotationPermutation(condition string, rotation mgl32.Vec3) {
	builder.AddPermutation(condition, []map[string]any{{
		"minecraft:rotation": map[string]any{
			"x": rotation.X(),
			"y": rotation.Y(),
			"z": rotation.Z(),
		},
	}})
}

// AddPermutation adds a permutation to the builder.
func (builder *ComponentBuilder) AddPermutation(condition string, components []map[string]any) {
	builder.permutations = append(builder.permutations, map[string]any{
		"condition":  condition,
		"components": components,
	})
}

// Values returns the values of a given trait.
func (builder *ComponentBuilder) Values(trait string) ([]any, bool) {
	values, ok := builder.traits[trait]
	return values, ok
}

// Trait finds a trait which satisfies all given values.
func (builder *ComponentBuilder) Trait(desired ...any) (string, bool) {
	for trait, values := range builder.traits {
		if len(values) != len(values) {
			// Not the same length, can't possibly be a match.
			continue
		}
		for i := range desired {
			if values[i] != desired[i] {
				continue
			}
		}
		return trait, true
	}
	return "", false
}

// Empty returns if there are no components or block properties in the builder.
func (builder *ComponentBuilder) Empty() bool {
	return len(builder.properties) == 0 && len(builder.components) == 0
}

// Construct constructs the final block components map and returns it. It also applies the default properties required
// for the block to work without modifying the original maps in the builder.
func (builder *ComponentBuilder) Construct() map[string]any {
	permutations := slices.Clone(builder.permutations)
	properties := slices.Clone(builder.properties)
	components := maps.Clone(builder.components)
	events := maps.Clone(builder.events)
	builder.applyDefaultProperties(&properties)
	builder.applyDefaultComponents(components)
	result := map[string]any{"components": components}
	if len(properties) > 0 {
		result["properties"] = properties
	}
	if len(permutations) > 0 {
		result["permutations"] = permutations
	}
	if len(events) > 0 {
		result["events"] = events
	}
	return result
}

// applyDefaultProperties applies the default properties to the provided map. It is important that this method does
// not modify the builder's properties map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultProperties(x *[]map[string]any) {
	for trait, values := range builder.traits {
		*x = append(*x, map[string]any{"enum": values, "name": trait})
	}
}

// applyDefaultComponents applies the default components to the provided map. It is important that this method does not
// modify the builder's components map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultComponents(x map[string]any) {
	base := builder.group[0]
	name := strings.Split(builder.identifier, ":")[1]
	materials := make(map[customblock.MaterialTarget]customblock.Material)
	for target := range base.Textures() {
		materials[target] = customblock.NewMaterial(fmt.Sprintf("%v_%v", name, target.Name()), customblock.OpaqueRenderMethod())
	}

	geometries := base.Geometries().Geometry
	if len(geometries) == 0 {
		panic("block needs at least one geometry")
	}

	geometry := geometries[0]
	model := customblock.NewModel(geometry.Description.Identifier, geometry.Origin(), geometry.Size())
	for target, material := range materials {
		model = model.WithMaterial(target, material)
	}
	for key, value := range model.Encode() {
		x[key] = value
	}
}
