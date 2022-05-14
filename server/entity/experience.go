package entity

import (
	"fmt"
	"math"
	"sync"
)

// ExperienceManager manages experience and levels for entities, and provides functions to add, remove, and calculate
// experience needed for upcoming levels.
type ExperienceManager struct {
	mu         sync.RWMutex
	experience int
}

// NewExperienceManager returns a new ExperienceManager with no experience.
func NewExperienceManager() *ExperienceManager {
	return &ExperienceManager{}
}

// Level returns the current experience level.
// TODO: Improve.
func (e *ExperienceManager) Level() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := e.experience
	level := 1
	nextLevel := true
	for nextLevel {
		xp := experienceForLevel(level)
		if xp <= experience {
			experience -= xp
			level++
		} else {
			nextLevel = false
		}
	}
	return level
}

// Progress returns the progress towards the next level, calculated using the current level and experience.
// TODO: Improve.
func (e *ExperienceManager) Progress() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	experience := e.experience
	level := 1
	nextLevel := true
	for nextLevel {
		xp := experienceForLevel(level)
		if xp <= experience {
			experience -= xp
			level++
		} else {
			nextLevel = false
		}
	}
	return experienceToProgress(experience, level)
}

// TotalExperience returns the total experience collected overall, including levels.
func (e *ExperienceManager) TotalExperience() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.experience
}

// SetTotalExperience sets the total experience collected overall, including levels.
func (e *ExperienceManager) SetTotalExperience(experience int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.experience = experience
}

// AddExperience adds experience to the manager's total experience.
func (e *ExperienceManager) AddExperience(amount int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	e.experience += amount
}

// SetLevel sets the experience level of the manager, recalculating the total experience.
// TODO: Improve.
func (e *ExperienceManager) SetLevel(level int) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if level < 0 || level > math.MaxInt32 {
		return fmt.Errorf("progress must be between 0 and 2,147,483,647, got %d", level)
	}
	progress := e.Progress()
	e.experience = experienceForLevels(level) + progressToExperience(level, progress)
	return nil
}

// SetProgress sets the experience progress of the manager, recalculating the total experience.
// TODO: Improve.
func (e *ExperienceManager) SetProgress(progress float64) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if progress < 0 || progress > 1 {
		return fmt.Errorf("progress must be between 0 and 1, got %f", progress)
	}
	level := e.Level()
	e.experience = experienceForLevels(level) + progressToExperience(level, progress)
	return nil
}

// progressToExperience returns the amount of experience needed for the given level and progress.
func progressToExperience(level int, progress float64) int {
	return int(float64(experienceForLevel(level)) * progress)
}

// experienceToProgress returns the progress towards the next level, calculated using the current level and experience.
func experienceToProgress(experience int, level int) float64 {
	return float64(experience) / float64(experienceForLevel(level))
}

// experienceForLevels calculates the amount of experience needed in total to reach a certain level.
func experienceForLevels(level int) int {
	if level <= 16 {
		return level*level + level*6
	} else if level <= 31 {
		return int(float64(level*level)*2.5 - 40.5*float64(level) + 360)
	}
	return int(float64(level*level)*4.5 - 162.5*float64(level) + 2220)
}

// experienceForLevel returns the amount experience needed to reach level + 1.
func experienceForLevel(level int) int {
	if level <= 15 {
		return 2*level + 7
	} else if level <= 30 {
		return 5*level - 38
	}
	return 9*level - 158
}
