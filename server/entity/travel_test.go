package entity

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestPortalTravelComputerInstantaneousByDimension(t *testing.T) {
	calls := 0
	tc := &PortalTravelComputer{Instantaneous: func(source, target world.Dimension) bool {
		calls++
		return source == world.End || target == world.End
	}}

	if !tc.instantaneous(world.Overworld, world.End) {
		t.Fatal("instantaneous(Overworld, End) = false, want true")
	}
	if !tc.instantaneous(world.End, world.Overworld) {
		t.Fatal("instantaneous(End, Overworld) = false, want true")
	}
	if tc.instantaneous(world.Overworld, world.Nether) {
		t.Fatal("instantaneous(Overworld, Nether) = true, want false")
	}
	if calls != 3 {
		t.Fatalf("Instantaneous closure called %d times, want 3", calls)
	}
}

func TestPortalTravelComputerStopPortalContact(t *testing.T) {
	t.Run("keeps timer after portal contact", func(t *testing.T) {
		tc := &PortalTravelComputer{inside: true, awaitingTravel: true, start: time.Now()}
		tc.StopPortalContact()
		if !tc.awaitingTravel {
			t.Fatal("StopPortalContact() reset travel timer after portal contact")
		}
		if tc.inside {
			t.Fatal("StopPortalContact() did not clear portal contact for the next tick")
		}
	})

	t.Run("resets timer without portal contact", func(t *testing.T) {
		tc := &PortalTravelComputer{awaitingTravel: true, start: time.Now()}
		tc.StopPortalContact()
		if tc.awaitingTravel {
			t.Fatal("StopPortalContact() kept travel timer without portal contact")
		}
	})
}

func TestEntProjectileTravelsThroughPortal(t *testing.T) {
	var overworld, nether *world.World
	overworld = world.Config{PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.Nether {
			return nether
		}
		return nil
	}}.New()
	nether = world.Config{Dim: world.Nether, PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.Nether {
			return overworld
		}
		return nil
	}}.New()
	t.Cleanup(func() {
		_ = overworld.Close()
		_ = nether.Close()
	})

	spawnRecorder := &entitySpawnRecorder{}
	nether.Handle(spawnRecorder)

	sourcePos := mgl64.Vec3{80.5, 64, 80.5}
	targetPortal := cube.Pos{10, 64, 10}
	<-nether.Exec(func(tx *world.Tx) {
		buildActivePortal(tx, targetPortal)
	})

	handle := world.EntitySpawnOpts{Position: sourcePos}.New(EnderPearlType, enderPearlConf)
	<-overworld.Exec(func(tx *world.Tx) {
		e := tx.AddEntity(handle)
		(block.Portal{Axis: cube.Z}).EntityInside(cube.PosFromVec3(sourcePos), tx, e)
		if _, ok := handle.Entity(tx); !ok {
			t.Fatal("non-terminal portal contact removed entity before the source transaction finished")
		}
	})

	waitForEntityWorld(t, handle, nether)
	if entityInWorld(handle, overworld) {
		t.Fatal("entity remained in the source world after portal travel")
	}
	if !spawnRecorder.called {
		t.Fatal("destination world did not fire an entity spawn event")
	}
	if got, want := spawnRecorder.pos, targetPortal.Vec3Middle(); !got.ApproxEqual(want) {
		t.Fatalf("destination spawn event position = %v, want %v", got, want)
	}

	<-nether.Exec(func(tx *world.Tx) {
		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("entity was not added to the Nether")
		}
		if got, want := cube.PosFromVec3(e.Position()), targetPortal; got != want {
			t.Fatalf("entity position after portal travel = %v, want %v", got, want)
		}
		ent, ok := e.(*Ent)
		if !ok {
			t.Fatalf("entity after portal travel has type %T, want *Ent", e)
		}
		projectile, ok := ent.Behaviour().(*ProjectileBehaviour)
		if !ok {
			t.Fatalf("entity behaviour after portal travel has type %T, want *ProjectileBehaviour", ent.Behaviour())
		}
		if !projectile.PortalTravel() {
			t.Fatal("projectile portal travel state was not preserved")
		}
	})
}

