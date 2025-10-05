package block

// CopperGolemPose represents a pose of a copper golem statue.
type CopperGolemPose struct {
	pose
}

type pose uint8

// StandingPose is the standing pose.
func StandingPose() CopperGolemPose {
	return CopperGolemPose{0}
}

// PointingPose is the pointing pose.
func PointingPose() CopperGolemPose {
	return CopperGolemPose{1}
}

// CrouchingPose is the crouching pose.
func CrouchingPose() CopperGolemPose {
	return CopperGolemPose{2}
}

// HeadButtonPose is the head button pressing pose.
func HeadButtonPose() CopperGolemPose {
	return CopperGolemPose{3}
}

// Uint8 returns the pose as a uint8.
func (p pose) Uint8() uint8 {
	return uint8(p)
}

// Name returns the pose as a string.
func (p pose) Name() string {
	switch p {
	case 0:
		return "Standing"
	case 1:
		return "Pointing"
	case 2:
		return "Crouching"
	case 3:
		return "Head Button"
	}
	panic("unknown copper golem pose")
}

// CopperGolemPoses returns all copper golem poses.
func CopperGolemPoses() []CopperGolemPose {
	return []CopperGolemPose{StandingPose(), PointingPose(), CrouchingPose(), HeadButtonPose()}
}
