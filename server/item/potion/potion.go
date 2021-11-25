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
	return Potion{5, []effect.Effect{effect.New(effect.NightVision{}, 1, 3*time.Minute)}}
}

// LongNightVision ...
func LongNightVision() Potion {
	return Potion{6, []effect.Effect{effect.New(effect.NightVision{}, 1, 8*time.Minute)}}
}

// Invisibility ...
func Invisibility() Potion {
	return Potion{7, []effect.Effect{effect.New(effect.Invisibility{}, 1, 3*time.Minute)}}
}

// LongInvisibility ...
func LongInvisibility() Potion {
	return Potion{8, []effect.Effect{effect.New(effect.Invisibility{}, 1, 8*time.Minute)}}
}

// Leaping ...
func Leaping() Potion {
	return Potion{9, []effect.Effect{effect.New(effect.JumpBoost{}, 1, 3*time.Minute)}}
}

// LongLeaping ...
func LongLeaping() Potion {
	return Potion{10, []effect.Effect{effect.New(effect.JumpBoost{}, 1, 8*time.Minute)}}
}

// StrongLeaping ...
func StrongLeaping() Potion {
	return Potion{11, []effect.Effect{effect.New(effect.JumpBoost{}, 2, 90*time.Second)}}
}

// FireResistance ...
func FireResistance() Potion {
	return Potion{12, []effect.Effect{effect.New(effect.FireResistance{}, 1, 3*time.Minute)}}
}

// LongFireResistance ...
func LongFireResistance() Potion {
	return Potion{13, []effect.Effect{effect.New(effect.FireResistance{}, 1, 8*time.Minute)}}
}

// Swiftness ...
func Swiftness() Potion {
	return Potion{14, []effect.Effect{effect.New(effect.Speed{}, 1, 3*time.Minute)}}
}

// LongSwiftness ...
func LongSwiftness() Potion {
	return Potion{15, []effect.Effect{effect.New(effect.Speed{}, 1, 8*time.Minute)}}
}

// StrongSwiftness ...
func StrongSwiftness() Potion {
	return Potion{16, []effect.Effect{effect.New(effect.Speed{}, 2, 90*time.Second)}}
}

// Slowness ...
func Slowness() Potion {
	return Potion{17, []effect.Effect{effect.New(effect.Slowness{}, 1, 90*time.Second)}}
}

// LongSlowness ...
func LongSlowness() Potion {
	return Potion{18, []effect.Effect{effect.New(effect.Slowness{}, 1, 4*time.Minute)}}
}

// WaterBreathing ...
func WaterBreathing() Potion {
	return Potion{19, []effect.Effect{effect.New(effect.WaterBreathing{}, 1, 3*time.Minute)}}
}

// LongWaterBreathing ...
func LongWaterBreathing() Potion {
	return Potion{20, []effect.Effect{effect.New(effect.WaterBreathing{}, 1, 8*time.Minute)}}
}

// Healing ...
func Healing() Potion {
	return Potion{21, []effect.Effect{effect.NewInstant(effect.InstantHealth{}, 1)}}
}

// StrongHealing ...
func StrongHealing() Potion {
	return Potion{22, []effect.Effect{effect.NewInstant(effect.InstantHealth{}, 2)}}
}

// Harming ...
func Harming() Potion {
	return Potion{23, []effect.Effect{effect.NewInstant(effect.InstantDamage{}, 1)}}
}

// StrongHarming ...
func StrongHarming() Potion {
	return Potion{24, []effect.Effect{effect.NewInstant(effect.InstantDamage{}, 2)}}
}

// Poison ...
func Poison() Potion {
	return Potion{25, []effect.Effect{effect.New(effect.Poison{}, 1, 45*time.Second)}}
}

// LongPoison ...
func LongPoison() Potion {
	return Potion{26, []effect.Effect{effect.New(effect.Poison{}, 1, 2*time.Minute)}}
}

// StrongPoison ...
func StrongPoison() Potion {
	return Potion{27, []effect.Effect{effect.New(effect.Poison{}, 2, 22500*time.Millisecond)}}
}

// Regeneration ...
func Regeneration() Potion {
	return Potion{28, []effect.Effect{effect.New(effect.Regeneration{}, 1, 45*time.Second)}}
}

// LongRegeneration ...
func LongRegeneration() Potion {
	return Potion{29, []effect.Effect{effect.New(effect.Regeneration{}, 1, 2*time.Minute)}}
}

// StrongRegeneration ...
func StrongRegeneration() Potion {
	return Potion{30, []effect.Effect{effect.New(effect.Regeneration{}, 2, 22*time.Second)}}
}

// Strength ...
func Strength() Potion {
	return Potion{31, []effect.Effect{effect.New(effect.Strength{}, 1, 3*time.Minute)}}
}

// LongStrength ...
func LongStrength() Potion {
	return Potion{32, []effect.Effect{effect.New(effect.Strength{}, 1, 8*time.Minute)}}
}

// StrongStrength ...
func StrongStrength() Potion {
	return Potion{33, []effect.Effect{effect.New(effect.Strength{}, 2, 90*time.Second)}}
}

// Weakness ...
func Weakness() Potion {
	return Potion{34, []effect.Effect{effect.New(effect.Weakness{}, 1, 90*time.Second)}}
}

// LongWeakness ...
func LongWeakness() Potion {
	return Potion{35, []effect.Effect{effect.New(effect.Weakness{}, 1, 4*time.Minute)}}
}

// Wither ...
func Wither() Potion {
	return Potion{36, []effect.Effect{effect.New(effect.Wither{}, 1, 40*time.Second)}}
}

// TurtleMaster ...
func TurtleMaster() Potion {
	return Potion{37, []effect.Effect{
		effect.New(effect.Resistance{}, 3, 20*time.Second),
		effect.New(effect.Slowness{}, 4, 20*time.Second),
	}}
}

// LongTurtleMaster ...
func LongTurtleMaster() Potion {
	return Potion{38, []effect.Effect{
		effect.New(effect.Resistance{}, 3, 40*time.Second),
		effect.New(effect.Slowness{}, 4, 40*time.Second),
	}}
}

// StrongTurtleMaster ...
func StrongTurtleMaster() Potion {
	return Potion{39, []effect.Effect{
		effect.New(effect.Resistance{}, 5, 20*time.Second),
		effect.New(effect.Slowness{}, 6, 20*time.Second),
	}}
}

// SlowFalling ...
func SlowFalling() Potion {
	return Potion{40, []effect.Effect{effect.New(effect.SlowFalling{}, 1, 90*time.Second)}}
}

// LongSlowFalling ...
func LongSlowFalling() Potion {
	return Potion{41, []effect.Effect{effect.New(effect.SlowFalling{}, 1, 4*time.Minute)}}
}

// StrongSlowness ...
func StrongSlowness() Potion {
	return Potion{42, []effect.Effect{effect.New(effect.Slowness{}, 4, 20*time.Second)}}
}

// Equals ...
func (p Potion) Equals(other Potion) bool {
	return p.Uint8() == other.Uint8()
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
