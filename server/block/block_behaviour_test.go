package block_test

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TestTorchBreaksWithoutSupport verifies that a torch is broken by a neighbour
// update on the tick after its supporting block is removed, using a
// synchronous World to make the tick deterministic.
func TestTorchBreaksWithoutSupport(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	support, torch := cube.Pos{0, 0, 0}, cube.Pos{0, 1, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(support, block.Stone{}, nil)
		tx.SetBlock(torch, block.Torch{Facing: cube.FaceDown}, nil)
		tx.SetBlock(support, block.Air{}, nil)
	})
	w.AdvanceTick()

	b, err := world.Call(context.Background(), w, func(tx *world.Tx) (world.Block, error) {
		return tx.Block(torch), nil
	})
	if err != nil {
		t.Fatalf("read torch block: %v", err)
	}
	if b != (block.Air{}) {
		t.Errorf("expected torch to break after removing its support, got %v", b)
	}
}

func TestSweetBerriesConsume(t *testing.T) {
	consumer := &sweetBerryConsumer{}
	block.SweetBerries{}.Consume(nil, consumer)

	if consumer.food != 2 || consumer.saturation != 1.2 {
		t.Fatalf("Saturate called with (%v, %v), want (2, 1.2)", consumer.food, consumer.saturation)
	}
}

func TestSweetBerryBushStageZeroSlowsEntities(t *testing.T) {
	entity := &sweetBerryTestEntity{velocity: mgl64.Vec3{1, 1, 1}}

	block.SweetBerries{Growth: 0}.EntityInside(cube.Pos{}, nil, entity)

	if got, want := entity.velocity, (mgl64.Vec3{0.8, 0.75, 0.8}); got != want {
		t.Fatalf("velocity = %v, want %v", got, want)
	}
	if !entity.fallDistanceReset {
		t.Fatal("fall distance was not reset")
	}
}

func TestSweetBerryBushDoesNotSlowItemEntities(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	w.Do(func(tx *world.Tx) {
		handle := entity.NewItem(world.EntitySpawnOpts{Velocity: mgl64.Vec3{1, 1, 1}}, item.NewStack(block.SweetBerries{}, 1))
		tx.AddEntity(handle)
		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("item entity was not added to the world")
		}

		block.SweetBerries{Growth: 3}.EntityInside(cube.Pos{}, tx, e)

		if got, want := e.(interface{ Velocity() mgl64.Vec3 }).Velocity(), (mgl64.Vec3{1, 1, 1}); got != want {
			t.Fatalf("velocity = %v, want %v", got, want)
		}
	}).Wait(context.Background())
}

func TestSweetBerryBushDoesNotGrowUnderOpaqueFullBlock(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	pos := cube.Pos{0, 64, 0}
	w.Do(func(tx *world.Tx) {
		bush := block.SweetBerries{Growth: 0}
		tx.SetBlock(pos, bush, nil)
		tx.SetBlock(pos.Side(cube.FaceUp), block.Glowstone{}, nil)

		bush.RandomTick(pos, tx, rand.New(sweetBerryGrowthSource{}))

		if got := tx.Block(pos); got != bush {
			t.Fatalf("bush = %v, want %v", got, bush)
		}
	}).Wait(context.Background())
}

type sweetBerryConsumer struct {
	sweetBerryTestEntity
	food       int
	saturation float64
}

func (c *sweetBerryConsumer) Saturate(food int, saturation float64) {
	c.food, c.saturation = food, saturation
}

func (*sweetBerryConsumer) AddEffect(effect.Effect)  {}
func (*sweetBerryConsumer) RemoveEffect(effect.Type) {}
func (*sweetBerryConsumer) Effects() []effect.Effect { return nil }
func (*sweetBerryConsumer) Absorption() float64      { return 0 }
func (*sweetBerryConsumer) SetAbsorption(float64)    {}
func (*sweetBerryConsumer) HeldItems() (item.Stack, item.Stack) {
	return item.Stack{}, item.Stack{}
}

func (*sweetBerryConsumer) SetHeldItems(item.Stack, item.Stack) {}
func (*sweetBerryConsumer) UsingItem() bool                     { return false }
func (*sweetBerryConsumer) ReleaseItem()                        {}
func (*sweetBerryConsumer) UseItem()                            {}

type sweetBerryTestEntity struct {
	velocity          mgl64.Vec3
	fallDistanceReset bool
}

func (*sweetBerryTestEntity) Close() error               { return nil }
func (*sweetBerryTestEntity) H() *world.EntityHandle     { return nil }
func (*sweetBerryTestEntity) Position() mgl64.Vec3       { return mgl64.Vec3{} }
func (*sweetBerryTestEntity) Rotation() cube.Rotation    { return cube.Rotation{} }
func (e *sweetBerryTestEntity) Velocity() mgl64.Vec3     { return e.velocity }
func (e *sweetBerryTestEntity) SetVelocity(v mgl64.Vec3) { e.velocity = v }
func (e *sweetBerryTestEntity) ResetFallDistance()       { e.fallDistanceReset = true }
func (*sweetBerryTestEntity) FallDistance() float64      { return 0 }

type sweetBerryGrowthSource struct{}

func (sweetBerryGrowthSource) Uint64() uint64 { return 1 }
