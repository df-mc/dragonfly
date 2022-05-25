package blockinternal

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
)

// Components returns all the components of the given custom block group. If the group; has no components, a nil map
// and false are returned.
func Components(identifier string, group []world.CustomBlock) (map[string]any, error) {
	if len(group) == 0 {
		// We don't have any blocks in the group so return false.
		return nil, nil
	}

	base := group[0]
	builder := NewComponentBuilder(identifier, group)
	if _, ok := base.(block.DirectionRotatable); ok {
		property, ok := builder.Trait(int32(cube.FaceNorth), int32(cube.FaceSouth), int32(cube.FaceEast), int32(cube.FaceWest))
		if !ok {
			return nil, fmt.Errorf("implemented rotatable with no direction trait")
		}

		builder.AddDirectionPermutation(property, cube.North, mgl32.Vec3{0, 180})
		builder.AddDirectionPermutation(property, cube.East, mgl32.Vec3{0, 90})
		builder.AddDirectionPermutation(property, cube.South, mgl32.Vec3{})
		builder.AddDirectionPermutation(property, cube.West, mgl32.Vec3{0, 270})

		// We don't actually need to define any events to trigger this, but we do need to define a trigger type?
		builder.AddComponent("minecraft:on_player_placing", map[string]any{
			"triggerType": "set_direction",
		})
	}
	if _, ok := base.(block.AxisRotatable); ok {
		property, ok := builder.Trait(int32(cube.Y), int32(cube.Z), int32(cube.X))
		if !ok {
			return nil, fmt.Errorf("implemented rotatable with no axis trait")
		}

		builder.AddAxisPermutation(property, cube.Y, mgl32.Vec3{})
		builder.AddAxisPermutation(property, cube.Z, mgl32.Vec3{90})
		builder.AddAxisPermutation(property, cube.X, mgl32.Vec3{0, 0, 90})

		// Refer to comment above.
		builder.AddComponent("minecraft:on_player_placing", map[string]any{
			"triggerType": "set_axis",
		})
	}
	if l, ok := base.(block.LightEmitter); ok {
		builder.AddComponent("minecraft:block_light_emission", map[string]any{
			"emission": float32(l.LightEmissionLevel() / 15),
		})
	}
	if d, ok := base.(block.LightDiffuser); ok {
		builder.AddComponent("minecraft:block_light_filter", map[string]any{
			"lightLevel": int32(d.LightDiffusionLevel()),
		})
	}
	if i, ok := base.(block.Breakable); ok {
		info := i.BreakInfo()
		builder.AddComponent("minecraft:destroy_time", map[string]any{
			"value": float32(info.Hardness),
		})
		// TODO: Explosion resistance.
	}
	if f, ok := base.(block.Frictional); ok {
		builder.AddComponent("minecraft:friction", map[string]any{
			"value": float32(f.Friction()),
		})
	}
	if f, ok := base.(block.Flammable); ok {
		info := f.FlammabilityInfo()
		builder.AddComponent("minecraft:flammable", map[string]any{
			"flame_odds": int32(info.Encouragement),
			"burn_odds":  int32(info.Flammability),
		})
	}
	if c, ok := base.(world.CustomItem); ok {
		category := c.Category()
		builder.AddComponent("minecraft:creative_category", map[string]any{
			"category": category.Name(),
			"group":    category.String(),
		})
	}
	return builder.Construct(), nil
}
