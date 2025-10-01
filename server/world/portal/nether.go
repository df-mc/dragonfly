package portal

import (
	"math"
	"math/rand"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// Nether contains information about a nether portal structure.
type Nether struct {
	w, h      int
	framed    bool
	axis      cube.Axis
	tx        *world.Tx
	spawnPos  cube.Pos
	positions []cube.Pos
}

const (
	// minimumNetherPortalWidth, maximumNetherPortalWidth controls the minimum and maximum width of a portal.
	minimumNetherPortalWidth, maximumNetherPortalWidth = 2, 21
	// minimumNetherPortalHeight, maximumNetherPortalHeight controls the minimum and maximum height of a portal.
	minimumNetherPortalHeight, maximumNetherPortalHeight = 3, 21
	// minimumArea is the minimum area of a portal.
	minimumArea = minimumNetherPortalWidth * minimumNetherPortalHeight
)

// NetherPortalFromPos returns Nether portal information from a given position in the frame.
func NetherPortalFromPos(tx *world.Tx, pos cube.Pos) (Nether, bool) {
	if tx.World().Dimension() == world.End {
		// Don't waste our time; we can't make a portal in the end.
		return Nether{}, false
	}

	axis, positions, width, height, completed, ok := multiAxisScan(pos, tx, []string{
		"minecraft:air",
		"minecraft:fire",
	})
	if !ok {
		axis, positions, width, height, completed, ok = multiAxisScan(pos, tx, []string{"minecraft:portal"})
	}
	return Nether{
		w: width, h: height,
		spawnPos:  pos,
		positions: positions,
		framed:    completed,
		axis:      axis,
		tx:        tx,
	}, ok
}

// FindOrCreateNetherPortal finds or creates a Nether portal at the given position.
func FindOrCreateNetherPortal(tx *world.Tx, pos cube.Pos, radius int) (Nether, bool) {
	n, ok := FindNetherPortal(tx, pos, radius)
	if ok {
		return n, true
	}
	return CreateNetherPortal(tx, pos)
}

// portalBlock represents a block that can be used as a portal to travel between dimensions.
type portalBlock interface {
	// Portal returns the dimension that the portal leads to.
	Portal() world.Dimension
}

// frameBlock represents a block that can be used as a frame for a Nether portal.
type frameBlock interface {
	// Frame returns true if the block is used as a frame for the given dimension.
	Frame(dimension world.Dimension) bool
}

// FindNetherPortal searches a provided radius for a Nether portal.
func FindNetherPortal(tx *world.Tx, pos cube.Pos, radius int) (Nether, bool) {
	if tx.World().Dimension() == world.End {
		// Don't waste our time - we can't make a portal in the end.
		return Nether{}, false
	}

	closestPos, closestDist, found := cube.Pos{}, math.MaxFloat64, false
	for x := pos.X() - radius; x < pos.X()+radius; x++ {
		for z := pos.Z() - radius; z < pos.Z()+radius; z++ {
			r := tx.World().Dimension().Range()
			for y := r.Max(); y >= r.Min(); y-- {
				selectedPos := cube.Pos{x, y, z}
				// Just check if it's a portal block, don't check destination dimension
				if _, ok := tx.Block(selectedPos).(portalBlock); ok {
					// Verify the portal is valid by checking if we can get portal info from it
					if portal, portalOk := NetherPortalFromPos(tx, selectedPos); portalOk && portal.Framed() {
						dist := selectedPos.Vec3().Sub(pos.Vec3()).Len()
						if dist < closestDist {
							closestDist, closestPos, found = dist, selectedPos, true
						}
					}
				}
			}
		}
	}
	if !found {
		// Don't waste our time if the search didn't work out.
		return Nether{}, false
	}
	return NetherPortalFromPos(tx, closestPos)
}

// CreateNetherPortal creates a Nether portal at the given position.
func CreateNetherPortal(tx *world.Tx, pos cube.Pos) (Nether, bool) {
	if tx.World().Dimension() == world.End {
		// You can't create a nether portal in the end.
		return Nether{}, false
	}

	resultPos, random, distance, a, r := pos, rand.Intn(4), -1.0, 0, tx.Range()
	searchValidArea := func(directions int, valid func(pos cube.Pos, riv int, coEff1, coEff2 int) bool) {
		for tempX := pos.X() - 16; tempX <= pos.X()+16; tempX++ {
			offsetX := float64(tempX-pos.X()) + 0.5
			for tempZ := pos.Z() - 16; tempZ <= pos.Z()+16; tempZ++ {
				offsetZ := float64(tempZ-pos.Z()) + 0.5
				for tempY := r.Max() - 1; tempY >= r.Min(); tempY-- {
					entryPos := cube.Pos{tempX, tempY, tempZ}
					if tx.Block(entryPos) != air() {
						continue
					}

					for tempY > r.Min() && tx.Block(entryPos.Side(cube.FaceDown)) == air() {
						tempY--
						entryPos[1]--
					}

					for riv := random; riv < random+directions; riv++ {
						coEff1 := riv % 2
						coEff2 := 1 - coEff1

						if !valid(entryPos, riv, coEff1, coEff2) {
							break
						}

						offsetY := float64(tempY-pos.Y()) + 0.5
						newDist := offsetX*offsetX + offsetY*offsetY + offsetZ*offsetZ
						if distance < 0.0 || newDist < distance {
							distance = newDist
							a = riv % directions
							resultPos = cube.Pos{tempX, tempY, tempZ}
						}
					}
				}
			}
		}
	}

	// Search for a valid area in all four directions, adding some extra space for comfort.
	searchValidArea(4, func(pos cube.Pos, riv int, coEff1, coEff2 int) bool {
		if riv%4 >= 2 {
			coEff1 = -coEff1
			coEff2 = -coEff2
		}

		for safeSpace1 := 0; safeSpace1 < 3; safeSpace1++ {
			for safeSpace2 := -1; safeSpace2 < 3; safeSpace2++ {
				for height := -1; height < 4; height++ {
					b := tx.Block(cube.Pos{
						pos.X() + safeSpace2*coEff1 + safeSpace1*coEff2,
						pos.Y() + height,
						pos.Z() + safeSpace2*coEff2 - safeSpace1*coEff1,
					})
					_, solid := b.Model().(model.Solid)
					if height < 0 && !solid || height >= 0 && b != air() {
						return false
					}
				}
			}
		}
		return true
	})

	if distance < 0.0 {
		// If we couldn't find a valid area under those specifications, we can search the two main directions instead,
		// reducing comfort but at least allowing us to have a portal in the area.
		searchValidArea(2, func(pos cube.Pos, riv int, coEff1, coEff2 int) bool {
			for safeSpace := 0; safeSpace < 3; safeSpace++ {
				for height := -1; height < 4; height++ {
					b := tx.Block(cube.Pos{
						pos.X() + safeSpace*coEff1,
						pos.Y() + height,
						pos.Z() + safeSpace*coEff2,
					})
					_, solid := b.Model().(model.Solid)
					if height < 0 && !solid || height >= 0 && b != air() {
						return false
					}
				}
			}
			return true
		})
	}

	coEff1 := a % 2
	coEff2 := 1 - coEff1
	if a%4 >= 2 {
		coEff1 = -coEff1
		coEff2 = -coEff2
	}

	axis := cube.X
	if coEff1 == 0 {
		axis = cube.Z
	}

	if distance < 0.0 {
		// If all else fails, we can simply create a floating platform in the void with the portal on it.
		resultPos[1] = int(math.Min(math.Max(float64(resultPos[1]), 70), float64(r.Max()-10)))
		for safeBeforeAfter := -1; safeBeforeAfter <= 1; safeBeforeAfter++ {
			for safeWidth := 0; safeWidth < 2; safeWidth++ {
				for height := -1; height < 3; height++ {
					entryPos := cube.Pos{
						resultPos.X() + safeWidth*coEff1 + safeBeforeAfter*coEff2,
						resultPos.Y() + height,
						resultPos.Z() + safeWidth*coEff2 - safeBeforeAfter*coEff1,
					}

					tx.SetBlock(entryPos, nil, nil)
					if height < 0 {
						tx.SetBlock(entryPos, obsidian(), nil)
					}
				}
			}
		}
	}

	// Build the portal frame and activate it.
	var positions []cube.Pos
	for width := -1; width < 3; width++ {
		for height := -1; height < 4; height++ {
			entryPos := cube.Pos{
				resultPos.X() + width*coEff1,
				resultPos.Y() + height,
				resultPos.Z() + width*coEff2,
			}

			if width == -1 || width == 2 || height == -1 || height == 3 {
				tx.SetBlock(entryPos, obsidian(), nil)
				continue
			}
			positions = append(positions, entryPos)
			tx.SetBlock(entryPos, portal(axis), nil)
		}
	}

	return Nether{
		w:         minimumNetherPortalWidth,
		h:         minimumNetherPortalHeight,
		framed:    true,
		spawnPos:  resultPos,
		positions: positions,
		axis:      axis,
		tx:        tx,
	}, true
}

// Bounds ...
func (n Nether) Bounds() (int, int) {
	return n.w, n.h
}

// Activate ...
func (n Nether) Activate() {
	for _, pos := range n.Positions() {
		n.tx.SetBlock(pos, portal(n.axis), nil)
	}
}

// Deactivate ...
func (n Nether) Deactivate() {
	for _, pos := range n.Positions() {
		n.tx.SetBlock(pos, nil, nil)
	}
}

// Framed ...
func (n Nether) Framed() bool {
	return n.framed
}

// Activated ...
func (n Nether) Activated() bool {
	for _, pos := range n.Positions() {
		if n.tx.Block(pos) != portal(n.axis) {
			return false
		}
	}
	return true
}

// Spawn ...
func (n Nether) Spawn() cube.Pos {
	return n.spawnPos
}

// Positions ...
func (n Nether) Positions() []cube.Pos {
	return n.positions
}
