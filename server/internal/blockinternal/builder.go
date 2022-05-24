package blockinternal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"
)

// ComponentBuilder represents a builder that can be used to construct a block components map to be sent to a client.
type ComponentBuilder struct {
	properties []map[string]any
	components map[string]any

	identifier string
	group      []world.CustomBlock
}

// NewComponentBuilder returns a new component builder with the provided block data.
func NewComponentBuilder(identifier string, group []world.CustomBlock) *ComponentBuilder {
	return &ComponentBuilder{
		properties: make([]map[string]any, 0),
		components: make(map[string]any),

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

// Empty returns if there are no components or block properties in the builder.
func (builder *ComponentBuilder) Empty() bool {
	return len(builder.properties) == 0 && len(builder.components) == 0
}

// Construct constructs the final block components map and returns it. It also applies the default properties required
// for the block to work without modifying the original maps in the builder.
func (builder *ComponentBuilder) Construct() map[string]any {
	properties := slices.Clone(builder.properties)
	components := maps.Clone(builder.components)
	builder.applyDefaultProperties(properties)
	builder.applyDefaultComponents(components)
	result := map[string]any{"components": components}
	if len(properties) > 0 {
		result["properties"] = properties
	}
	return result
}

// applyDefaultProperties applies the default properties to the provided map. It is important that this method does
// not modify the builder's properties map directly otherwise Empty() will return false in future use of the builder.
func (builder *ComponentBuilder) applyDefaultProperties(x []map[string]any) {
	traits := make(map[string][]any)
	for _, b := range builder.group {
		_, properties := b.EncodeBlock()
		for trait, value := range properties {
			if _, ok := traits[trait]; !ok {
				traits[trait] = []any{}
			}
			traits[trait] = append(traits[trait], value)
		}
	}
	for trait, values := range traits {
		x = append(x, map[string]any{"enum": values, "name": trait})
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
