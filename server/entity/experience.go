package entity

import "math"

var maxLevel int32 = math.MaxInt32

type ExperienceManager struct {
	level    int32
	progress float64
	totalXP  int
}

// NewExperienceManager return a ExperienceManager with default value of level, progress and totalXP that are equal to 0.
func NewExperienceManager() *ExperienceManager {
	return &ExperienceManager{}
}

// Level get xp level.
func (e *ExperienceManager) Level() int32 {
	return e.level
}

// Progress get the progress of the xp, the value are Between 0.00 and 1.00.
func (e *ExperienceManager) Progress() float64 {
	return e.progress
}

// MaxLevel get the max xp level.
func (e *ExperienceManager) MaxLevel() int32 {
	return maxLevel
}

// TotalXP get the total xp collected.
func (e *ExperienceManager) TotalXP() int {
	return e.totalXP
}

// GetExperienceToNextLevel get the total experience netted to level up.
func (e *ExperienceManager) GetExperienceToNextLevel() int {
	return e.GetExperienceForLevel(e.level)
}

// GetExperienceNettedToLevelUp get xp that you need to level up with reducing xp that you already have.
func (e *ExperienceManager) GetExperienceNettedToLevelUp() int {
	return e.GetExperienceToNextLevel() - e.ProgressToXp(e.progress, e.level)
}

// GetExperienceForLevel get the amount xp to level up in a specific level.
func (e *ExperienceManager) GetExperienceForLevel(level int32) int {
	// I have make like MINET
	if level > 30 {
		return int(9*level - 158)
	} else if level > 15 {
		return int(5*level - 38)
	} else {
		return int(2*level + 7)
	}
}

// GetExperienceForLevels calculate amount of xp for reach a level.
func (e *ExperienceManager) GetExperienceForLevels(level int32) int {
	xp := 0
	for i := int32(0); i <= level; i++ {
		xp += e.GetExperienceForLevel(i)
	}
	return xp
}

// ProgressToXp get number of xp that you have (not total but start on level) whit progress.
func (e *ExperienceManager) ProgressToXp(progress float64, level int32) int {
	return int(float64(e.GetExperienceForLevel(level)) * progress)
}

// XpToProgress get progress with the amount of xp and a level.
func (e *ExperienceManager) XpToProgress(xp int, level int32) float64 {
	return float64(xp) / float64(e.GetExperienceForLevel(level))
}

// AddXP add xp, that check if you level up.
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

// SetLevel set a new level, this recalculates the total xp.
func (e *ExperienceManager) SetLevel(level int32) {
	if level < 0 {
		level = 0
	} else if level > maxLevel {
		level = maxLevel
	}
	e.level = level
	e.totalXP = e.GetExperienceForLevels(level) + e.ProgressToXp(e.progress, level)
}

// SetProgress set a new progress, this recalculates the total xp.
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
