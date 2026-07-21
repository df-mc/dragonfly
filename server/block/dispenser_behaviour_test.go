package block_test

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

func runDispenserWorld(t *testing.T, w *world.World, f func(*world.Tx)) {
	t.Helper()
	if err := w.Do(f).Wait(context.Background()); err != nil {
		t.Fatalf("run world task: %v", err)
	}
}

func TestEmptyDispenserPlaysFailureClick(t *testing.T) {
	w := world.Config{Synchronous: true}.New()
	defer func() { _ = w.Close() }()

	viewer := &soundViewer{}
	loader := world.NewLoader(1, w, viewer)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		loader.Load(tx, 1)
		d := block.NewDispenser()
		d.ScheduledTick(cube.Pos{}, tx, rand.New(rand.NewPCG(1, 2)))
		loader.Close(tx)
	})

	if len(viewer.sounds) != 1 {
		t.Fatalf("expected one dispenser failure sound, got %d", len(viewer.sounds))
	}
	if _, ok := viewer.sounds[0].(sound.ClickFail); !ok {
		t.Fatalf("expected dispenser failure click, got %T", viewer.sounds[0])
	}
}

func TestDispenserDoesNotReplayTargetIgnitionSound(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	viewer := &soundViewer{}
	loader := world.NewLoader(1, w, viewer)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		loader.Load(tx, 1)
		pos := cube.Pos{}
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.FlintAndSteel{}, 1))
		tx.SetBlock(pos.Side(cube.FaceEast), block.Campfire{Extinguished: true}, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(3, 4)))
		loader.Close(tx)
	})

	ignitions := 0
	for _, played := range viewer.sounds {
		if _, ok := played.(sound.Ignite); ok {
			ignitions++
		}
	}
	if ignitions != 1 {
		t.Fatalf("expected target block to own one ignition sound, got %d", ignitions)
	}
}

func TestDispenserRetainsFlintAndSteelWhenFireAlreadyExists(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{}
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		stack := item.NewStack(item.FlintAndSteel{}, 1)
		_ = d.Inventory(tx, pos).SetItem(0, stack)
		tx.SetBlock(pos.Side(cube.FaceEast), block.Fire{}, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(5, 6)))

		got, _ := d.Inventory(tx, pos).Item(0)
		if got.Durability() != stack.Durability() {
			t.Fatalf("expected existing fire not to damage flint and steel: durability %d, want %d", got.Durability(), stack.Durability())
		}
	})
}

type soundViewer struct {
	world.NopViewer
	sounds []world.Sound
}

func (v *soundViewer) ViewSound(_ mgl64.Vec3, played world.Sound) {
	v.sounds = append(v.sounds, played)
}

func TestDispenserDispensesAfterFourTickPulse(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	powerPos := cube.Pos{1, 1, 0}
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceNorth
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.Stick{}, 1))
		tx.SetBlock(pos, d, nil)
		tx.SetBlock(powerPos, block.RedstoneBlock{}, nil)
	})
	w.AdvanceTick()
	runDispenserWorld(t, w, func(tx *world.Tx) {
		tx.SetBlock(powerPos, block.Air{}, nil)
	})
	w.AdvanceTick()

	for range 2 {
		w.AdvanceTick()
	}
	if got := entityCount(t, w); got != 0 {
		t.Fatalf("dispenser fired before its four-tick delay: got %d entities", got)
	}

	w.AdvanceTick()
	if got := entityCount(t, w); got != 1 {
		t.Fatalf("expected one dispensed item entity after four ticks, got %d", got)
	}
	runDispenserWorld(t, w, func(tx *world.Tx) {
		if stack, _ := tx.Block(pos).(block.Dispenser).Inventory(tx, pos).Item(0); !stack.Empty() {
			t.Fatalf("expected selected dispenser slot to be consumed, got %v", stack)
		}
	})
}

// TestDispenserLaunchesOwnerlessProjectiles covers the projectiles a dispenser launches without an owner entity. Every
// one of these reaches its entity constructor with a nil owner, so the constructors must tolerate that.
func TestDispenserLaunchesOwnerlessProjectiles(t *testing.T) {
	for _, test := range []struct {
		name string
		it   world.Item
		want string
	}{
		{name: "snowball", it: item.Snowball{}, want: "minecraft:snowball"},
		{name: "egg", it: item.Egg{}, want: "minecraft:egg"},
		{name: "splash potion", it: item.SplashPotion{}, want: "minecraft:splash_potion"},
		{name: "lingering potion", it: item.LingeringPotion{}, want: "minecraft:lingering_potion"},
		{name: "bottle of enchanting", it: item.BottleOfEnchanting{}, want: "minecraft:xp_bottle"},
		{name: "firework", it: item.Firework{}, want: "minecraft:fireworks_rocket"},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
			defer func() { _ = w.Close() }()

			pos := cube.Pos{0, 0, 0}
			runDispenserWorld(t, w, func(tx *world.Tx) {
				d := block.NewDispenser()
				d.Facing = cube.FaceEast
				_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(test.it, 1))
				tx.SetBlock(pos, d, nil)
				d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(17, 18)))

				for e := range tx.Entities() {
					if got := e.H().Type().EncodeEntity(); got != test.want {
						t.Fatalf("expected dispenser to launch %q, got %q", test.want, got)
					}
					return
				}
				t.Fatalf("expected dispenser to launch a %s entity", test.name)
			})
		})
	}
}

