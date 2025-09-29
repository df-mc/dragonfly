package potion

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"time"
)

// Potion holds the effects given by the potion type.
type Potion struct {
	potion
}

func Water() Potion {
	return Potion{}
}

func Mundane() Potion {
	return Potion{1}
}

func LongMundane() Potion {
	return Potion{2}
}

func Thick() Potion {
	return Potion{3}
}

func Awkward() Potion {
	return Potion{4}
}

func NightVision() Potion {
	return Potion{5}
}

func LongNightVision() Potion {
	return Potion{6}
}

func Invisibility() Potion {
	return Potion{7}
}

func LongInvisibility() Potion {
	return Potion{8}
}

func Leaping() Potion {
	return Potion{9}
}

func LongLeaping() Potion {
	return Potion{10}
}

func StrongLeaping() Potion {
	return Potion{11}
}

func FireResistance() Potion {
	return Potion{12}
}

func LongFireResistance() Potion {
	return Potion{13}
}

func Swiftness() Potion {
	return Potion{14}
}

func LongSwiftness() Potion {
	return Potion{15}
}

func StrongSwiftness() Potion {
	return Potion{16}
}

func Slowness() Potion {
	return Potion{17}
}

func LongSlowness() Potion {
	return Potion{18}
}

func WaterBreathing() Potion {
	return Potion{19}
}

func LongWaterBreathing() Potion {
	return Potion{20}
}

func Healing() Potion {
	return Potion{21}
}

func StrongHealing() Potion {
	return Potion{22}
}

func Harming() Potion {
	return Potion{23}
}

func StrongHarming() Potion {
	return Potion{24}
}

func Poison() Potion {
	return Potion{25}
}

func LongPoison() Potion {
	return Potion{26}
}

func StrongPoison() Potion {
	return Potion{27}
}

func Regeneration() Potion {
	return Potion{28}
}

func LongRegeneration() Potion {
	return Potion{29}
}

func StrongRegeneration() Potion {
	return Potion{30}
}

func Strength() Potion {
	return Potion{31}
}

func LongStrength() Potion {
	return Potion{32}
}

func StrongStrength() Potion {
	return Potion{33}
}

func Weakness() Potion {
	return Potion{34}
}

func LongWeakness() Potion {
	return Potion{35}
}

func Wither() Potion {
	return Potion{36}
}

func TurtleMaster() Potion {
	return Potion{37}
}

func LongTurtleMaster() Potion {
	return Potion{38}
}

func StrongTurtleMaster() Potion {
	return Potion{39}
}

func SlowFalling() Potion {
	return Potion{40}
}

func LongSlowFalling() Potion {
	return Potion{41}
}

func StrongSlowness() Potion {
	return Potion{42}
}

// From returns a Potion by the ID given.
func From(id int32) Potion {
	return Potion{potion(id)}
}

