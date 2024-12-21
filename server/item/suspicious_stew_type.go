package item

import (
	"time"

	"github.com/df-mc/dragonfly/server/entity/effect"
)

// StewType represents a type of suspicious stew.
type StewType struct {
	stewType
}

// NightVisionPoppyStew returns suspicious stew night vision effect.
func NightVisionPoppyStew() StewType {
	return StewType{0}
}

// JumpBoostStew returns suspicious stew jump boost effect.
func JumpBoostStew() StewType {
	return StewType{1}
}

// WeaknessStew returns suspicious stew weakness effect.
func WeaknessStew() StewType {
	return StewType{2}
}

// BlindnessBluetStew returns suspicious stew blindness effect.
func BlindnessBluetStew() StewType {
	return StewType{3}
}

// PoisonStew returns suspicious stew poison effect.
func PoisonStew() StewType {
	return StewType{4}
}

// SaturationDandelionStew returns suspicious stew saturation effect.
func SaturationDandelionStew() StewType {
	return StewType{5}
}

// SaturationOrchidStew returns suspicious stew saturation effect.
func SaturationOrchidStew() StewType {
	return StewType{6}
}

// FireResistanceStew returns suspicious stew fire resistance effect.
func FireResistanceStew() StewType {
	return StewType{7}
}

// RegenerationStew returns suspicious stew regeneration effect.
func RegenerationStew() StewType {
	return StewType{8}
}

// WitherStew returns suspicious stew wither effect.
func WitherStew() StewType {
	return StewType{9}
}

// NightVisionTorchflowerStew returns suspicious stew night vision effect.
func NightVisionTorchflowerStew() StewType {
	return StewType{10}
}

// BlindnessEyeblossomStew returns suspicious stew blindness effect.
func BlindnessEyeblossomStew() StewType {
	return StewType{11}
}

// NauseaStew returns suspicious stew nausea effect.
func NauseaStew() StewType {
	return StewType{12}
}

// StewTypes ...
func StewTypes() []StewType {
	return []StewType{NightVisionPoppyStew(), JumpBoostStew(), WeaknessStew(), BlindnessBluetStew(), PoisonStew(), SaturationDandelionStew(), SaturationOrchidStew(), FireResistanceStew(), RegenerationStew(), WitherStew(), NightVisionTorchflowerStew(), BlindnessEyeblossomStew(), NauseaStew()}
}

type stewType uint8

// Uint8 returns the stew as a uint8.
func (s stewType) Uint8() uint8 {
	return uint8(s)
}

// Effects returns suspicious stew effects.
func (s stewType) Effects() []effect.Effect {
	var effects []effect.Effect
	switch s.Uint8() {
	case 0, 10:
		effects = append(effects, effect.New(effect.NightVision, 1, time.Second*5))
	case 1:
		effects = append(effects, effect.New(effect.JumpBoost, 1, time.Second*5))
	case 2:
		effects = append(effects, effect.New(effect.Weakness, 1, time.Second*7))
	case 3, 11:
		effects = append(effects, effect.New(effect.Blindness, 1, time.Second*6))
	case 4:
		effects = append(effects, effect.New(effect.Poison, 1, time.Second*11))
	case 5, 6:
		effects = append(effects, effect.New(effect.Saturation, 1, time.Second*3/10))
	case 7:
		effects = append(effects, effect.New(effect.FireResistance, 1, time.Second*3))
	case 8:
		effects = append(effects, effect.New(effect.Regeneration, 1, time.Second*7))
	case 9:
		effects = append(effects, effect.New(effect.Wither, 1, time.Second*7))
	case 12:
		effects = append(effects, effect.New(effect.Nausea, 1, time.Second*7))
	default:
		panic("should never happen")
	}

	return effects
}
