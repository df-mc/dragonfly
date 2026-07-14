package entity

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type projectileShieldTarget struct {
	h       *world.EntityHandle
	pos     mgl64.Vec3
	blocked bool
}

func (*projectileShieldTarget) Close() error                              { return nil }
func (t *projectileShieldTarget) H() *world.EntityHandle                  { return t.h }
func (t *projectileShieldTarget) Position() mgl64.Vec3                    { return t.pos }
func (*projectileShieldTarget) Rotation() cube.Rotation                   { return cube.Rotation{} }
func (*projectileShieldTarget) Health() float64                           { return 20 }
func (*projectileShieldTarget) MaxHealth() float64                        { return 20 }
func (*projectileShieldTarget) SetMaxHealth(float64)                      {}
func (*projectileShieldTarget) Dead() bool                                { return false }
func (*projectileShieldTarget) Heal(float64, world.HealingSource) float64 { return 0 }
func (*projectileShieldTarget) KnockBack(mgl64.Vec3, float64, float64)    {}
func (*projectileShieldTarget) Velocity() mgl64.Vec3                      { return mgl64.Vec3{} }
func (*projectileShieldTarget) SetVelocity(mgl64.Vec3)                    {}
func (*projectileShieldTarget) AddEffect(effect.Effect)                   {}
func (*projectileShieldTarget) RemoveEffect(effect.Type)                  {}
func (*projectileShieldTarget) Effects() []effect.Effect                  { return nil }
func (*projectileShieldTarget) Speed() float64                            { return 0 }
func (*projectileShieldTarget) SetSpeed(float64)                          {}
func (t *projectileShieldTarget) Hurt(float64, world.DamageSource) (float64, world.HurtResult) {
	if t.blocked {
		return 0, world.HurtBlocked
	}
	return 0, world.HurtDamaged
}

type projectileShieldTargetConfig struct{ blocked bool }

func (c projectileShieldTargetConfig) Apply(data *world.EntityData) { data.Data = c.blocked }

type projectileShieldTargetType struct{}

func (projectileShieldTargetType) Open(_ *world.Tx, h *world.EntityHandle, data *world.EntityData) world.Entity {
	return &projectileShieldTarget{h: h, pos: data.Pos, blocked: data.Data.(bool)}
}
func (projectileShieldTargetType) EncodeEntity() string { return "dragonfly:shield_target" }
func (projectileShieldTargetType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.3, 0, -0.3, 0.3, 1.8, 0.3)
}
func (projectileShieldTargetType) DecodeNBT(map[string]any, *world.EntityData) {}
func (projectileShieldTargetType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type projectileTestParticle struct{ count *int }

func (p projectileTestParticle) Spawn(*world.World, mgl64.Vec3) { *p.count++ }

type projectileTestSound struct{ count *int }

func (s projectileTestSound) Play(*world.World, mgl64.Vec3) { *s.count++ }

func TestShieldDeflectsProjectile(t *testing.T) {
	for _, tt := range []struct {
		name   string
		damage float64
	}{
		{name: "zero damage", damage: 0},
		{name: "damaging", damage: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pos := mgl64.Vec3{0, 0, 1}
			velocity := mgl64.Vec3{0, 0, -1}
			behaviour := ProjectileBehaviourConfig{Damage: tt.damage}.New()
			projectile := &Ent{
				handle: world.EntitySpawnOpts{}.New(SnowballType, ProjectileBehaviourConfig{}),
				data:   &world.EntityData{Pos: pos, Data: behaviour},
			}
			if !behaviour.hitEntity(&projectileShieldTarget{blocked: true}, projectile, velocity) {
				t.Fatal("shield-blocked projectile was not deflected")
			}
			if got, want := projectile.Velocity(), velocity.Mul(-1); got != want {
				t.Fatalf("projectile velocity = %v, want %v", got, want)
			}
			if !projectile.Position().Sub(pos).Normalize().ApproxEqual(projectile.Velocity().Normalize()) {
				t.Fatal("projectile did not move away from blocker")
			}
		})
	}

	t.Run("suppresses hit effects", func(t *testing.T) {
		w := world.Config{Synchronous: true, Entities: world.EntityRegistryConfig{}.New([]world.EntityType{SnowballType, projectileShieldTargetType{}})}.New()
		defer w.Close()
		var particles, sounds int
		hit := false
		projectile := world.EntitySpawnOpts{Position: mgl64.Vec3{0, 0.5, 0}, Velocity: mgl64.Vec3{0, 0, 1}}.New(SnowballType, ProjectileBehaviourConfig{
			Damage: 0, Particle: projectileTestParticle{&particles}, ParticleCount: 1, Sound: projectileTestSound{&sounds},
			Hit: func(*Ent, *world.Tx, trace.Result) { hit = true },
		})
		target := world.EntitySpawnOpts{Position: mgl64.Vec3{0, 0, 0.8}}.New(projectileShieldTargetType{}, projectileShieldTargetConfig{blocked: true})
		w.Do(func(tx *world.Tx) {
			tx.AddEntity(target)
			tx.AddEntity(projectile).(*Ent).Tick(tx, 0)
		})
		if hit || particles != 0 || sounds != 0 {
			t.Fatalf("deflection emitted hit effects: callback=%v particles=%v sounds=%v", hit, particles, sounds)
		}
	})
}

