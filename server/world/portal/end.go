package portal

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// TryActivateEndPortal checks if the frame ring around a position is complete and fills the portal if it is.
func TryActivateEndPortal(tx *world.Tx, around cube.Pos) bool {
	y := around.Y()
	for x := around.X() - 2; x <= around.X()+2; x++ {
		for z := around.Z() - 2; z <= around.Z()+2; z++ {
			center := cube.Pos{x, y, z}
			if !validEndPortalRing(tx, center) {
				continue
			}
			portalBlock, ok := world.BlockByName("minecraft:end_portal", nil)
			if !ok {
				panic("could not find end portal block")
			}
			for innerX := -1; innerX <= 1; innerX++ {
				for innerZ := -1; innerZ <= 1; innerZ++ {
					tx.SetBlock(center.Add(cube.Pos{innerX, 0, innerZ}), portalBlock, nil)
				}
			}
			return true
		}
	}
	return false
}

func validEndPortalRing(tx *world.Tx, center cube.Pos) bool {
	if endPortalRingMatches(tx, center, false) {
		return true
	}
	return endPortalRingMatches(tx, center, true)
}

func endPortalRingMatches(tx *world.Tx, center cube.Pos, outward bool) bool {
	for _, spec := range []struct {
		offset cube.Pos
		facing cube.Direction
	}{
		{offset: cube.Pos{-2, 0, -1}, facing: cube.East},
		{offset: cube.Pos{-2, 0, 0}, facing: cube.East},
		{offset: cube.Pos{-2, 0, 1}, facing: cube.East},
		{offset: cube.Pos{2, 0, -1}, facing: cube.West},
		{offset: cube.Pos{2, 0, 0}, facing: cube.West},
		{offset: cube.Pos{2, 0, 1}, facing: cube.West},
		{offset: cube.Pos{-1, 0, -2}, facing: cube.South},
		{offset: cube.Pos{0, 0, -2}, facing: cube.South},
		{offset: cube.Pos{1, 0, -2}, facing: cube.South},
		{offset: cube.Pos{-1, 0, 2}, facing: cube.North},
		{offset: cube.Pos{0, 0, 2}, facing: cube.North},
		{offset: cube.Pos{1, 0, 2}, facing: cube.North},
	} {
		facing := spec.facing
		if outward {
			facing = facing.Opposite()
		}
		if !matchesEndPortalFrame(tx.Block(center.Add(spec.offset)), facing) {
			return false
		}
	}
	return true
}

func matchesEndPortalFrame(b world.Block, facing cube.Direction) bool {
	name, properties := b.EncodeBlock()
	if normalizeBlockName(name) != "end_portal_frame" {
		return false
	}
	if properties == nil {
		return false
	}
	dir, ok := properties["minecraft:cardinal_direction"].(string)
	if !ok || dir != facing.String() {
		return false
	}
	eye, ok := properties["end_portal_eye_bit"].(bool)
	return ok && eye
}
