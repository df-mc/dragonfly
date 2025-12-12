package blockinternal

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/customblock"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Components returns all the components for the custom block, including permutations and properties.
func Components(identifier string, b world.CustomBlock, blockID int32) map[string]any {
	components := componentsFromProperties(b.Properties())
	builder := NewComponentBuilder(identifier, components, blockID)
	if emitter, ok := b.(block.LightEmitter); ok {
		builder.AddComponent("minecraft:block_light_emission", map[string]any{
			"emission": float32(emitter.LightEmissionLevel() / 15),
		})
	}
	if diffuser, ok := b.(block.LightDiffuser); ok {
		builder.AddComponent("minecraft:block_light_filter", map[string]any{
			"lightLevel": int32(diffuser.LightDiffusionLevel()),
		})
	}
	if breakable, ok := b.(block.Breakable); ok {
		info := breakable.BreakInfo()
		builder.AddComponent("minecraft:destructible_by_mining", map[string]any{"value": float32(info.Hardness)})
	}
	if frictional, ok := b.(block.Frictional); ok {
		builder.AddComponent("minecraft:friction", map[string]any{"value": float32(frictional.Friction())})
	}
	if flammable, ok := b.(block.Flammable); ok {
		info := flammable.FlammabilityInfo()
		builder.AddComponent("minecraft:flammable", map[string]any{
			"flame_odds": int32(info.Encouragement),
			"burn_odds":  int32(info.Flammability),
		})
	}
	if permutable, ok := b.(block.Permutable); ok {
		for name, values := range permutable.States() {
			builder.AddProperty(name, values)
		}
		for _, permutation := range permutable.Permutations() {
			builder.AddPermutation(permutation.Condition, componentsFromProperties(permutation.Properties))
		}
	}
	if item, ok := b.(world.CustomItem); ok {
		builder.SetMenuCategory(item.Category())
	}
	return builder.Construct()
}

// componentsFromProperties builds a base components map that includes all the common data between a regular block and
// a custom permutation.
func componentsFromProperties(props customblock.Properties) map[string]any {
	components := make(map[string]any)
	if props.CollisionBox != (cube.BBox{}) {
		components["minecraft:collision_box"] = collisionBoxComponent(props.CollisionBox)
	}
	if props.SelectionBox != (cube.BBox{}) {
		components["minecraft:selection_box"] = selectionBoxComponent(props.SelectionBox)
	}
	if props.Geometry != "" {
		components["minecraft:geometry"] = map[string]any{"identifier": props.Geometry}
	} else if props.Cube {
		components["minecraft:unit_cube"] = map[string]any{}
	}
	if props.MapColour != "" {
		components["minecraft:map_color"] = map[string]any{"value": props.MapColour}
	}
	if props.Textures != nil {
		materials := map[string]any{}
		for target, material := range props.Textures {
			materials[target] = material.Encode()
		}
		components["minecraft:material_instances"] = map[string]any{
			"mappings":  map[string]any{},
			"materials": materials,
		}
	}
	transformation := make(map[string]any)
	if props.Rotation != (cube.Pos{}) {
		transformation["RX"] = int32(props.Rotation.X())
		transformation["RY"] = int32(props.Rotation.Y())
		transformation["RZ"] = int32(props.Rotation.Z())
	}
	if props.Translation != (mgl64.Vec3{}) {
		transformation["TX"] = float32(props.Translation.X())
		transformation["TY"] = float32(props.Translation.Y())
		transformation["TZ"] = float32(props.Translation.Z())
	}
	if props.Scale != (mgl64.Vec3{}) {
		transformation["SX"] = float32(props.Scale.X())
		transformation["SY"] = float32(props.Scale.Y())
		transformation["SZ"] = float32(props.Scale.Z())
	} else if len(transformation) > 0 {
		transformation["SX"] = float32(1.0)
		transformation["SY"] = float32(1.0)
		transformation["SZ"] = float32(1.0)
	}
	if len(transformation) > 0 {
		components["minecraft:transformation"] = transformation
	}
	return components
}

// collisionBoxComponent returns the component data for a collision box, using absolute min/max coordinates in pixels.
func collisionBoxComponent(box cube.BBox) map[string]any {
	min, max := box.Min(), box.Max()
	return map[string]any{
		"enabled": true,
		"boxes": []map[string]any{
			{
				"minX": float32(min.X() * 16),
				"minY": float32(min.Y() * 16),
				"minZ": float32(min.Z() * 16),
				"maxX": float32(max.X() * 16),
				"maxY": float32(max.Y() * 16),
				"maxZ": float32(max.Z() * 16),
			},
		},
	}
}

// selectionBoxComponent returns the component data for a selection box, translating coordinates to the origin/size
// format the client expects.
func selectionBoxComponent(box cube.BBox) map[string]any {
	min, max := box.Min(), box.Max()
	originX, originY, originZ := min.X()*16, min.Y()*16, min.Z()*16
	sizeX, sizeY, sizeZ := (max.X()-min.X())*16, (max.Y()-min.Y())*16, (max.Z()-min.Z())*16
	return map[string]any{
		"enabled": true,
		"origin":  []float32{float32(originX) - 8, float32(originY), float32(originZ) - 8},
		"size":    []float32{float32(sizeX), float32(sizeY), float32(sizeZ)},
	}
}
