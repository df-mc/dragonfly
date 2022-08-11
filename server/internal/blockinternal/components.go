package blockinternal

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

// Components returns all the components of the given custom block group. If the group has no components, a nil map
// and false are returned.
func Components(identifier string, group []world.CustomBlock) (map[string]any, error) {
	if len(group) == 0 {
		// We don't have any blocks in the group so return false.
		return nil, nil
	}

	base := group[0]
	builder := NewComponentBuilder(identifier, group)
	if r, ok := base.(block.Rotatable); ok {
		rotation := r.Rotation()
		builder.AddComponent("minecraft:rotation", map[string]any{
			"x": float32(rotation.X()),
			"y": float32(rotation.Y()),
			"z": float32(rotation.Z()),
		})
	}
	if r, ok := base.(block.PermutableRotatable); ok {
		rotation, exists, rotations := r.Rotation()
		if exists {
			builder.AddComponent("minecraft:rotation", map[string]any{
				"x": float32(rotation.X()),
				"y": float32(rotation.Y()),
				"z": float32(rotation.Z()),
			})
		}
		for condition, value := range rotations {
			builder.AddPermutation(condition, map[string]any{"minecraft:rotation": map[string]any{
				"x": float32(value.X()),
				"y": float32(value.Y()),
				"z": float32(value.Z()),
			}})
		}
	}
	if l, ok := base.(block.LightEmitter); ok {
		builder.AddComponent("minecraft:block_light_emission", map[string]any{
			"emission": float32(l.LightEmissionLevel() / 15),
		})
	}
	if l, ok := base.(block.PermutableLightEmitter); ok {
		level, exists, levels := l.LightEmissionLevel()
		if exists {
			builder.AddComponent("minecraft:block_light_emission", map[string]any{
				"emission": float32(level / 15),
			})
		}
		for condition, value := range levels {
			builder.AddPermutation(condition, map[string]any{"minecraft:block_light_emission": map[string]any{
				"emission": float32(value / 15),
			}})
		}
	}
	if d, ok := base.(block.LightDiffuser); ok {
		builder.AddComponent("minecraft:block_light_filter", map[string]any{
			"lightLevel": int32(d.LightDiffusionLevel()),
		})
	}
	if d, ok := base.(block.PermutableLightDiffuser); ok {
		level, exists, levels := d.LightDiffusionLevel()
		if exists {
			builder.AddComponent("minecraft:block_light_filter", map[string]any{"lightLevel": int32(level)})
		}
		for condition, value := range levels {
			builder.AddPermutation(condition, map[string]any{"minecraft:block_light_filter": map[string]any{
				"lightLevel": int32(value),
			}})
		}
	}
	if i, ok := base.(block.Breakable); ok {
		info := i.BreakInfo()
		builder.AddComponent("minecraft:destroy_time", map[string]any{"value": float32(info.Hardness)})
		// TODO: Explosion resistance.
	}
	if i, ok := base.(block.PermutableBreakable); ok {
		info, exists, infos := i.BreakInfo()
		if exists {
			builder.AddComponent("minecraft:destroy_time", map[string]any{"value": float32(info.Hardness)})
		}
		for condition, value := range infos {
			builder.AddPermutation(condition, map[string]any{"minecraft:destroy_time": map[string]any{
				"value": float32(value.Hardness),
			}})
		}
		// TODO: Explosion resistance.
	}
	if f, ok := base.(block.Frictional); ok {
		builder.AddComponent("minecraft:friction", map[string]any{"value": float32(f.Friction())})
	}
	if f, ok := base.(block.PermutableFrictional); ok {
		friction, exists, frictions := f.Friction()
		if exists {
			builder.AddComponent("minecraft:friction", map[string]any{"value": float32(friction)})
		}
		for condition, value := range frictions {
			builder.AddPermutation(condition, map[string]any{"minecraft:friction": map[string]any{
				"value": float32(value),
			}})
		}
	}
	if f, ok := base.(block.Flammable); ok {
		info := f.FlammabilityInfo()
		builder.AddComponent("minecraft:flammable", map[string]any{
			"flame_odds": int32(info.Encouragement),
			"burn_odds":  int32(info.Flammability),
		})
	}
	if f, ok := base.(block.PermutableFlammable); ok {
		info, exists, infos := f.FlammabilityInfo()
		if exists {
			builder.AddComponent("minecraft:flammable", map[string]any{
				"flame_odds": int32(info.Encouragement),
				"burn_odds":  int32(info.Flammability),
			})
		}
		for condition, value := range infos {
			builder.AddPermutation(condition, map[string]any{"minecraft:flammable": map[string]any{
				"flame_odds": int32(value.Encouragement),
				"burn_odds":  int32(value.Flammability),
			}})
		}
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
