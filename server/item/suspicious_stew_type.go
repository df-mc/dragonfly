package item

import (
	"fmt"
	"time"

	"github.com/df-mc/dragonfly/server/entity/effect"
)

// StewEffect represents a type of suspicious stew.
type StewEffect struct {
	stewEffect
}

// NightVisionStew returns suspicious stew night vision effect.
func NightVisionStew() StewEffect {
	return StewEffect{0}

}

// JumpBoostStew returns suspicious stew jump boost effect.
func JumpBoostStew() StewEffect {
	return StewEffect{1}

}

// WeaknessStew returns suspicious stew weakness effect.
func WeaknessStew() StewEffect {
	return StewEffect{2}

}

// BlindnessStew returns suspicious stew blindness effect.

func BlindnessStew() StewEffect {
	return StewEffect{3}

}

// PoisonStew returns suspicious stew poison effect.
func PoisonStew() StewEffect {
	return StewEffect{4}

}

// SaturationDandelionStew returns suspicious stew saturation effect.
func SaturationDandelionStew() StewEffect {
	return StewEffect{5}

}

// SaturationOrchidStew returns suspicious stew saturation effect.
func SaturationOrchidStew() StewEffect {
	return StewEffect{6}

}

// FireResistanceStew returns suspicious stew fire resistance effect.
func FireResistanceStew() StewEffect {
	return StewEffect{7}

}

// RegenerationStew returns suspicious stew regeneration effect.
func RegenerationStew() StewEffect {
	return StewEffect{8}

}

// WitherStew returns suspicious stew wither effect.
func WitherStew() StewEffect {
	return StewEffect{9}

}

// StewEffects ...
func StewEffects() []StewEffect {
	return []StewEffect{NightVisionStew(), JumpBoostStew(), WeaknessStew(), BlindnessStew(), PoisonStew(), SaturationDandelionStew(), SaturationOrchidStew(), FireResistanceStew(), RegenerationStew(), WitherStew()}

}

type stewEffect uint8

// Uint8 returns the stew as a uint8.
func (s stewEffect) Uint8() uint8 {
	return uint8(s)
}

// Type returns suspicious stew effect type.
func (s stewEffect) Type() []effect.Effect {
	effects := []effect.Effect{}

	switch s.Uint8() {
	case 0:
		effects = append(effects, effect.New(effect.NightVision{}, 1, time.Second*4))
	case 1:
		effects = append(effects, effect.New(effect.JumpBoost{}, 1, time.Second*4))
	case 2:
		effects = append(effects, effect.New(effect.Weakness{}, 1, time.Second*7))
	case 3:
		effects = append(effects, effect.New(effect.Blindness{}, 1, time.Second*6))
	case 4:
		effects = append(effects, effect.New(effect.Poison{}, 1, time.Second*10))
	case 5:
		effects = append(effects, effect.New(effect.Saturation{}, 1, time.Second*3/10))
	case 6:
		effects = append(effects, effect.New(effect.Saturation{}, 1, time.Second*3/10))
	case 7:
		effects = append(effects, effect.New(effect.FireResistance{}, 1, time.Second*2))
	case 8:
		effects = append(effects, effect.New(effect.Regeneration{}, 1, time.Second*6))
	case 9:
		effects = append(effects, effect.New(effect.Wither{}, 1, time.Second*6))
	default:
		panic(fmt.Errorf("invalid stewEffect passed: %v", s.Uint8()))
	}

	return effects
}
