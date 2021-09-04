package entity

import (
	"math"
	"sync"
)

var maxLevel int32 = math.MaxInt32

type XPManager struct {
	mu       sync.RWMutex
	level    int32
	progress float64
	totalXP  int
}

// NewXPManager return a XPManager with default value of level, progress and totalXP that are equal to 0.
func NewXPManager() *XPManager {
	return &XPManager{}
}

// Level get xp level.
func (e *XPManager) Level() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.level
}

// Progress get the progress of the xp, the value are Between 0.00 and 1.00.
func (e *XPManager) Progress() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.progress
}

// MaxLevel get the max xp level.
func (e *XPManager) MaxLevel() int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return maxLevel
}

// TotalXP get the total xp collected.
func (e *XPManager) TotalXP() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.totalXP
}

// XPToNextLevel get the total XP netted to level up.
func (e *XPManager) XPToNextLevel() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.XPForLevel(e.level)
}

// XPNeededToLevelUp get xp that you need to level up with reducing xp that you already have.
func (e *XPManager) XPNeededToLevelUp() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.XPToNextLevel() - e.ProgressToXp(e.progress, e.level)
}

// XPForLevel get the amount xp to level up in a specific level.
func (e *XPManager) XPForLevel(level int32) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// I have make like MINET
	if level > 30 {
		return int(9*level - 158)
	} else if level > 15 {
		return int(5*level - 38)
	} else {
		return int(2*level + 7)
	}
}

// XPForLevels calculate amount of xp for reach a level.
func (e *XPManager) XPForLevels(level int32) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	xp := 0
	for i := int32(0); i <= level; i++ {
		xp += e.XPForLevel(i)
	}
	return xp
}

// ProgressToXp get number of xp that you have (not total but start on level) whit progress.
func (e *XPManager) ProgressToXp(progress float64, level int32) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return int(float64(e.XPForLevel(level)) * progress)
}

// XpToProgress get progress with the amount of xp and a level.
func (e *XPManager) XpToProgress(xp int, level int32) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return float64(xp) / float64(e.XPForLevel(level))
}

// AddXP add xp, that check if you level up.
func (e *XPManager) AddXP(amount int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// level up
	for e.XPNeededToLevelUp()-amount <= 0 {
		amount -= e.XPToNextLevel() - e.XPNeededToLevelUp()
		if e.level == maxLevel {
			return
		}
		e.level++
	}
	e.progress = e.XpToProgress(amount, e.level)
}

// SetLevel set a new level, this recalculates the total xp.
func (e *XPManager) SetLevel(level int32) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if level < 0 {
		level = 0
	} else if level > maxLevel {
		level = maxLevel
	}
	e.level = level
	e.totalXP = e.XPForLevels(level) + e.ProgressToXp(e.progress, level)
}

// SetProgress set a new progress, this recalculates the total xp.
func (e *XPManager) SetProgress(progress float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
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
