package entity

import "math"

var maxLevel = math.MaxInt32

type ExperienceManager struct {
	level    int
	progress float64
	totalXP  int
}

func NewExperienceManager() *ExperienceManager {
	return &ExperienceManager{
		0,
		0,
		0,
	}
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
func (e *ExperienceManager) TotalXP() int {
	return e.totalXP
}

// SetTotalXP internal
func (e *ExperienceManager) SetTotalXP(amount int) {
	e.totalXP = amount
}
func (e *ExperienceManager) GetExperienceToNextLevel() int {
	return e.GetExperienceForLevel(e.level)
}
func (e *ExperienceManager) GetExperienceNettedToLevelUp() int {
	return e.GetExperienceToNextLevel() - e.ProgressToXp(e.progress, e.level)
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

// GetExperienceForLevels calculate amount of xp for reach a level
func (e *ExperienceManager) GetExperienceForLevels(level int) int {
	xp := 0
	for i := 0; i <= level; i++ {
		xp += e.GetExperienceForLevel(i)
	}
	return xp
}

// ProgressToXp get number of xp that you have (not total but start on level) whit progress
func (e *ExperienceManager) ProgressToXp(progress float64, level int) int {
	return int(float64(e.GetExperienceForLevel(level)) * progress)
}
func (e *ExperienceManager) XpToProgress(xp int, level int) float64 {
	return float64(xp) / float64(e.GetExperienceForLevel(level))
}
func (e *ExperienceManager) AddXP(amount int) {
	// level up
	for e.GetExperienceNettedToLevelUp()-amount <= 0 {
		amount -= e.GetExperienceToNextLevel() - e.GetExperienceNettedToLevelUp()
		if e.level == maxLevel {
			return
		}
		e.level++
	}
	e.progress = e.XpToProgress(amount, e.level)
}
func (e *ExperienceManager) SetLevel(level int) {
	if level < 0 {
		level = 0
	} else if level > maxLevel {
		level = maxLevel
	}
	e.level = level
	e.totalXP = e.GetExperienceForLevels(level) + e.ProgressToXp(e.progress, level)
}
func (e *ExperienceManager) SetProgress(progress float64) {
	if progress < 0 {
		progress = 0
	} else if progress >= 1 {
		progress = 0.99
	}
	if e.progress == progress {
		return
	} else if e.progress > progress {
		e.totalXP += e.ProgressToXp(progress, e.level) - e.ProgressToXp(e.progress, e.level)
	} else {
		e.totalXP = e.ProgressToXp(e.progress, e.level) - e.ProgressToXp(progress, e.level)
	}
	e.progress = progress
}
