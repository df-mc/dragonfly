package item_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
	_ "github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestCrossbowArrowVelocityMatchesBedrock(t *testing.T) {
	var velocity mgl64.Vec3
	registry := world.EntityRegistryConfig{
		Arrow: func(opts world.EntitySpawnOpts, _ world.ArrowSpawnConfig) *world.EntityHandle {
			velocity = opts.Velocity
			return opts.New(crossbowTestEntityType{}, crossbowTestEntityConfig{})
		},
	}.New([]world.EntityType{crossbowTestEntityType{}})
	w := world.Config{Synchronous: true, Entities: registry}.New()
	defer w.Close()

	crossbow := item.Crossbow{Item: item.NewStack(item.Arrow{}, 1)}
	releaser := crossbowTestReleaser{
		rotation: cube.Rotation{90, 0},
		held:     item.NewStack(crossbow, 1),
	}
	w.Do(func(tx *world.Tx) {
		if !crossbow.ReleaseCharge(&releaser, tx, &item.UseContext{}) {
			t.Fatal("expected charged crossbow to fire")
		}
	})

	want := releaser.Rotation().Vec3().Mul(5)
	if !velocity.ApproxEqual(want) {
		t.Fatalf("expected Bedrock crossbow arrow velocity %v, got %v", want, velocity)
	}
}

type crossbowTestReleaser struct {
	rotation cube.Rotation
	held     item.Stack
}

func (*crossbowTestReleaser) Close() error                          { return nil }
func (*crossbowTestReleaser) H() *world.EntityHandle                { return nil }
func (*crossbowTestReleaser) Position() mgl64.Vec3                  { return mgl64.Vec3{} }
func (r *crossbowTestReleaser) Rotation() cube.Rotation             { return r.rotation }
func (r *crossbowTestReleaser) HeldItems() (item.Stack, item.Stack) { return r.held, item.Stack{} }
func (r *crossbowTestReleaser) SetHeldItems(main, _ item.Stack)     { r.held = main }
func (*crossbowTestReleaser) UsingItem() bool                       { return false }
func (*crossbowTestReleaser) ReleaseItem()                          {}
func (*crossbowTestReleaser) UseItem()                              {}
func (*crossbowTestReleaser) GameMode() world.GameMode              { return world.GameModeSurvival }
func (*crossbowTestReleaser) PlaySound(world.Sound)                 {}

type crossbowTestEntityConfig struct{}

func (crossbowTestEntityConfig) Apply(*world.EntityData) {}

type crossbowTestEntityType struct{}

func (crossbowTestEntityType) Open(_ *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return crossbowTestEntity{handle: handle, data: data}
}
func (crossbowTestEntityType) EncodeEntity() string { return "test:crossbow_arrow" }
func (crossbowTestEntityType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, -0.125, -0.125, 0.125, 0.125, 0.125)
}
func (crossbowTestEntityType) DecodeNBT(map[string]any, *world.EntityData) {}
func (crossbowTestEntityType) EncodeNBT(*world.EntityData) map[string]any  { return nil }

type crossbowTestEntity struct {
	handle *world.EntityHandle
	data   *world.EntityData
}

func (crossbowTestEntity) Close() error              { return nil }
func (e crossbowTestEntity) H() *world.EntityHandle  { return e.handle }
func (e crossbowTestEntity) Position() mgl64.Vec3    { return e.data.Pos }
func (e crossbowTestEntity) Rotation() cube.Rotation { return e.data.Rot }