func TestEntTravelsThroughPortalOnTick(t *testing.T) {
	var overworld, nether *world.World
	overworld = world.Config{PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.Nether {
			return nether
		}
		return nil
	}}.New()
	nether = world.Config{Dim: world.Nether, PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.Nether {
			return overworld
		}
		return nil
	}}.New()
	t.Cleanup(func() {
		_ = overworld.Close()
		_ = nether.Close()
	})

	sourcePortal, targetPortal := cube.Pos{80, 64, 80}, cube.Pos{10, 64, 10}
	<-overworld.Exec(func(tx *world.Tx) {
		buildActivePortal(tx, sourcePortal)
	})
	<-nether.Exec(func(tx *world.Tx) {
		buildActivePortal(tx, targetPortal)
	})

	handle := world.EntitySpawnOpts{Position: sourcePortal.Vec3Middle().Sub(mgl64.Vec3{1})}.New(testMovingEntType{}, testMoveConfig{delta: mgl64.Vec3{1}})
	<-overworld.Exec(func(tx *world.Tx) {
		e := tx.AddEntity(handle)
		ticker, ok := e.(world.TickerEntity)
		if !ok {
			t.Fatalf("entity has type %T, want world.TickerEntity", e)
		}
		ticker.Tick(tx, 1)
	})

	waitForEntityWorld(t, handle, nether)
	if entityInWorld(handle, overworld) {
		t.Fatal("entity remained in the source world after tick-driven portal travel")
	}
	<-nether.Exec(func(tx *world.Tx) {
		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("entity was not added to the Nether")
		}
		if got, want := cube.PosFromVec3(e.Position()), targetPortal; got != want {
			t.Fatalf("entity position after tick-driven portal travel = %v, want %v", got, want)
		}
		if got := e.(*Ent).Age(); got != 0 {
			t.Fatalf("entity age after terminal portal travel tick = %v, want 0", got)
		}
	})
}

func TestEntTravelsThroughEndPortal(t *testing.T) {
	var overworld, end *world.World
	overworld = world.Config{PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.End {
			return end
		}
		return nil
	}}.New()
	end = world.Config{Dim: world.End, PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.End {
			return overworld
		}
		return nil
	}}.New()
	t.Cleanup(func() {
		_ = overworld.Close()
		_ = end.Close()
	})

	sourcePortal := cube.Pos{50, 64, 50}
	<-overworld.Exec(func(tx *world.Tx) {
		tx.SetBlock(sourcePortal, block.EndPortal{}, nil)
	})

	handle := world.EntitySpawnOpts{Position: sourcePortal.Vec3Middle().Sub(mgl64.Vec3{1})}.New(testMovingEntType{}, testMoveConfig{delta: mgl64.Vec3{1}})
	<-overworld.Exec(func(tx *world.Tx) {
		e := tx.AddEntity(handle)
		ticker, ok := e.(world.TickerEntity)
		if !ok {
			t.Fatalf("entity has type %T, want world.TickerEntity", e)
		}
		ticker.Tick(tx, 1)
	})

	waitForEntityWorld(t, handle, end)
	if entityInWorld(handle, overworld) {
		t.Fatal("entity remained in the source world after End portal travel")
	}
	<-end.Exec(func(tx *world.Tx) {
		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("entity was not added to the End")
		}
		want := mgl64.Vec3{100.5, 49, 0.5}
		if got := e.Position(); !got.ApproxEqual(want) {
			t.Fatalf("entity position after End travel = %v, want %v", got, want)
		}
		// Spawn platform: 5x5 obsidian at y=48 around x=100, z=0.
		for dx := -2; dx <= 2; dx++ {
			for dz := -2; dz <= 2; dz++ {
				p := cube.Pos{100 + dx, 48, dz}
				if _, ok := tx.Block(p).(block.Obsidian); !ok {
					t.Fatalf("obsidian platform missing at %v: got %T", p, tx.Block(p))
				}
			}
		}
	})
}

