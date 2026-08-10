package effect

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TestPoisonInterval verifies how often poison deals damage. Vanilla halves the interval with every level, starting
// at 25 ticks, where regeneration starts at 50. Both are the same expression with a different constant.
func TestPoisonInterval(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{level: 1, want: 25},
		{level: 2, want: 12},
		{level: 3, want: 6},
	}
	for _, tt := range tests {
		e := &effectTestEntity{}
		eff := New(Poison, tt.level, time.Minute)
		var first, second int
		for tick := 1; tick <= 200; tick++ {
			before := e.hurt
			eff = eff.TickDuration()
			Poison.Apply(e, eff)
			if e.hurt > before {
				if first == 0 {
					first = tick
				} else if second == 0 {
					second = tick
					break
				}
			}
		}
		if got := second - first; got != tt.want {
			t.Errorf("poison %v deals damage every %v ticks, want every %v", tt.level, got, tt.want)
		}
	}
}

// effectTestEntity is a minimal living entity that counts the times it was hurt.
type effectTestEntity struct {
	world.Entity
	hurt int
}

func (e *effectTestEntity) Health() float64      { return 20 }
func (e *effectTestEntity) MaxHealth() float64   { return 20 }
func (e *effectTestEntity) SetMaxHealth(float64) {}
func (e *effectTestEntity) Hurt(float64, world.DamageSource) (float64, bool) {
	e.hurt++
	return 1, true
}
func (e *effectTestEntity) Heal(float64, world.HealingSource) float64 { return 0 }
func (e *effectTestEntity) Speed() float64                            { return 0 }
func (e *effectTestEntity) SetSpeed(float64)                          {}
func (e *effectTestEntity) Position() mgl64.Vec3                      { return mgl64.Vec3{} }
func (e *effectTestEntity) Rotation() cube.Rotation                   { return cube.Rotation{} }

// TestFatalPoisonInterval verifies that fatal poison deals damage as often as poison does. The two differ only in
// whether the damage can kill.
func TestFatalPoisonInterval(t *testing.T) {
	e, f := &effectTestEntity{}, &effectTestEntity{}
	poison, fatal := New(Poison, 1, time.Minute), New(FatalPoison, 1, time.Minute)
	for range 200 {
		poison, fatal = poison.TickDuration(), fatal.TickDuration()
		Poison.Apply(e, poison)
		FatalPoison.Apply(f, fatal)
	}
	if e.hurt != f.hurt {
		t.Errorf("fatal poison dealt damage %v times where poison dealt it %v times", f.hurt, e.hurt)
	}
}
