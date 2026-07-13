package entity

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestWindChargeKnockback(t *testing.T) {
	tests := []struct {
		name     string
		burst    mgl64.Vec3
		position mgl64.Vec3
		eye      mgl64.Vec3
		velocity mgl64.Vec3
		exposure float64
		want     mgl64.Vec3
	}{
		{
			name:     "uses eye direction and adds to existing velocity",
			burst:    mgl64.Vec3{},
			position: mgl64.Vec3{1.2, 0, 0},
			eye:      mgl64.Vec3{1.2, 1.6, 0},
			velocity: mgl64.Vec3{0.1, 0.2, -0.3},
			exposure: 0.5,
			want: mgl64.Vec3{0.1, 0.2, -0.3}.Add(
				mgl64.Vec3{1.2, 1.6, 0}.Normalize().Mul(0.305),
			),
		},
		{
			name:     "directly above receives vertical impulse",
			burst:    mgl64.Vec3{},
			position: mgl64.Vec3{0, 1.2, 0},
			eye:      mgl64.Vec3{0, 2.82, 0},
			exposure: 1,
			want:     mgl64.Vec3{0, 0.61, 0},
		},
		{
			name:     "outside diameter is unchanged",
			burst:    mgl64.Vec3{},
			position: mgl64.Vec3{2.5, 0, 0},
			eye:      mgl64.Vec3{2.5, 1.62, 0},
			velocity: mgl64.Vec3{0.1, 0.2, -0.3},
			exposure: 1,
			want:     mgl64.Vec3{0.1, 0.2, -0.3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := windChargeKnockback(test.burst, test.position, test.eye, test.velocity, test.exposure)
			if !got.ApproxEqualThreshold(test.want, 1e-12) {
				t.Fatalf("knockback velocity = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWindChargeBlockExplosionPosition(t *testing.T) {
	hit := mgl64.Vec3{1, 2, 3}
	tests := []struct {
		face cube.Face
		want mgl64.Vec3
	}{
		{face: cube.FaceDown, want: mgl64.Vec3{1, 1.75, 3}},
		{face: cube.FaceUp, want: mgl64.Vec3{1, 2.25, 3}},
		{face: cube.FaceNorth, want: mgl64.Vec3{1, 2, 2.75}},
		{face: cube.FaceSouth, want: mgl64.Vec3{1, 2, 3.25}},
		{face: cube.FaceWest, want: mgl64.Vec3{0.75, 2, 3}},
		{face: cube.FaceEast, want: mgl64.Vec3{1.25, 2, 3}},
	}
	for _, test := range tests {
		t.Run(test.face.String(), func(t *testing.T) {
			if got := windChargeBlockExplosionPosition(hit, test.face); got != test.want {
				t.Fatalf("explosion position = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWindChargeCanHitType(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "minecraft:wind_charge_projectile", want: false},
		{name: "minecraft:end_crystal", want: false},
		{name: "minecraft:ender_crystal", want: false},
		{name: "minecraft:player", want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := windChargeCanHitType(test.name); got != test.want {
				t.Fatalf("windChargeCanHitType(%q) = %t, want %t", test.name, got, test.want)
			}
		})
	}
}

func TestWindChargeBoundingBoxMatchesJavaOffset(t *testing.T) {
	want := cube.Box(-0.15625, -0.15, -0.15625, 0.15625, 0.1625, 0.15625)
	if got := WindChargeType.BBox(nil); got != want {
		t.Fatalf("bounding box = %v, want %v", got, want)
	}
}

func TestWindChargeClosesAfterHittingNonLivingEntity(t *testing.T) {
	targetType := windChargeTestEntityType{}
	registry := world.EntityRegistryConfig{}.New([]world.EntityType{WindChargeType, targetType})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer w.Close()

	ownerHandle := world.EntitySpawnOpts{Position: mgl64.Vec3{-10, 1, 0}}.New(targetType, PassiveBehaviourConfig{})
	targetHandle := world.EntitySpawnOpts{Position: mgl64.Vec3{1, 1, 0}}.New(targetType, PassiveBehaviourConfig{})
	var chargeHandle *world.EntityHandle
	w.Do(func(tx *world.Tx) {
		owner := tx.AddEntity(ownerHandle)
		tx.AddEntity(targetHandle)
		chargeHandle = NewWindCharge(world.EntitySpawnOpts{
			Position: mgl64.Vec3{0, 1, 0},
			Velocity: mgl64.Vec3{2, 0, 0},
		}, owner)
		tx.AddEntity(chargeHandle)
	})

	w.AdvanceTick()
	w.AdvanceTick()
	if !chargeHandle.Closed() {
		t.Fatal("wind charge remained open after hitting a non-living entity")
	}
}

type windChargeTestEntityType struct{}

func (windChargeTestEntityType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}
func (windChargeTestEntityType) EncodeEntity() string { return "test:wind_charge_target" }
func (windChargeTestEntityType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.25, 0, -0.25, 0.25, 0.5, 0.25)
}
func (windChargeTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (windChargeTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