func TestDispenserLaunchesArrows(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.Arrow{}, 1))
		tx.SetBlock(pos, d, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(1, 2)))
	})

	runDispenserWorld(t, w, func(tx *world.Tx) {
		for e := range tx.Entities() {
			if got := e.H().Type().EncodeEntity(); got != "minecraft:arrow" {
				t.Fatalf("expected dispenser to launch an arrow, got %q", got)
			}
			return
		}
		t.Fatal("expected dispenser to create an arrow entity")
	})
}

func TestDispenserFillsBucketFromSource(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	front := pos.Side(cube.FaceEast)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.Bucket{}, 1))
		tx.SetBlock(pos, d, nil)
		tx.SetLiquid(front, block.Water{Depth: 8})
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(3, 4)))

		stack, _ := d.Inventory(tx, pos).Item(0)
		bucket, ok := stack.Item().(item.Bucket)
		if !ok || bucket.Empty() {
			t.Fatalf("expected source water to fill the bucket, got %v", stack)
		}
		if _, ok := tx.Liquid(front); ok {
			t.Fatal("expected filled source water to be removed")
		}
	})
}

func TestDispenserPrimesTNT(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceSouth
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(block.TNT{}, 1))
		tx.SetBlock(pos, d, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(5, 6)))
	})

	runDispenserWorld(t, w, func(tx *world.Tx) {
		for e := range tx.Entities() {
			if got := e.H().Type().EncodeEntity(); got != "minecraft:tnt" {
				t.Fatalf("expected dispenser to prime TNT, got %q", got)
			}
			return
		}
		t.Fatal("expected dispenser to create a primed TNT entity")
	})
}

func TestDispenserUsesFlintAndSteel(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	front := pos.Side(cube.FaceEast)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.FlintAndSteel{}, 1))
		tx.SetBlock(pos, d, nil)
		tx.SetBlock(front.Side(cube.FaceDown), block.Stone{}, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(7, 8)))

		if _, ok := tx.Block(front).(block.Fire); !ok {
			t.Fatalf("expected flint and steel to light fire, got %T", tx.Block(front))
		}
		stack, _ := d.Inventory(tx, pos).Item(0)
		if stack.Durability() != stack.MaxDurability()-1 {
			t.Fatalf("expected flint and steel to take one durability, got %d", stack.Durability())
		}
	})
}

func TestDispenserFillsGlassBottle(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	front := pos.Side(cube.FaceEast)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.GlassBottle{}, 1))
		tx.SetBlock(pos, d, nil)
		tx.SetLiquid(front, block.Water{Depth: 8})
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(9, 10)))

		stack, _ := d.Inventory(tx, pos).Item(0)
		if _, ok := stack.Item().(item.Potion); !ok {
			t.Fatalf("expected glass bottle to become a water potion, got %v", stack)
		}
		if _, ok := tx.Liquid(front); !ok {
			t.Fatal("expected bottling water not to consume the source")
		}
	})
}

func TestDispenserWaxesCopper(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	front := pos.Side(cube.FaceEast)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.Honeycomb{}, 1))
		tx.SetBlock(pos, d, nil)
		tx.SetBlock(front, block.Copper{}, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(11, 12)))

		if copper := tx.Block(front).(block.Copper); !copper.Waxed {
			t.Fatal("expected honeycomb to wax the copper block")
		}
	})
}

func TestDispenserAppliesBoneMeal(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	front := pos.Side(cube.FaceEast)
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.BoneMeal{}, 1))
		tx.SetBlock(pos, d, nil)
		tx.SetBlock(front, block.Carrot{}, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(13, 14)))

		if carrot := tx.Block(front).(block.Carrot); carrot.Growth == 0 {
			t.Fatal("expected bone meal to grow the crop")
		}
		stack, _ := d.Inventory(tx, pos).Item(0)
		if !stack.Empty() {
			t.Fatalf("expected successful bone meal use to consume one item, got %v", stack)
		}
	})
}

func TestDispenserRetainsBoneMealWhenUseFails(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer func() { _ = w.Close() }()

	pos := cube.Pos{0, 0, 0}
	runDispenserWorld(t, w, func(tx *world.Tx) {
		d := block.NewDispenser()
		d.Facing = cube.FaceEast
		_ = d.Inventory(tx, pos).SetItem(0, item.NewStack(item.BoneMeal{}, 1))
		tx.SetBlock(pos, d, nil)
		d.ScheduledTick(pos, tx, rand.New(rand.NewPCG(15, 16)))

		stack, _ := d.Inventory(tx, pos).Item(0)
		if stack.Count() != 1 {
			t.Fatalf("expected failed bone meal use to retain the item, got %v", stack)
		}
	})
}

func entityCount(t *testing.T, w *world.World) int {
	t.Helper()
	count := 0
	runDispenserWorld(t, w, func(tx *world.Tx) {
		for range tx.Entities() {
			count++
		}
	})
	return count
}
