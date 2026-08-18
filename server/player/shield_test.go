package player

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestShieldShouldBlockDamage(t *testing.T) {
	tests := []struct {
		name                             string
		raw, beforeHandler, afterHandler float64
		info                             world.ShieldBlockInfo
		want                             bool
	}{
		{name: "positive damage", raw: 4, beforeHandler: 4, afterHandler: 4, want: true},
		{name: "mitigated before handler", raw: 4, want: true},
		{name: "reduced to zero by handler", raw: 4, beforeHandler: 4},
		{name: "zero damage projectile", info: world.ShieldBlockInfo{BlockZeroDamage: true}, want: true},
		{name: "unmarked zero damage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shieldShouldBlockDamage(tt.raw, tt.beforeHandler, tt.afterHandler, tt.info); got != tt.want {
				t.Fatalf("shieldShouldBlockDamage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpecialAttackerDisablesShield(t *testing.T) {
	attacker := specialShieldDisabler{}
	got, ok := shieldDisableCooldownFrom(entity.AttackDamageSource{Attacker: attacker})
	if !ok || got != shieldDisableCooldown {
		t.Fatalf("shield disable cooldown = (%v, %v), want (%v, true)", got, ok, shieldDisableCooldown)
	}
}

type specialShieldDisabler struct{}

func (specialShieldDisabler) CanDisableShield() bool  { return true }
func (specialShieldDisabler) Close() error            { return nil }
func (specialShieldDisabler) H() *world.EntityHandle  { return nil }
func (specialShieldDisabler) Position() mgl64.Vec3    { return mgl64.Vec3{} }
func (specialShieldDisabler) Rotation() cube.Rotation { return cube.Rotation{} }
