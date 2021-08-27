package entity

var maxLevel = 21474836477

type ExperienceManager struct {
	level    int
	progress float64
}

func (e *ExperienceManager) Level() int {
	return e.level
}
func (e *ExperienceManager) Progress() float64 {
	return e.progress
}
func (e *ExperienceManager) MaxLevel() int {
	return maxLevel
}
func (e *ExperienceManager) GetExperienceForReachNextLevel() int {
	return e.GetExperienceForLevel(e.level)
}
func (e *ExperienceManager) GetExperienceForLevel(level int) int {
	// I have make like MINET
	if level > 30 {
		return 9*level - 158
	} else if level > 15 {
		return 5*level - 38
	} else {
		return 2*level + 7
	}
}

// ProgressToXp get number of xp that you have (not total but start on level) whit progress
func (e *ExperienceManager) ProgressToXp(progress float64, level int) int {
	return int(float64(e.GetExperienceForLevel(level)) * progress)
}
func (e *ExperienceManager) XpToProgress(xp int, level int) float64 {
	return float64(xp) / float64(e.GetExperienceForLevel(level))
}
