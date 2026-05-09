package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type projectileShieldTarget struct {
	pos        mgl64.Vec3
	h          *world.EntityHandle
	blocked    bool
	vulnerable bool
}

func (t *projectileShieldTarget) Close() error                           { return nil }
func (t *projectileShieldTarget) H() *world.EntityHandle                 { return t.h }
func (t *projectileShieldTarget) Position() mgl64.Vec3                   { return t.pos }
func (t *projectileShieldTarget) Rotation() cube.Rotation                { return cube.Rotation{} }
func (t *projectileShieldTarget) Health() float64                        { return 20 }
func (t *projectileShieldTarget) MaxHealth() float64                     { return 20 }
func (t *projectileShieldTarget) SetMaxHealth(float64)                   {}
func (t *projectileShieldTarget) Dead() bool                             { return false }
func (t *projectileShieldTarget) Heal(float64, world.HealingSource)      {}
func (t *projectileShieldTarget) KnockBack(mgl64.Vec3, float64, float64) {}
func (t *projectileShieldTarget) Velocity() mgl64.Vec3                   { return mgl64.Vec3{} }
func (t *projectileShieldTarget) SetVelocity(mgl64.Vec3)                 {}
func (t *projectileShieldTarget) AddEffect(effect.Effect)                {}
func (t *projectileShieldTarget) RemoveEffect(effect.Type)               {}
func (t *projectileShieldTarget) Effects() []effect.Effect               { return nil }
func (t *projectileShieldTarget) Speed() float64                         { return 0 }
func (t *projectileShieldTarget) SetSpeed(float64)                       {}

func (t *projectileShieldTarget) Hurt(_ float64, src world.DamageSource) (float64, bool) {
	if s, ok := src.(ProjectileDamageSource); ok && t.blocked {
		s.Projectile.(*Ent).Behaviour().(interface{ MarkShieldBlocked() }).MarkShieldBlocked()
	}
	return 0, t.vulnerable
}

type projectileShieldTargetConfig struct {
	blocked    bool
	vulnerable bool
}

func (c projectileShieldTargetConfig) Apply(data *world.EntityData) {
	data.Data = c
}

type projectileShieldTargetType struct{}

func (projectileShieldTargetType) Open(_ *world.Tx, h *world.EntityHandle, data *world.EntityData) world.Entity {
	conf := data.Data.(projectileShieldTargetConfig)
	return &projectileShieldTarget{pos: data.Pos, h: h, blocked: conf.blocked, vulnerable: conf.vulnerable}
}
func (projectileShieldTargetType) EncodeEntity() string { return "dragonfly:shield_target" }
func (projectileShieldTargetType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3)
}
func (projectileShieldTargetType) DecodeNBT(map[string]any, *world.EntityData) {}
func (projectileShieldTargetType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type projectileTestParticle struct {
	count *int
}

func (p projectileTestParticle) Spawn(*world.World, mgl64.Vec3) {
	(*p.count)++
}

type projectileTestSound struct {
	count *int
}

func (s projectileTestSound) Play(*world.World, mgl64.Vec3) {
	(*s.count)++
}

func newProjectileShieldTestEnt(pos mgl64.Vec3, behaviour *ProjectileBehaviour) *Ent {
	return &Ent{
		handle: world.EntitySpawnOpts{}.New(SnowballType, ProjectileBehaviourConfig{}),
		data:   &world.EntityData{Pos: pos, Data: behaviour},
	}
}

func TestProjectileDeflectsAfterShieldBlock(t *testing.T) {
	pos := mgl64.Vec3{0, 0, 1}
	behaviour := ProjectileBehaviourConfig{Damage: 2}.New()
	projectile := newProjectileShieldTestEnt(pos, behaviour)
	velocity := mgl64.Vec3{0, 0, -1}

	blocked := behaviour.hitEntity(&projectileShieldTarget{blocked: true}, projectile, velocity)
	if !blocked {
		t.Fatal("expected shield-blocked projectile hit to be handled as a deflection")
	}
	if got, want := projectile.Velocity(), velocity.Mul(-1); got != want {
		t.Fatalf("expected deflected projectile velocity %v, got %v", want, got)
	}
	if !projectile.Position().Sub(pos).Normalize().ApproxEqual(projectile.Velocity().Normalize()) {
		t.Fatalf("expected projectile to move away from blocker after deflection, position changed from %v to %v with velocity %v", pos, projectile.Position(), projectile.Velocity())
	}
}

func TestProjectileDeflectsZeroDamageShieldBlock(t *testing.T) {
	pos := mgl64.Vec3{0, 0, 1}
	behaviour := ProjectileBehaviourConfig{Damage: 0}.New()
	projectile := newProjectileShieldTestEnt(pos, behaviour)
	velocity := mgl64.Vec3{0, 0, -1}

	blocked := behaviour.hitEntity(&projectileShieldTarget{blocked: true}, projectile, velocity)
	if !blocked {
		t.Fatal("expected zero damage shield-blocked projectile hit to be handled as a deflection")
	}
	if got, want := projectile.Velocity(), velocity.Mul(-1); got != want {
		t.Fatalf("expected deflected projectile velocity %v, got %v", want, got)
	}
}

func TestProjectileDeflectionSkipsHitCallback(t *testing.T) {
	w := world.Config{Entities: world.EntityRegistryConfig{}.New([]world.EntityType{SnowballType, projectileShieldTargetType{}})}.New()
	defer func() {
		_ = w.Close()
	}()
	var particles, sounds int
	hit := false
	projectile := world.EntitySpawnOpts{
		Position: mgl64.Vec3{0, 0.5, 0},
		Velocity: mgl64.Vec3{0, 0, 1},
	}.New(SnowballType, ProjectileBehaviourConfig{
		Damage:        0,
		Particle:      projectileTestParticle{count: &particles},
		ParticleCount: 1,
		Sound:         projectileTestSound{count: &sounds},
		Hit: func(*Ent, *world.Tx, trace.Result) {
			hit = true
		},
	})
	target := world.EntitySpawnOpts{Position: mgl64.Vec3{0, 0, 0.8}}.New(projectileShieldTargetType{}, projectileShieldTargetConfig{blocked: true})

	<-w.Exec(func(tx *world.Tx) {
		tx.AddEntity(target)
		tx.AddEntity(projectile).(*Ent).Tick(tx, 0)
	})
	if hit {
		t.Fatal("expected shield-deflected projectile not to run hit callback")
	}
	if particles != 0 {
		t.Fatalf("expected shield-deflected projectile not to spawn hit particles, got %v", particles)
	}
	if sounds != 0 {
		t.Fatalf("expected shield-deflected projectile not to play hit sound, got %v", sounds)
	}
}
