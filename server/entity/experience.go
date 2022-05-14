package entity

import (
	"fmt"
	"math"
	"sync"
)

// ExperienceManager manages experience and levels for entities, and provides functions to add, remove, and calculate
// experience needed for upcoming levels.
type ExperienceManager struct {
	mu       sync.RWMutex
	total    int
	level    int
	progress float64
}

// NewExperienceManager returns a new ExperienceManager with no experience.
func NewExperienceManager() *ExperienceManager {
	return &ExperienceManager{}
}

// Level returns the current experience level.
func (e *ExperienceManager) Level() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.level
}

// Experience returns the amount of experience the manager currently has.
func (e *ExperienceManager) Experience() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return int(float64(experienceForLevel(e.level)) * e.progress)
}

// Progress returns the progress towards the next level, calculated using the current level and experience.
func (e *ExperienceManager) Progress() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.progress
}

// Total returns the total experience collected overall, including levels.
func (e *ExperienceManager) Total() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.total
}

// SetTotal sets the total experience collected.
func (e *ExperienceManager) SetTotal(total int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.total = total
}

// Add adds experience to the total experience.
func (e *ExperienceManager) Add(amount int) (level int, progress float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	amount = int(math.Min(float64(amount), float64(math.MaxInt32-e.total)))
	e.total += amount
	e.level, e.progress = progressFromExperience(experienceForLevels(e.level) + int(float64(experienceForLevel(e.level))*e.progress) + amount)
	return e.level, e.progress
}

// SetLevelAndProgress sets the level and progress of the manager, recalculating the total experience.
func (e *ExperienceManager) SetLevelAndProgress(level int, progress float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if level < 0 || level > math.MaxInt32 {
		panic(fmt.Sprintf("level must be between 0 and 2,147,483,647, got %d", level))
	}
	if progress < 0 || progress > 1 {
		panic(fmt.Sprintf("progress must be between 0 and 1, got %f", progress))
	}
	e.level, e.progress = level, progress
}

// Reset ...
func (e *ExperienceManager) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.total = 0
	e.level = 0
	e.progress = 0
}

// progressFromExperience ...
func progressFromExperience(experience int) (level int, progress float64) {
	var a, b, c float64
	if experience <= experienceForLevels(16) {
		a, b = 1.0, 6.0
	} else if experience <= experienceForLevels(31) {
		a, b, c = 2.5, -40.5, 360.0
	} else {
		a, b, c = 4.5, -162.5, 2220.0
	}

	var sol float64
	if d := b*b - 4*a*(c-float64(experience)); d > 0 {
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
