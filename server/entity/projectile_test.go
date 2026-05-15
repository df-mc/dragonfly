package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestProjectileClosesAfterNonSurvivingBlockCollision(t *testing.T) {
	w := world.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	var closed bool
	<-w.Exec(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{1, 0, 0}, block.Stone{}, nil)

		conf := ProjectileBehaviourConfig{
			Drag: 0,
		}
		handle := world.EntitySpawnOpts{
			Position: mgl64.Vec3{0, 0.5, 0.5},
			Velocity: mgl64.Vec3{2, 0, 0},
		}.New(SnowballType, conf)
		projectile := tx.AddEntity(handle).(*Ent)

		projectile.Tick(tx, 1)

		behaviour := projectile.Behaviour().(*ProjectileBehaviour)
		closed = behaviour.close
	})
	if !closed {
		t.Fatal("expected non-surviving projectile to close after block collision")
	}
}

func TestProjectileClosesAfterNoDamageEntityCollision(t *testing.T) {
	w := world.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	var closed bool
	<-w.Exec(func(tx *world.Tx) {
		target := world.EntitySpawnOpts{Position: mgl64.Vec3{1, 0.25, 0}}.New(testLivingType{}, PassiveBehaviourConfig{})
		tx.AddEntity(target)

		conf := ProjectileBehaviourConfig{
			Damage: -1,
			Drag:   0,
		}
		handle := world.EntitySpawnOpts{
			Position: mgl64.Vec3{0, 0.25, 0},
			Velocity: mgl64.Vec3{2, 0, 0},
		}.New(SnowballType, conf)
		projectile := tx.AddEntity(handle).(*Ent)

		projectile.Tick(tx, 1)

		behaviour := projectile.Behaviour().(*ProjectileBehaviour)
		closed = behaviour.close
	})
	if !closed {
		t.Fatal("expected non-damaging projectile to close after entity collision")
	}
}

type testLivingType struct{}

func (testLivingType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &testLiving{tx: tx, handle: handle, data: data}
}

func (testLivingType) EncodeEntity() string { return "dragonfly:test_living" }
func (testLivingType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.25, -0.25, -0.25, 0.25, 0.25, 0.25)
}
func (testLivingType) DecodeNBT(map[string]any, *world.EntityData) {}
func (testLivingType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type testLiving struct {
	tx     *world.Tx
	handle *world.EntityHandle
	data   *world.EntityData
}

func (l *testLiving) H() *world.EntityHandle  { return l.handle }
func (l *testLiving) Position() mgl64.Vec3    { return l.data.Pos }
func (l *testLiving) Rotation() cube.Rotation { return l.data.Rot }
func (l *testLiving) Health() float64         { return 20 }
func (l *testLiving) MaxHealth() float64      { return 20 }
func (l *testLiving) SetMaxHealth(float64)    {}
func (l *testLiving) Dead() bool              { return false }
func (l *testLiving) Close() error {
	l.tx.RemoveEntity(l)
	return l.handle.Close()
}
func (l *testLiving) Hurt(float64, world.DamageSource) (float64, bool) {
	return 0, true
}
func (l *testLiving) Heal(float64, world.HealingSource)      {}
func (l *testLiving) KnockBack(mgl64.Vec3, float64, float64) {}
func (l *testLiving) Velocity() mgl64.Vec3                   { return l.data.Vel }
func (l *testLiving) SetVelocity(v mgl64.Vec3)               { l.data.Vel = v }
func (l *testLiving) AddEffect(effect.Effect)                {}
func (l *testLiving) RemoveEffect(effect.Type)               {}
func (l *testLiving) Effects() []effect.Effect               { return nil }
func (l *testLiving) Speed() float64                         { return 0 }
func (l *testLiving) SetSpeed(float64)                       {}
