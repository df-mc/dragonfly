package entity

import "time"

// AttackImmunity keeps track of the brief invulnerability an entity has after taking damage. While a window
// is active, damage up to the amount that armed it is absorbed entirely and only the excess of a stronger
// hit is dealt, as in vanilla. The zero value is an AttackImmunity without an active window.
type AttackImmunity struct {
	until time.Time
	last  float64
}

// Reduce filters damage through the immunity window. If the window is active, the damage is reduced by the
// amount that armed it, and immune is true. Callers deal the returned damage only if it is positive.
func (a *AttackImmunity) Reduce(damage float64) (left float64, immune bool) {
	if !time.Now().Before(a.until) {
		return damage, false
	}
	return damage - a.last, true
}

// Arm starts a new immunity window with the duration passed, during which damage up to the damage value
// passed is absorbed.
func (a *AttackImmunity) Arm(d time.Duration, damage float64) {
	a.until, a.last = time.Now().Add(d), damage
}
