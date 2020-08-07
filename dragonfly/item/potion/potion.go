package potion

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
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
	return Potion{5, []effect.Effect{effect.NightVision{}.WithSettings(time.Duration(3)*time.Minute, 1, false)}}
}

// LongNightVision ...
func LongNightVision() Potion {
	return Potion{6, []effect.Effect{effect.NightVision{}.WithSettings(time.Duration(8)*time.Minute, 1, false)}}
}

// Invisibility ...
func Invisibility() Potion {
	return Potion{7, []effect.Effect{effect.Invisibility{}.WithSettings(time.Duration(3)*time.Minute, 1, false)}}
}

// LongInvisibility ...
func LongInvisibility() Potion {
	return Potion{8, []effect.Effect{effect.Invisibility{}.WithSettings(time.Duration(8)*time.Minute, 1, false)}}
}

// Leaping ...
func Leaping() Potion {
	return Potion{9, []effect.Effect{effect.JumpBoost{}.WithSettings(time.Duration(3)*time.Minute, 1, false)}}
}

// LongLeaping ...
func LongLeaping() Potion {
	return Potion{10, []effect.Effect{effect.JumpBoost{}.WithSettings(time.Duration(8)*time.Minute, 1, false)}}
}

// StrongLeaping ...
func StrongLeaping() Potion {
	return Potion{11, []effect.Effect{effect.JumpBoost{}.WithSettings(time.Duration(90)*time.Second, 2, false)}}
}

// FireResistance ...
func FireResistance() Potion {
	return Potion{potion: 12} //TODO: Implement fire resistance
}

// LongFireResistance ...
func LongFireResistance() Potion {
	return Potion{potion: 13} //TODO: Implement fire resistance
}

// Swiftness ...
func Swiftness() Potion {
	return Potion{14, []effect.Effect{effect.Speed{}.WithSettings(time.Duration(3)*time.Minute, 1, false)}}
}

// LongSwiftness ...
func LongSwiftness() Potion {
	return Potion{15, []effect.Effect{effect.Speed{}.WithSettings(time.Duration(8)*time.Minute, 1, false)}}
}

// StrongSwiftness ...
func StrongSwiftness() Potion {
	return Potion{16, []effect.Effect{effect.Speed{}.WithSettings(time.Duration(90)*time.Second, 2, false)}}
}

// Slowness ...
func Slowness() Potion {
	return Potion{17, []effect.Effect{effect.Slowness{}.WithSettings(time.Duration(90)*time.Second, 1, false)}}
}

// LongSlowness ...
func LongSlowness() Potion {
	return Potion{18, []effect.Effect{effect.Slowness{}.WithSettings(time.Duration(4)*time.Minute, 1, false)}}
}

// WaterBreathing ...
func WaterBreathing() Potion {
	return Potion{19, []effect.Effect{effect.WaterBreathing{}.WithSettings(time.Duration(3)*time.Minute, 1, false)}}
}

// LongWaterBreathing ...
func LongWaterBreathing() Potion {
	return Potion{20, []effect.Effect{effect.WaterBreathing{}.WithSettings(time.Duration(8)*time.Minute, 1, false)}}
}

// Healing ...
func Healing() Potion {
	return Potion{21, []effect.Effect{effect.InstantHealth{}.WithSettings(time.Duration(0), 1, false)}}
}

// StrongHealing ...
func StrongHealing() Potion {
	return Potion{22, []effect.Effect{effect.InstantHealth{}.WithSettings(time.Duration(0), 2, false)}}
}

// Harming ...
func Harming() Potion {
	return Potion{23, []effect.Effect{effect.InstantDamage{}.WithSettings(time.Duration(0), 1, false)}}
}

// StrongHarming ...
func StrongHarming() Potion {
	return Potion{24, []effect.Effect{effect.InstantDamage{}.WithSettings(time.Duration(0), 2, false)}}
}

