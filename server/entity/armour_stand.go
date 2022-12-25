package entity

import (
    "github.com/df-mc/dragonfly/server/world"
    "github.com/df-mc/dragonfly/server/block/cube"
    "github.com/go-gl/mathgl/mgl64"
)

// ArmourStand ...
type ArmourStand struct {
    transform
    poseIndex int
    yaw, pitch float64
}

// NewArmourStand ...
func NewArmourStand(pos mgl64.Vec3) *ArmourStand {
    	as := &ArmourStand{}
    	as.transform = newTransform(as, pos)
	as.poseIndex = 6
    	return as
}

// PoseIndex returns the index of the pose held by the armour stand.
func (as *ArmourStand) PoseIndex() int {
	return as.poseIndex
}

// Type ...
func (*ArmourStand) Type() world.EntityType {
    return ArmourStandType{}
}

// ArmourStandType is a world.EntityType implementation for ArmourStand.
type ArmourStandType struct{}

func (ArmourStandType) EncodeEntity() string { return "minecraft:armor_stand" }

func (ArmourStandType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.5, 0, -0.5, 0.5, 2, 0.5)
}

func (as ArmourStand) DecodeNBT(m map[string]any) world.Entity {
	return &ArmourStand {
		as.transform,
		as.poseIndex,
	}
}

func (as ArmourStand) EncodeNBT(e world.Entity) map[string]any {
	return map[string]any{
		"Pos": []float32{0, 2, 0},
		"Pose": map[string]any {
			"PoseIndex": 6,
			"LastSignal": 0,
		},
	}
}
