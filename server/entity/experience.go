package entity

import (
	"math"
	"sync"
)

var maxLevel int32 = math.MaxInt32

type ExperienceManager struct {
	mu              sync.RWMutex
	totalExperience int
}

// NewExperienceManager return a ExperienceManager with default value of level, progress and totalExperience that are equal to 0.
func NewExperienceManager() *ExperienceManager {
	return &ExperienceManager{}
}

// Level get experience level.
func (e *ExperienceManager) Level() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := e.totalExperience
	var level int32 = 1
	nextLevel := true
	for nextLevel {
		xp := e.ExperienceForLevel(level)
		if xp <= experience {
			experience -= xp
			level++
		} else {
			nextLevel = false
		}
	}
	return level
}

// Progress get the progress of the experience bar, the value are Between 0.00 and 1.00.
func (e *ExperienceManager) Progress() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := e.totalExperience
	var level int32 = 1
	nextLevel := true
	for nextLevel {
		xp := e.ExperienceForLevel(level)
		if xp <= experience {
			experience -= xp
			level++
		} else {
			nextLevel = false
		}
	}
	return e.ExperienceToProgress(experience, level)
}

func (e *ExperienceManager) LevelAndProgress() (int32, float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := e.totalExperience
	var level int32 = 1
	nextLevel := true
	for nextLevel {
		xp := e.ExperienceForLevel(level)
		if xp <= experience {
			experience -= xp
			level++
		} else {
			nextLevel = false
		}
	}
	return level, e.ExperienceToProgress(experience, level)
}

// MaxLevel get the max experience level.
func (e *ExperienceManager) MaxLevel() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return maxLevel
}

// TotalExperience get the total experience collected.
func (e *ExperienceManager) TotalExperience() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.totalExperience
}

// SetTotalExperience set the total experience
func (e *ExperienceManager) SetTotalExperience(experience int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.totalExperience = experience
}

// ExperienceToNextLevel get the total experience netted to level up.
func (e *ExperienceManager) ExperienceToNextLevel() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ExperienceForLevel(e.Level())
}

// ExperienceNeededToLevelUp get experience that you need to level up with reducing experience that you already have.
func (e *ExperienceManager) ExperienceNeededToLevelUp() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := e.totalExperience
	var level int32 = 1
	nextLevel := true
	for nextLevel {
		xp := e.ExperienceForLevel(level)
		if xp <= experience {
			experience -= xp
			level++
		} else {
			nextLevel = false
		}
	}
	return e.ExperienceForLevel(level) - experience
}

// ExperienceForLevel get the amount experience to level up in a specific level.
func (e *ExperienceManager) ExperienceForLevel(level int32) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// I have make like MINET
	if level > 30 {
		return int(9*level - 158)
	} else if level > 15 {
		return int(5*level - 38)
	}
	return int(2*level + 7)
}

// ExperienceForLevels calculate amount of experience for reach a level.
func (e *ExperienceManager) ExperienceForLevels(level int32) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := 0
	for i := int32(0); i <= level; i++ {
		experience += e.ExperienceForLevel(i)
	}
	return experience
}

// ProgressToExperience get the amount of experience that you have (not total but start on level) whit progress.
func (e *ExperienceManager) ProgressToExperience(level int32, progress float64) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return int(float64(e.ExperienceForLevel(level)) * progress)
}

// ExperienceToProgress get progress with the amount of experience and a level.
func (e *ExperienceManager) ExperienceToProgress(experience int, level int32) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return float64(experience) / float64(e.ExperienceForLevel(level))
}

// AddExperience add experience, that check if you level up.
func (e *ExperienceManager) AddExperience(amount int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.totalExperience += amount
}

// SetLevel set a new level, this recalculates the total experience.
func (e *ExperienceManager) SetLevel(level int32) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if level < 0 {
		level = 0
	} else if level > maxLevel {
		level = maxLevel
	}
	progress := e.Progress()
	e.totalExperience = e.ExperienceForLevels(level) + e.ProgressToExperience(level, progress)

}

// SetProgress set a new progress, this recalculates the total experience.
func (e *ExperienceManager) SetProgress(progress float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if progress < 0 {
		progress = 0
	} else if progress >= 1 {
		progress = 0.99
	}
	lvl := e.Level()
	e.totalExperience = e.ExperienceForLevels(lvl) + e.ProgressToExperience(lvl, progress)
}
