package item

import (
	"time"

	"github.com/df-mc/dragonfly/server/entity/effect"
)

// StewEffect represents a type of suspicious stew.
type StewEffect struct {
	effect effect.Effect
	stewEffect
}

// Type returns suspicious stew effect type.
func (s StewEffect) Type() effect.Effect {
	return s.effect

}

// NightVisionStew returns suspicious stew night vision effect.
func NightVisionStew() StewEffect {
	return StewEffect{effect.New(effect.NightVision{}, 1, time.Second*4), 0}

}

// JumpBoostStew returns suspicious stew jump boost effect.
func JumpBoostStew() StewEffect {
	return StewEffect{effect.New(effect.JumpBoost{}, 1, time.Second*4), 1}

}

// WeaknessStew returns suspicious stew weakness effect.
func WeaknessStew() StewEffect {
	return StewEffect{effect.New(effect.Weakness{}, 1, time.Second*7), 2}

}

// BlindnessStew returns suspicious stew blindness effect.

func BlindnessStew() StewEffect {
	return StewEffect{effect.New(effect.Blindness{}, 1, time.Second*6), 3}

}

// PoisonStew returns suspicious stew poison effect.
func PoisonStew() StewEffect {
	return StewEffect{effect.New(effect.Poison{}, 1, time.Second*10), 4}

}

// SaturationDandelionStew returns suspicious stew saturation effect.
func SaturationDandelionStew() StewEffect {
	return StewEffect{effect.New(effect.Saturation{}, 1, time.Second*3/10), 5}

}

// SaturationOrchidStew returns suspicious stew saturation effect.
func SaturationOrchidStew() StewEffect {
	return StewEffect{effect.New(effect.Saturation{}, 1, time.Second*3/10), 6}

}

// FireResistanceStew returns suspicious stew fire resistance effect.
func FireResistanceStew() StewEffect {
	return StewEffect{effect.New(effect.FireResistance{}, 1, time.Second*2), 7}

}

// RegenerationStew returns suspicious stew regeneration effect.
func RegenerationStew() StewEffect {
	return StewEffect{effect.New(effect.Regeneration{}, 1, time.Second*6), 8}

}

// WitherStew returns suspicious stew wither effect.
func WitherStew() StewEffect {
	return StewEffect{effect.New(effect.Wither{}, 1, time.Second*6), 9}

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
