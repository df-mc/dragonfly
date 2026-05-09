package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// TravelThroughPortal handles the entity touching a portal block.
func (e *Ent) TravelThroughPortal(tx *world.Tx, target world.Dimension) {
	if tc := e.portalTravelComputer(); tc != nil {
		if e.deferPortalTravel {
			tc.queuePortalTravel(tx, target)
			return
		}
		tc.EnterPortal(e, tx, target)
	}
}

// portalTravelComputer returns the behaviour's portal travel state, if any.
func (e *Ent) portalTravelComputer() *PortalTravelComputer {
	if b, ok := e.Behaviour().(portalTravelComputerProvider); ok {
		return b.PortalTravelComputer()
	}
	return nil
}

// stopPortalContact resets portal contact state when no portal was touched.
func (e *Ent) stopPortalContact() {
	if tc := e.portalTravelComputer(); tc != nil {
		tc.StopPortalContact()
	}
}

// pendingPortalTravel reports whether this tick queued terminal portal travel.
func (e *Ent) pendingPortalTravel() bool {
	if tc := e.portalTravelComputer(); tc != nil {
		return tc.hasPendingPortalTravel()
	}
	return false
}

// finishPendingPortalTravel starts queued terminal portal travel, if present.
func (e *Ent) finishPendingPortalTravel(tx *world.Tx) bool {
	if tc := e.portalTravelComputer(); tc != nil {
		return tc.finishPendingPortalTravel(e, tx)
	}
	return false
}

type portalBlock interface {
	Portal() world.Dimension
}

// checkPortalInsiders checks whether the entity is inside portal blocks.
// Other EntityInsider blocks are intentionally left to entity physics.
func (e *Ent) checkPortalInsiders() bool {
	box := e.H().Type().BBox(e).Translate(e.Position()).Grow(-0.0001)
	low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())

	for y := low[1]; y <= high[1]; y++ {
		for x := low[0]; x <= high[0]; x++ {
			for z := low[2]; z <= high[2]; z++ {
				blockPos := cube.Pos{x, y, z}
				if p, ok := e.tx.Block(blockPos).(portalBlock); ok {
					e.TravelThroughPortal(e.tx, p.Portal())
					if e.pendingPortalTravel() {
						return true
					}
				}
			}
		}
	}
	return false
}
