package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// PistonArmCollision is the collision block for the piston arm.
type PistonArmCollision struct {
	transparent
	// Facing is the direction the arm faces.
	Facing cube.Face
	// Sticky is true if the piston is sticky.
	Sticky bool
}

// BreakInfo ...
func (p PistonArmCollision) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, nil)
}

// EncodeItem ...
func (p PistonArmCollision) EncodeItem() (name string, meta int16) {
	return "minecraft:piston_arm_collision", 0
}

// EncodeBlock ...
func (p PistonArmCollision) EncodeBlock() (string, map[string]any) {
	return "minecraft:piston_arm_collision", map[string]any{"facing_direction": int32(p.Facing)}
}

const hashPistonArmCollision = 12346 // Temporary constant

// Hash ...
func (p PistonArmCollision) Hash() (uint64, uint64) {
	return hashPistonArmCollision, uint64(p.Facing) | uint64(boolByte(p.Sticky))<<3
}

// Model ...
func (p PistonArmCollision) Model() world.BlockModel {
	return model.Empty{}
}

// allPistonArms ...
func allPistonArms() (arms []world.Block) {
	for _, face := range cube.Faces() {
		for _, sticky := range []bool{false, true} {
			arms = append(arms, PistonArmCollision{Facing: face, Sticky: sticky})
		}
	}
	return
}
