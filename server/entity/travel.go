package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"sync"
	"time"
)

// TravelComputer handles the interdimensional travelling of an entity.
type TravelComputer struct {
	// Instantaneous is a function that returns true if the entity given can travel instantly.
	Instantaneous func() bool

	mu             sync.RWMutex
	start          time.Time
	awaitingTravel bool
	travelling     bool
	timedOut       bool
}

// Traveller represents a world.Entity that can travel between dimensions.
type Traveller interface {
	// Teleport teleports the entity to the position given.
	Teleport(pos mgl64.Vec3)

	world.Entity
}

// portalBlock represents a block that can be used as a portal to travel between dimensions.
type portalBlock interface {
	// Portal returns the dimension that the portal leads to.
	Portal() world.Dimension
}

// TickTravelling checks if the player is colliding with a nether portal block. If so, it teleports the player
// to the other dimension after four seconds or instantly if instantaneous is true.
func (t *TravelComputer) TickTravelling(e Traveller) {
	w := e.World()
	box := e.BBox().Translate(e.Position()).Grow(0.25)

	min, max := box.Min(), box.Max()
	minX, minY, minZ := int(math.Floor(min[0])), int(math.Floor(min[1])), int(math.Floor(min[2]))
	maxX, maxY, maxZ := int(math.Ceil(max[0])), int(math.Ceil(max[1])), int(math.Ceil(max[2]))
	travelling, target := false, world.Dimension(nil)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				pos := cube.Pos{x, y, z}
				b := w.Block(pos)
				if p, ok := b.(portalBlock); ok {
					for _, blockBox := range b.Model().BBox(pos, w) {
						if blockBox.Translate(pos.Vec3()).IntersectsWith(box) {
							travelling, target = true, p.Portal()
							break
						}
					}
				}
			}
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if !travelling {
		// No portals found. Check if we aren't travelling and if so, reset.
		if !t.travelling {
			t.timedOut, t.awaitingTravel = false, false
		}
		return
	}

	switch target {
	case world.Nether:
		timeOut, awaitingTravel, start := t.timedOut, t.awaitingTravel, t.start
		if !timeOut {
			if t.Instantaneous() || (awaitingTravel && time.Since(start) >= time.Second*4) {
				t.mu.Unlock()
				t.Travel(e, w, w.PortalDestination(world.Nether))
				t.mu.Lock()
			} else if !awaitingTravel {
				t.start, t.awaitingTravel = time.Now(), true
			}
		}
	}
}

// Travel moves the player to the given Nether or Overworld world, and translates the player's current position based
// on the source world.
func (t *TravelComputer) Travel(e Traveller, source *world.World, destination *world.World) {
	sourceDimension, targetDimension := source.Dimension(), destination.Dimension()
	pos := cube.PosFromVec3(e.Position())
	if sourceDimension == world.Overworld {
		pos = cube.Pos{pos.X() / 8, pos.Y() + sourceDimension.Range().Min(), pos.Z() / 8}
	} else if sourceDimension == world.Nether {
		pos = cube.Pos{pos.X() * 8, pos.Y() - targetDimension.Range().Min(), pos.Z() * 8}
	}

	t.mu.Lock()
	t.travelling, t.timedOut, t.awaitingTravel = true, true, false
	t.mu.Unlock()

	go func() {
		spawn := pos.Vec3Middle()
		if netherPortal, ok := portal.FindOrCreateNetherPortal(destination, pos, 128); ok {
			spawn = netherPortal.Spawn().Vec3Middle()
		}

		destination.AddEntity(e)
		e.Teleport(spawn)

		t.mu.Lock()
		t.travelling = false
		t.mu.Unlock()
	}()
}
