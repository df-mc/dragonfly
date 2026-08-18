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

func TestShieldBlockVisualState(t *testing.T) {
	p := &Player{playerData: &playerData{}}
	w := new(world.World)
	p.recordShieldBlock(0, w, 10)
	if blocked, damaged := p.ShieldBlockState(); !blocked || damaged {
		t.Fatalf("non-damaging block state = (%v, %v), want (true, false)", blocked, damaged)
	}
	p.recordShieldBlock(4, w, 10)
	if blocked, damaged := p.ShieldBlockState(); blocked || !damaged {
		t.Fatalf("damaging block state = (%v, %v), want (false, true)", blocked, damaged)
	}
	if p.clearShieldBlockState(w, 10) {
		t.Fatal("clearShieldBlockState() cleared state during its originating tick")
	}
	if !p.clearShieldBlockState(w, 11) {
		t.Fatal("clearShieldBlockState() did not report a transient state")
	}
	if blocked, damaged := p.ShieldBlockState(); blocked || damaged {
		t.Fatalf("cleared block state = (%v, %v), want (false, false)", blocked, damaged)
	}
}

func TestShieldBlockVisualStateClearsAcrossWorlds(t *testing.T) {
	p := &Player{playerData: &playerData{}}
	p.recordShieldBlock(4, new(world.World), 100)
	if !p.clearShieldBlockState(new(world.World), 1) {
		t.Fatal("clearShieldBlockState() retained state from a different world tick domain")
	}
}

func TestDeadPlayerClearsShieldBlockVisualState(t *testing.T) {
	p := &Player{
		data: &world.EntityData{},
		playerData: &playerData{
			health: entity.NewHealthManager(0, 20),
		},
	}
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	w.Do(func(tx *world.Tx) {
		p.tx = tx
		p.recordShieldBlock(4, tx.World(), 10)
		p.Tick(tx, 11)
	})
	if blocked, damaged := p.ShieldBlockState(); blocked || damaged {
		t.Fatalf("dead player retained shield block state (%v, %v)", blocked, damaged)
	}
}

type specialShieldDisabler struct{}

func (specialShieldDisabler) CanDisableShield() bool  { return true }
func (specialShieldDisabler) Close() error            { return nil }
func (specialShieldDisabler) H() *world.EntityHandle  { return nil }
func (specialShieldDisabler) Position() mgl64.Vec3    { return mgl64.Vec3{} }
func (specialShieldDisabler) Rotation() cube.Rotation { return cube.Rotation{} }