// Effects returns the effects of the potion.
func (p Potion) Effects() []effect.Effect {
	switch p {
	case NightVision():
		return []effect.Effect{effect.New(effect.NightVision, 1, 3*time.Minute)}
	case LongNightVision():
		return []effect.Effect{effect.New(effect.NightVision, 1, 8*time.Minute)}
	case Invisibility():
		return []effect.Effect{effect.New(effect.Invisibility, 1, 3*time.Minute)}
	case LongInvisibility():
		return []effect.Effect{effect.New(effect.Invisibility, 1, 8*time.Minute)}
	case Leaping():
		return []effect.Effect{effect.New(effect.JumpBoost, 1, 3*time.Minute)}
	case LongLeaping():
		return []effect.Effect{effect.New(effect.JumpBoost, 1, 8*time.Minute)}
	case StrongLeaping():
		return []effect.Effect{effect.New(effect.JumpBoost, 2, 90*time.Second)}
	case FireResistance():
		return []effect.Effect{effect.New(effect.FireResistance, 1, 3*time.Minute)}
	case LongFireResistance():
		return []effect.Effect{effect.New(effect.FireResistance, 1, 8*time.Minute)}
	case Swiftness():
		return []effect.Effect{effect.New(effect.Speed, 1, 3*time.Minute)}
	case LongSwiftness():
		return []effect.Effect{effect.New(effect.Speed, 1, 8*time.Minute)}
	case StrongSwiftness():
		return []effect.Effect{effect.New(effect.Speed, 2, 90*time.Second)}
	case Slowness():
		return []effect.Effect{effect.New(effect.Slowness, 1, 90*time.Second)}
	case LongSlowness():
		return []effect.Effect{effect.New(effect.Slowness, 1, 4*time.Minute)}
	case StrongSlowness():
		return []effect.Effect{effect.New(effect.Slowness, 4, 20*time.Second)}
	case WaterBreathing():
		return []effect.Effect{effect.New(effect.WaterBreathing, 1, 3*time.Minute)}
	case LongWaterBreathing():
		return []effect.Effect{effect.New(effect.WaterBreathing, 1, 8*time.Minute)}
	case Healing():
		return []effect.Effect{effect.NewInstant(effect.InstantHealth, 1)}
	case StrongHealing():
		return []effect.Effect{effect.NewInstant(effect.InstantHealth, 2)}
	case Harming():
		return []effect.Effect{effect.NewInstant(effect.InstantDamage, 1)}
	case StrongHarming():
		return []effect.Effect{effect.NewInstant(effect.InstantDamage, 2)}
	case Poison():
		return []effect.Effect{effect.New(effect.Poison, 1, 45*time.Second)}
	case LongPoison():
		return []effect.Effect{effect.New(effect.Poison, 1, 2*time.Minute)}
	case StrongPoison():
		return []effect.Effect{effect.New(effect.Poison, 2, 22500*time.Millisecond)}
	case Regeneration():
		return []effect.Effect{effect.New(effect.Regeneration, 1, 45*time.Second)}
	case LongRegeneration():
		return []effect.Effect{effect.New(effect.Regeneration, 1, 2*time.Minute)}
	case StrongRegeneration():
		return []effect.Effect{effect.New(effect.Regeneration, 2, 22500*time.Millisecond)}
	case Strength():
		return []effect.Effect{effect.New(effect.Strength, 1, 3*time.Minute)}
	case LongStrength():
		return []effect.Effect{effect.New(effect.Strength, 1, 8*time.Minute)}
	case StrongStrength():
		return []effect.Effect{effect.New(effect.Strength, 2, 90*time.Second)}
	case Weakness():
		return []effect.Effect{effect.New(effect.Weakness, 1, 90*time.Second)}
	case LongWeakness():
		return []effect.Effect{effect.New(effect.Weakness, 1, 4*time.Minute)}
	case Wither():
		return []effect.Effect{effect.New(effect.Wither, 1, 40*time.Second)}
	case TurtleMaster():
		return []effect.Effect{
			effect.New(effect.Resistance, 3, 20*time.Second),
			effect.New(effect.Slowness, 4, 20*time.Second),
		}
	case LongTurtleMaster():
		return []effect.Effect{
			effect.New(effect.Resistance, 3, 40*time.Second),
			effect.New(effect.Slowness, 4, 40*time.Second),
		}
	case StrongTurtleMaster():
		return []effect.Effect{
			effect.New(effect.Resistance, 5, 20*time.Second),
			effect.New(effect.Slowness, 6, 20*time.Second),
		}
	case SlowFalling():
		return []effect.Effect{effect.New(effect.SlowFalling, 1, 90*time.Second)}
	case LongSlowFalling():
		return []effect.Effect{effect.New(effect.SlowFalling, 1, 4*time.Minute)}
	}
	return []effect.Effect{}
}

type potion uint8

// Uint8 returns the potion type as a uint8.
func (p potion) Uint8() uint8 {
	return uint8(p)
}

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
