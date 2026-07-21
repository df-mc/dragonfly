package block_test

import (
	"context"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestHopperInitialisesRedstoneLockOnPlacement(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(pos.Side(cube.FaceEast), block.RedstoneBlock{}, nil)
		ctx := &item.UseContext{}
		if !block.NewHopper().UseOnBlock(pos, cube.FaceUp, mgl64.Vec3{}, tx, placementUser{tx: tx}, ctx) {
			t.Fatal("expected hopper placement to succeed")
		}
	}).Wait(context.Background())
	w.AdvanceTick()
	w.Do(func(tx *world.Tx) {
		if !tx.Block(pos).(block.Hopper).Powered {
			t.Fatal("expected hopper placed beside existing power to be locked")
		}
	}).Wait(context.Background())
}

func TestDispenserInitialisesRedstoneTriggerOnPlacement(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	w.Do(func(tx *world.Tx) {
		tx.SetBlock(cube.Pos{1, 1, 0}, block.RedstoneBlock{}, nil)
		ctx := &item.UseContext{}
		if !block.NewDispenser().UseOnBlock(pos, cube.FaceUp, mgl64.Vec3{}, tx, placementUser{tx: tx}, ctx) {
			t.Fatal("expected dispenser placement to succeed")
		}
		d := tx.Block(pos).(block.Dispenser)
		if err := d.Inventory(tx, pos).SetItem(0, item.NewStack(item.Stick{}, 1)); err != nil {
			t.Fatalf("put item in placed dispenser: %v", err)
		}
	}).Wait(context.Background())
	w.AdvanceTick()
	w.Do(func(tx *world.Tx) {
		d := tx.Block(pos).(block.Dispenser)
		if !d.Triggered {
			t.Fatal("expected dispenser placed beside existing quasi-connectivity power to be triggered")
		}
	}).Wait(context.Background())

	for range 4 {
		w.AdvanceTick()
	}
	if got := entityCount(t, w); got != 1 {
		t.Fatalf("expected placed dispenser to fire after four ticks, got %d entities", got)
	}
}

type placementUser struct {
	tx *world.Tx
}

func (u placementUser) PlaceBlock(pos cube.Pos, b world.Block, ctx *item.UseContext) {
	u.tx.SetBlock(pos, b, nil)
	ctx.SubtractFromCount(1)
}

func (placementUser) Close() error                        { return nil }
func (placementUser) H() *world.EntityHandle              { return nil }
func (placementUser) Position() mgl64.Vec3                { return mgl64.Vec3{0, 0, 3} }
func (placementUser) Rotation() cube.Rotation             { return cube.Rotation{} }
func (placementUser) HeldItems() (item.Stack, item.Stack) { return item.Stack{}, item.Stack{} }
func (placementUser) SetHeldItems(item.Stack, item.Stack) {}
func (placementUser) UsingItem() bool                     { return false }
func (placementUser) ReleaseItem()                        {}
func (placementUser) UseItem()                            {}