type tntTestEntityType struct{}

func (tntTestEntityType) Open(*world.Tx, *world.EntityHandle, *world.EntityData) world.Entity {
	return nil
}
func (tntTestEntityType) EncodeEntity() string                        { return "dragonfly:test_entity" }
func (tntTestEntityType) BBox(world.Entity) cube.BBox                 { return cube.Box(0, 0, 0, 0, 0, 0) }
func (tntTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (tntTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
func (tntTestEntityType) Apply(*world.EntityData)                     {}

func TestShieldTNTConfig(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer w.Close()
	source := world.EntitySpawnOpts{}.New(tntTestEntityType{}, tntTestEntityType{})
	w.Do(func(tx *world.Tx) {
		tests := []struct {
			name      string
			source    *world.EntityHandle
			blockable bool
		}{
			{name: "unavailable source remains blockable", source: source, blockable: true},
			{name: "explicitly unblockable", blockable: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				conf := tntExplosionConfig(tx, tt.source, tt.blockable)
				if conf.UnblockableByShield == tt.blockable {
					t.Fatalf("UnblockableByShield = %v, want %v", conf.UnblockableByShield, !tt.blockable)
				}
				if conf.Source != nil {
					t.Fatal("unavailable source was attached to explosion config")
				}
			})
		}
	})
}

func TestShieldTNTNBT(t *testing.T) {
	var defaults world.EntityData
	TNTType.DecodeNBT(map[string]any{"Fuse": uint8(5)}, &defaults)
	if _, ok := TNTType.EncodeNBT(&defaults)["DragonflyUnblockableByShield"]; ok {
		t.Fatal("missing blockability tag did not default to shield-blockable")
	}

	var decoded world.EntityData
	TNTType.DecodeNBT(map[string]any{"Fuse": uint8(5), "DragonflyUnblockableByShield": uint8(1)}, &decoded)
	if got := TNTType.EncodeNBT(&decoded)["DragonflyUnblockableByShield"]; got != uint8(1) {
		t.Fatalf("unblockable setting after NBT round trip = %#v, want 1", got)
	}

	for _, tt := range []struct {
		name string
		fuse time.Duration
		want uint8
	}{
		{name: "negative fuse", fuse: -time.Second, want: 0},
		{name: "oversized fuse", fuse: 20 * time.Second, want: 255},
	} {
		t.Run(tt.name, func(t *testing.T) {
			data := world.EntityData{Data: tntBehaviourConfig{Fuse: tt.fuse, BlockableByShield: true}.New()}
			if got := TNTType.EncodeNBT(&data)["Fuse"]; got != tt.want {
				t.Fatalf("encoded fuse = %#v, want %v", got, tt.want)
			}
		})
	}
}
