package potion

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"time"
)

// Potion holds the effects given by the potion type.
type Potion struct {
	potion
	Effects []effect.Effect
}

// Water ...
func Water() Potion {
	return Potion{}
}

// Mundane ...
func Mundane() Potion {
	return Potion{potion: 1}
}

// LongMundane ...
func LongMundane() Potion {
	return Potion{potion: 2}
}

// Thick ...
func Thick() Potion {
	return Potion{potion: 3}
}

// Awkward ...
func Awkward() Potion {
	return Potion{potion: 4}
}

// NightVision ...
func NightVision() Potion {
	return Potion{5, []effect.Effect{effect.NightVision{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongNightVision ...
func LongNightVision() Potion {
	return Potion{6, []effect.Effect{effect.NightVision{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// Invisibility ...
func Invisibility() Potion {
	return Potion{7, []effect.Effect{effect.Invisibility{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongInvisibility ...
func LongInvisibility() Potion {
	return Potion{8, []effect.Effect{effect.Invisibility{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// Leaping ...
func Leaping() Potion {
	return Potion{9, []effect.Effect{effect.JumpBoost{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongLeaping ...
func LongLeaping() Potion {
	return Potion{10, []effect.Effect{effect.JumpBoost{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// StrongLeaping ...
func StrongLeaping() Potion {
	return Potion{11, []effect.Effect{effect.JumpBoost{}.WithSettings(90*time.Second, 2, false, false)}}
}

// FireResistance ...
func FireResistance() Potion {
	return Potion{12, []effect.Effect{effect.FireResistance{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongFireResistance ...
func LongFireResistance() Potion {
	return Potion{13, []effect.Effect{effect.FireResistance{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// Swiftness ...
func Swiftness() Potion {
	return Potion{14, []effect.Effect{effect.Speed{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongSwiftness ...
func LongSwiftness() Potion {
	return Potion{15, []effect.Effect{effect.Speed{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// StrongSwiftness ...
func StrongSwiftness() Potion {
	return Potion{16, []effect.Effect{effect.Speed{}.WithSettings(90*time.Second, 2, false, false)}}
}

// Slowness ...
func Slowness() Potion {
	return Potion{17, []effect.Effect{effect.Slowness{}.WithSettings(90*time.Second, 1, false, false)}}
}

// LongSlowness ...
func LongSlowness() Potion {
	return Potion{18, []effect.Effect{effect.Slowness{}.WithSettings(4*time.Minute, 1, false, false)}}
}

// WaterBreathing ...
func WaterBreathing() Potion {
	return Potion{19, []effect.Effect{effect.WaterBreathing{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongWaterBreathing ...
func LongWaterBreathing() Potion {
	return Potion{20, []effect.Effect{effect.WaterBreathing{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// Healing ...
func Healing() Potion {
	return Potion{21, []effect.Effect{effect.InstantHealth{}.WithSettings(0, 1, false, false)}}
}

// StrongHealing ...
func StrongHealing() Potion {
	return Potion{22, []effect.Effect{effect.InstantHealth{}.WithSettings(0, 2, false, false)}}
}

// Harming ...
func Harming() Potion {
	return Potion{23, []effect.Effect{effect.InstantDamage{}.WithSettings(0, 1, false, false)}}
}

// StrongHarming ...
func StrongHarming() Potion {
	return Potion{24, []effect.Effect{effect.InstantDamage{}.WithSettings(0, 2, false, false)}}
}

// Poison ...
func Poison() Potion {
	return Potion{25, []effect.Effect{effect.Poison{}.WithSettings(45*time.Second, 1, false, false)}}
}

// LongPoison ...
func LongPoison() Potion {
	return Potion{26, []effect.Effect{effect.Poison{}.WithSettings(2*time.Minute, 1, false, false)}}
}

// StrongPoison ...
func StrongPoison() Potion {
	return Potion{27, []effect.Effect{effect.Poison{}.WithSettings(22500*time.Millisecond, 2, false, false)}}
}

// Regeneration ...
func Regeneration() Potion {
	return Potion{28, []effect.Effect{effect.Regeneration{}.WithSettings(45*time.Second, 1, false, false)}}
}

// LongRegeneration ...
func LongRegeneration() Potion {
	return Potion{29, []effect.Effect{effect.Regeneration{}.WithSettings(2*time.Minute, 1, false, false)}}
}

// StrongRegeneration ...
func StrongRegeneration() Potion {
	return Potion{30, []effect.Effect{effect.Regeneration{}.WithSettings(22*time.Second, 2, false, false)}}
}

// Strength ...
func Strength() Potion {
	return Potion{31, []effect.Effect{effect.Strength{}.WithSettings(3*time.Minute, 1, false, false)}}
}

// LongStrength ...
func LongStrength() Potion {
	return Potion{32, []effect.Effect{effect.Strength{}.WithSettings(8*time.Minute, 1, false, false)}}
}

// StrongStrength ...
func StrongStrength() Potion {
	return Potion{33, []effect.Effect{effect.Strength{}.WithSettings(90*time.Second, 2, false, false)}}
}

// Weakness ...
func Weakness() Potion {
	return Potion{34, []effect.Effect{effect.Weakness{}.WithSettings(90*time.Second, 1, false, false)}}
}

// LongWeakness ...
func LongWeakness() Potion {
	return Potion{35, []effect.Effect{effect.Weakness{}.WithSettings(4*time.Minute, 1, false, false)}}
}

// Wither ...
func Wither() Potion {
	return Potion{36, []effect.Effect{effect.Wither{}.WithSettings(40*time.Second, 1, false, false)}}
}

// TurtleMaster ...
func TurtleMaster() Potion {
	return Potion{37, []effect.Effect{
		effect.Resistance{}.WithSettings(20*time.Second, 3, false, false),
		effect.Slowness{}.WithSettings(20*time.Second, 4, false, false),
	}}
}

// LongTurtleMaster ...
func LongTurtleMaster() Potion {
	return Potion{38, []effect.Effect{
		effect.Resistance{}.WithSettings(40*time.Second, 3, false, false),
		effect.Slowness{}.WithSettings(40*time.Second, 4, false, false),
	}}
}

// StrongTurtleMaster ...
func StrongTurtleMaster() Potion {
	return Potion{39, []effect.Effect{
		effect.Resistance{}.WithSettings(20*time.Second, 5, false, false),
		effect.Slowness{}.WithSettings(20*time.Second, 6, false, false),
	}}
}

// SlowFalling ...
func SlowFalling() Potion {
	return Potion{40, []effect.Effect{effect.SlowFalling{}.WithSettings(90*time.Second, 1, false, false)}}
}

// LongSlowFalling ...
func LongSlowFalling() Potion {
	return Potion{41, []effect.Effect{effect.SlowFalling{}.WithSettings(4*time.Minute, 1, false, false)}}
}

// StrongSlowness ...
func StrongSlowness() Potion {
	return Potion{42, []effect.Effect{effect.Slowness{}.WithSettings(20*time.Second, 4, false, false)}}
}

type potion uint8

// Uint8 returns the potion type as a uint8.
func (p potion) Uint8() uint8 {
	return uint8(p)
}

// All ...
func All() []Potion {
	return []Potion{
		Water(), Mundane(), LongMundane(), Thick(), Awkward(), NightVision(), LongNightVision(), Invisibility(),
		LongInvisibility(), Leaping(), LongLeaping(), StrongLeaping(), FireResistance(), LongFireResistance(),
		Swiftness(), LongSwiftness(), StrongSwiftness(), Slowness(), LongSlowness(), WaterBreathing(),
		LongWaterBreathing(), Healing(), StrongHealing(), Harming(), StrongHarming(), Poison(), LongPoison(),
		StrongPoison(), Regeneration(), LongRegeneration(), StrongRegeneration(), Strength(), LongStrength(),
		StrongStrength(), Weakness(), LongWeakness(), Wither(), TurtleMaster(), LongTurtleMaster(), StrongTurtleMaster(),
		SlowFalling(), LongSlowFalling(), StrongSlowness(),
	}
}