func TestEndPortalReturnsToOverworldSpawn(t *testing.T) {
	var overworld, end *world.World
	overworld = world.Config{PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.End {
			return end
		}
		return nil
	}}.New()
	end = world.Config{Dim: world.End, PortalDestination: func(dim world.Dimension) *world.World {
		if dim == world.End {
			return overworld
		}
		return nil
	}}.New()
	t.Cleanup(func() {
		_ = overworld.Close()
		_ = end.Close()
	})

	endPortalPos := cube.Pos{200, 60, 200}
	<-end.Exec(func(tx *world.Tx) {
		tx.SetBlock(endPortalPos, block.EndPortal{}, nil)
	})

	handle := world.EntitySpawnOpts{Position: endPortalPos.Vec3Middle().Sub(mgl64.Vec3{1})}.New(testMovingEntType{}, testMoveConfig{delta: mgl64.Vec3{1}})
	<-end.Exec(func(tx *world.Tx) {
		e := tx.AddEntity(handle)
		e.(world.TickerEntity).Tick(tx, 1)
	})

	waitForEntityWorld(t, handle, overworld)
	<-overworld.Exec(func(tx *world.Tx) {
		e, ok := handle.Entity(tx)
		if !ok {
			t.Fatal("entity did not return to the Overworld")
		}
		want := tx.World().Spawn().Vec3Middle()
		if got := e.Position(); !got.ApproxEqual(want) {
			t.Fatalf("entity position after End→Overworld return = %v, want %v", got, want)
		}
	})
}

func waitForEntityWorld(t *testing.T, handle *world.EntityHandle, w *world.World) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if entityInWorld(handle, w) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for entity to change worlds")
}

func entityInWorld(handle *world.EntityHandle, w *world.World) bool {
	result := make(chan bool, 1)
	go func() {
		var ok bool
		running := handle.ExecWorld(func(tx *world.Tx, _ world.Entity) {
			ok = tx.World() == w
		})
		result <- running && ok
	}()

	select {
	case ok := <-result:
		return ok
	case <-time.After(50 * time.Millisecond):
		return false
	}
}

func buildActivePortal(tx *world.Tx, origin cube.Pos) {
	for x := range 2 {
		p := origin.Add(cube.Pos{0, 0, x})
		tx.SetBlock(p.Side(cube.FaceDown), block.Obsidian{}, nil)
		tx.SetBlock(p.Add(cube.Pos{0, 3}), block.Obsidian{}, nil)
	}
	for y := range 3 {
		p := origin.Add(cube.Pos{0, y})
		tx.SetBlock(p.Side(cube.FaceNorth), block.Obsidian{}, nil)
		tx.SetBlock(p.Add(cube.Pos{0, 0, 2}), block.Obsidian{}, nil)
		for x := range 2 {
			tx.SetBlock(p.Add(cube.Pos{0, 0, x}), block.Portal{Axis: cube.Z}, nil)
		}
	}
}

type entitySpawnRecorder struct {
	world.NopHandler

	called bool
	pos    mgl64.Vec3
}

func (r *entitySpawnRecorder) HandleEntitySpawn(_ *world.Tx, e world.Entity) {
	r.called = true
	r.pos = e.Position()
}

type testMoveConfig struct {
	delta mgl64.Vec3
}

func (c testMoveConfig) Apply(data *world.EntityData) {
	data.Data = &testMoveBehaviour{BaseBehaviour: NewBaseBehaviour(), delta: c.delta}
}

type testMoveBehaviour struct {
	BaseBehaviour

	delta mgl64.Vec3
}

func (b *testMoveBehaviour) Tick(e *Ent, _ *world.Tx) *Movement {
	e.data.Pos = e.data.Pos.Add(b.delta)
	return nil
}

type testMovingEntType struct{}

func (testMovingEntType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (testMovingEntType) EncodeEntity() string { return "minecraft:test_moving_ent" }
func (testMovingEntType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}
func (testMovingEntType) DecodeNBT(map[string]any, *world.EntityData) {}
func (testMovingEntType) EncodeNBT(*world.EntityData) map[string]any  { return nil }