// Poison ...
func Poison() Potion {
	return Potion{25, []effect.Effect{effect.Poison{}.WithSettings(time.Duration(45)*time.Second, 1, false)}}
}

// LongPoison ...
func LongPoison() Potion {
	return Potion{26, []effect.Effect{effect.Poison{}.WithSettings(time.Duration(2)*time.Minute, 1, false)}}
}

// StrongPoison ...
func StrongPoison() Potion {
	return Potion{27, []effect.Effect{effect.Poison{}.WithSettings(time.Duration(22500)*time.Millisecond, 2, false)}}
}

// Regeneration ...
func Regeneration() Potion {
	return Potion{28, []effect.Effect{effect.Regeneration{}.WithSettings(time.Duration(45)*time.Second, 1, false)}}
}

// LongRegeneration ...
func LongRegeneration() Potion {
	return Potion{29, []effect.Effect{effect.Regeneration{}.WithSettings(time.Duration(2)*time.Minute, 1, false)}}
}

// StrongRegeneration ...
func StrongRegeneration() Potion {
	return Potion{30, []effect.Effect{effect.Regeneration{}.WithSettings(time.Duration(22)*time.Second, 2, false)}}
}

// Strength ...
func Strength() Potion {
	return Potion{31, []effect.Effect{effect.Strength{}.WithSettings(time.Duration(3)*time.Minute, 1, false)}}
}

// LongStrength ...
func LongStrength() Potion {
	return Potion{32, []effect.Effect{effect.Strength{}.WithSettings(time.Duration(8)*time.Minute, 1, false)}}
}

// StrongStrength ...
func StrongStrength() Potion {
	return Potion{33, []effect.Effect{effect.Strength{}.WithSettings(time.Duration(90)*time.Second, 2, false)}}
}

// Weakness ...
func Weakness() Potion {
	return Potion{34, []effect.Effect{effect.Weakness{}.WithSettings(time.Duration(90)*time.Second, 1, false)}}
}

// LongWeakness ...
func LongWeakness() Potion {
	return Potion{35, []effect.Effect{effect.Weakness{}.WithSettings(time.Duration(4)*time.Minute, 1, false)}}
}

// Wither ...
func Wither() Potion {
	return Potion{36, []effect.Effect{effect.Wither{}.WithSettings(time.Duration(40)*time.Second, 1, false)}}
}

// TurtleMaster ...
func TurtleMaster() Potion {
	return Potion{37, []effect.Effect{
		effect.Resistance{}.WithSettings(time.Duration(20)*time.Second, 3, false),
		effect.Slowness{}.WithSettings(time.Duration(20)*time.Second, 4, false),
	}}
}

// LongTurtleMaster ...
func LongTurtleMaster() Potion {
	return Potion{38, []effect.Effect{
		effect.Resistance{}.WithSettings(time.Duration(40)*time.Second, 3, false),
		effect.Slowness{}.WithSettings(time.Duration(40)*time.Second, 4, false),
	}}
}

// StrongTurtleMaster ...
func StrongTurtleMaster() Potion {
	return Potion{39, []effect.Effect{
		effect.Resistance{}.WithSettings(time.Duration(20)*time.Second, 5, false),
		effect.Slowness{}.WithSettings(time.Duration(20)*time.Second, 6, false),
	}}
}

// SlowFalling ...
func SlowFalling() Potion {
	return Potion{40, []effect.Effect{effect.SlowFalling{}.WithSettings(time.Duration(90)*time.Second, 1, false)}}
}

// LongSlowFalling ...
func LongSlowFalling() Potion {
	return Potion{41, []effect.Effect{effect.SlowFalling{}.WithSettings(time.Duration(4)*time.Minute, 1, false)}}
}

// StrongSlowness ...
func StrongSlowness() Potion {
	return Potion{42, []effect.Effect{effect.Slowness{}.WithSettings(time.Duration(20)*time.Second, 4, false)}}
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
