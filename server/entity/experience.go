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
	d          float64
}

// NewExperienceManager returns a new ExperienceManager with no experience.
func NewExperienceManager() *ExperienceManager {
	return &ExperienceManager{}
}

// Experience returns the amount of experience the manager currently has.
func (e *ExperienceManager) Experience() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.experience
}

// Add adds experience to the total experience and recalculates the level and progress if necessary.
func (e *ExperienceManager) Add(amount int) (level int, progress float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.experience += amount
	return progressFromExperience(e.total())
}

// Remove removes experience from the total experience and recalculates the level and progress if necessary.
func (e *ExperienceManager) Remove(amount int) (level int, progress float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.experience -= amount
	return progressFromExperience(e.total())
}

// total returns the total amount of experience including the extra decimals provided for more accuracy.
func (e *ExperienceManager) total() float64 {
	return float64(e.experience) + e.d
}

// Level returns the current experience level.
func (e *ExperienceManager) Level() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	level, _ := progressFromExperience(e.total())
	return level
}

// SetLevel sets the level of the manager.
func (e *ExperienceManager) SetLevel(level int) {
	if level < 0 || level > math.MaxInt32 {
		panic(fmt.Sprintf("level must be between 0 and 2,147,483,647, got %d", level))
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	_, progress := progressFromExperience(e.total())
	e.experience = experienceForLevels(level) + int(float64(experienceForLevel(level))*progress)
}

// Progress returns the progress towards the next level.
func (e *ExperienceManager) Progress() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, progress := progressFromExperience(e.total())
	return progress
}

// SetProgress sets the progress of the manager.
func (e *ExperienceManager) SetProgress(progress float64) {
	if progress < 0 || progress > 1 {
		panic(fmt.Sprintf("progress must be between 0 and 1, got %f", progress))
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	currentLevel, _ := progressFromExperience(e.total())
	progressExp := float64(experienceForLevel(currentLevel)) * progress
	e.experience = experienceForLevels(currentLevel) + int(progressExp)
	e.d = progressExp - math.Trunc(progressExp)
}

// Reset resets the total experience, level, and progress of the manager to zero.
func (e *ExperienceManager) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.experience, e.d = 0, 0
}

// progressFromExperience returns the level and progress from the total experience given.
func progressFromExperience(experience float64) (level int, progress float64) {
	var a, b, c float64
	if experience <= float64(experienceForLevels(16)) {
		a, b = 1.0, 6.0
	} else if experience <= float64(experienceForLevels(31)) {
		a, b, c = 2.5, -40.5, 360.0
	} else {
		a, b, c = 4.5, -162.5, 2220.0
	}

	var sol float64
	if d := b*b - 4*a*(c-experience); d > 0 {
		s := math.Sqrt(d)
		sol = math.Max((-b+s)/(2*a), (-b-s)/(2*a))
	} else if d == 0 {
		sol = -b / (2 * a)
	}
	return int(sol), sol - math.Trunc(sol)
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
