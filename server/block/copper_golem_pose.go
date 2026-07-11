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

// SittingPose is the sitting pose.
func SittingPose() CopperGolemPose {
	return CopperGolemPose{1}
}

// RunningPose is the running pose.
func RunningPose() CopperGolemPose {
	return CopperGolemPose{2}
}

// StarPose is the head button pressing pose.
func StarPose() CopperGolemPose {
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
		return "Sitting"
	case 2:
		return "Running"
	case 3:
		return "Star"
	}
	panic("unknown copper golem pose")
}

// CopperGolemPoses returns all copper golem poses.
func CopperGolemPoses() []CopperGolemPose {
	return []CopperGolemPose{StandingPose(), SittingPose(), RunningPose(), StarPose()}
}
